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

// custom_isp has NO v4 equivalent. Verified against iKuai 4.0.303: the v3
// /Action/call RPC returns 404, and no /custom-isp REST endpoint exists. The
// "自定义运营商" concept was removed in v4. This job therefore aborts with a
// clear error so job_runs records the limitation; a future reimplementation
// would build on ip-objects + routing rules, but the data model differs
// enough that it is tracked separately.

// customISPItem mirrors one row returned by custom_isp/show.
type customISPItem struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
	IPGroup string `json:"ipgroup"`
}

// SyncCustomISP runs the custom-isp sync against one router. The scheduler
// calls this once per registered router. Every execution (success, partial, or
// failed) is persisted to job_runs.
//
// NOTE: custom_isp has no v4 equivalent (verified on iKuai 4.0.303). This job
// always fails fast with errV4ActionCallGone so the limitation is visible in
// job_runs rather than silently doing nothing.
func SyncCustomISP(routerID string, cfg *conf.CronCustomISPConfig, jobsCfg *conf.JobsConfig) error {
	tag := cfg.GetTag()
	name := "custom-isp/" + tag
	markRunning(routerID, name)
	start := time.Now()
	err := errV4ActionCallGone
	markDone(routerID, name, err)
	recordRun(routerID, "custom-isp", tag, "failed", 0, 0, time.Since(start), err.Error())
	logger.Error(fmt.Sprintf("custom-isp/%s[%s]: %v", tag, routerID, err))
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// --- build existing chunk map: chunkIdx -> id ---
	data, err := m.ActionCall(ctx, "custom_isp", "show", map[string]string{
		"TYPE":  "total,data",
		"limit": "0,500",
	})
	if err != nil {
		return 0, 0, fmt.Errorf("custom-isp/%s: get existing: %w", tag, err)
	}
	var items []customISPItem
	if err := json.Unmarshal(data, &items); err != nil {
		// some firmwares nest the list under a "data" key
		var wrap struct {
			Data []customISPItem `json:"data"`
		}
		if jerr := json.Unmarshal(data, &wrap); jerr == nil && len(wrap.Data) > 0 {
			items = wrap.Data
		} else {
			return 0, 0, fmt.Errorf("custom-isp/%s: decode list: %w", tag, err)
		}
	}

	existing := make(map[int]int) // chunkIdx -> id
	for _, item := range items {
		if item.Name != tagName {
			continue
		}
		if idx := parseChunkComment(item.Comment); idx > 0 {
			existing[idx] = item.ID
		}
	}

	// --- chunk and sync ---
	chunkSize := jobsCfg.MaxPerRecord.ISPLimit()
	chunks := splitChunks(rows, chunkSize)
	logger.Info(fmt.Sprintf("custom-isp/%s[%s]: %d rows → %d chunks (max %d each)", tag, routerID, len(rows), len(chunks), chunkSize))

	for i, chunk := range chunks {
		chunkIdx := i + 1
		comment := buildChunkComment(chunkIdx)
		ipGroup := strings.Join(chunk, ",")

		if id, ok := existing[chunkIdx]; ok {
			if _, err := m.ActionCall(ctx, "custom_isp", "edit", map[string]any{
				"id":      id,
				"name":    tagName,
				"ipgroup": ipGroup,
				"comment": comment,
			}); err != nil {
				failed++
				logger.Error(fmt.Sprintf("custom-isp/%s[%s]: edit chunk %d (id=%d): %v", tag, routerID, chunkIdx, id, err))
				continue
			}
			delete(existing, chunkIdx)
			changed++
			logger.Info(fmt.Sprintf("custom-isp/%s[%s]: edited chunk %d id=%d entries=%d", tag, routerID, chunkIdx, id, len(chunk)))
		} else {
			if _, err := m.ActionCall(ctx, "custom_isp", "add", map[string]any{
				"name":    tagName,
				"ipgroup": ipGroup,
				"comment": comment,
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
		if _, err := m.ActionCall(ctx, "custom_isp", "del", map[string]any{
			"id": fmt.Sprintf("%d", id),
		}); err != nil {
			failed++
			logger.Error(fmt.Sprintf("custom-isp/%s[%s]: delete stale chunk %d (id=%d): %v", tag, routerID, idx, id, err))
		} else {
			changed++
			logger.Info(fmt.Sprintf("custom-isp/%s[%s]: deleted stale chunk %d id=%d", tag, routerID, idx, id))
		}
	}

	logger.Info(fmt.Sprintf("custom-isp/%s[%s]: done total=%d chunks=%d changed=%d failed=%d",
		tag, routerID, len(rows), len(chunks), changed, failed))
	return changed, failed, nil
}
