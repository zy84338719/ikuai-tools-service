package job

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zy84338719/ikuai-tools-service/internal/conf"
	"github.com/zy84338719/ikuai-tools-service/internal/ikuai"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
)

// custom_isp ("自定义运营商") is implemented on v4 IP object groups
// (objects/ip-objects). The v3 /Action/call RPC is gone (404 on iKuai v4), so
// the job manages IP groups instead: each chunk becomes an ip-objects record
// whose group_value holds the CIDR list. Verified against iKuai 4.0.303.

// ipObject mirrors one row of /ip-objects[ip_data].
type ipObject struct {
	ID         int64           `json:"id"`
	GroupName  string          `json:"group_name"`
	GroupValue []ipObjectEntry `json:"group_value"`
	Comment    string          `json:"comment"`
}
type ipObjectEntry struct {
	IP      string `json:"ip"`
	Comment string `json:"comment"`
}

// ipObjectsList wraps the {ip_data:[]} envelope (ip-objects uses ip_data, not
// the usual data/results).
type ipObjectsList struct {
	IPData []ipObject `json:"ip_data"`
}

// SyncCustomISP runs the custom-isp sync against one router, mapping onto v4
// ip-objects. The scheduler calls this once per registered router.
func SyncCustomISP(routerID string, cfg *conf.CronCustomISPConfig, jobsCfg *conf.JobsConfig) error {
	tag := cfg.GetTag()
	name := "custom-isp/" + tag
	markRunning(routerID, name)
	start := time.Now()
	changed, failed, err := syncCustomISP(routerID, cfg, jobsCfg)
	markDone(routerID, name, err)

	status := "success"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
	} else if failed > 0 {
		status = "partial"
	}
	recordRun(routerID, "custom-isp", tag, status, changed, failed, time.Since(start), errMsg)
	return err
}

func syncCustomISP(routerID string, cfg *conf.CronCustomISPConfig, jobsCfg *conf.JobsConfig) (changed, failed int, err error) {
	tag := cfg.GetTag()
	tagName := buildTagName(tag)

	// --- fetch ---
	var rows []string
	for _, u := range cfg.Url {
		lines, ferr := fetchLines(u, jobsCfg.GhProxy)
		if ferr != nil {
			logger.Error(fmt.Sprintf("custom-isp/%s[%s]: fetch %s failed: %v", tag, routerID, u, ferr))
			continue
		}
		logger.Info(fmt.Sprintf("custom-isp/%s[%s]: fetched %s rows=%d", tag, routerID, u, len(lines)))
		rows = append(rows, lines...)
	}
	if len(rows) == 0 {
		logger.Info(fmt.Sprintf("custom-isp/%s[%s]: no rows fetched, skipping", tag, routerID))
		return 0, 0, nil
	}
	rows = dedupe(rows)

	m := ikuai.GetRegistry().Get(routerID)
	if m == nil || m.Client() == nil {
		return 0, 0, errNoClient
	}
	api := m.API()
	if api == nil {
		return 0, 0, errNoClient
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// --- existing chunk map: chunkIdx -> ip-objects id ---
	raw, err := api.Objects().ListObjectsIpObjects(ctx, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("custom-isp/%s[%s]: list ip-objects: %w", tag, routerID, err)
	}
	var list ipObjectsList
	if err := json.Unmarshal(raw, &list); err != nil {
		return 0, 0, fmt.Errorf("custom-isp/%s[%s]: decode ip-objects: %w", tag, routerID, err)
	}
	existing := make(map[int]int64) // chunkIdx -> id
	for _, obj := range list.IPData {
		// group_name pattern: "<tagName>-<idx>"; comment also carries ikb:<idx>.
		idx := chunkIndexOfName(obj.GroupName, tagName)
		if idx <= 0 {
			idx = parseChunkComment(obj.Comment)
		}
		if idx > 0 {
			existing[idx] = obj.ID
		}
	}

	// --- chunk and sync ---
	chunkSize := jobsCfg.MaxPerRecord.ISPLimit()
	chunks := splitChunks(rows, chunkSize)
	logger.Info(fmt.Sprintf("custom-isp/%s[%s]: %d rows → %d chunks (max %d each)", tag, routerID, len(rows), len(chunks), chunkSize))

	for i, chunk := range chunks {
		chunkIdx := i + 1
		groupName := fmt.Sprintf("%s-%d", tagName, chunkIdx)
		comment := buildChunkComment(chunkIdx)
		entries := make([]ipObjectEntry, 0, len(chunk))
		for _, ip := range chunk {
			entries = append(entries, ipObjectEntry{IP: ip})
		}

		if id, ok := existing[chunkIdx]; ok {
			if err := api.Objects().UpdateObjectsIpObjects(ctx, map[string]any{
				"id":          id,
				"group_name":  groupName,
				"group_value": entries,
				"comment":     comment,
			}); err != nil {
				failed++
				logger.Error(fmt.Sprintf("custom-isp/%s[%s]: edit chunk %d (id=%d): %v", tag, routerID, chunkIdx, id, err))
				continue
			}
			delete(existing, chunkIdx)
			changed++
			logger.Info(fmt.Sprintf("custom-isp/%s[%s]: edited chunk %d id=%d entries=%d", tag, routerID, chunkIdx, id, len(chunk)))
		} else {
			if _, err := api.Objects().CreateObjectsIpObjects(ctx, map[string]any{
				"group_name":  groupName,
				"group_value": entries,
				"comment":     comment,
			}); err != nil {
				failed++
				logger.Error(fmt.Sprintf("custom-isp/%s[%s]: add chunk %d: %v", tag, routerID, chunkIdx, err))
				continue
			}
			changed++
			logger.Info(fmt.Sprintf("custom-isp/%s[%s]: added chunk %d entries=%d", tag, routerID, chunkIdx, len(chunk)))
		}
	}

	// --- delete stale chunks ---
	for idx, id := range existing {
		if err := api.Objects().DeleteObjectsIpObjects(ctx, id); err != nil {
			failed++
			logger.Error(fmt.Sprintf("custom-isp/%s[%s]: delete stale chunk %d (id=%d): %v", tag, routerID, idx, id, err))
		} else {
			changed++
			logger.Info(fmt.Sprintf("custom-isp/%s[%s]: deleted stale chunk %d id=%d", tag, routerID, idx, id))
		}
	}

	logger.Info(fmt.Sprintf("custom-isp/%s[%s]: done rows=%d chunks=%d changed=%d failed=%d",
		tag, routerID, len(rows), len(chunks), changed, failed))
	return changed, failed, nil
}

// chunkIndexOfName extracts the trailing "-<idx>" from a group_name built as
// "<tagName>-<idx>". Returns 0 if the name doesn't match.
func chunkIndexOfName(groupName, tagName string) int {
	prefix := tagName + "-"
	if !strings.HasPrefix(groupName, prefix) {
		return 0
	}
	n := 0
	for _, ch := range groupName[len(prefix):] {
		if ch < '0' || ch > '9' {
			return 0
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		return 0
	}
	return n
}
