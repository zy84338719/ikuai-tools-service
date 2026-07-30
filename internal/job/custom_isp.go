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

// custom_isp has no v4 REST endpoint (the "custom ISP" concept was removed from
// the v4 surface). The router firmware still serves the legacy /Action/call
// RPC even on v4, so these jobs use Manager.Client().Post against /Action/call
// with func_name="custom_isp" — the same protocol ikuai-bypass uses.
//
// TODO(router-verify): confirm against your firmware that /Action/call is still
// reachable with a v4 token; if not, this job must be reworked onto ip-objects.

// actionCallBody/actionCallResp/actionCall live on ikuai.Manager (ActionCall);
// the legacy /Action/call RPC is shared by both jobs and the firewall handlers.

// customISPItem mirrors one row returned by custom_isp/show.
type customISPItem struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
	IPGroup string `json:"ipgroup"`
}

func SyncCustomISP(cfg *conf.CronCustomISPConfig, jobsCfg *conf.JobsConfig) error {
	tag := cfg.GetTag()
	name := "custom-isp/" + tag
	markRunning(name)
	err := syncCustomISP(cfg, jobsCfg)
	markDone(name, err)
	return err
}

func syncCustomISP(cfg *conf.CronCustomISPConfig, jobsCfg *conf.JobsConfig) error {
	start := time.Now()
	tag := cfg.GetTag()
	tagName := buildTagName(tag)

	// --- fetch ---
	var rows []string
	for _, u := range cfg.Url {
		lines, err := fetchLines(u, jobsCfg.GhProxy)
		if err != nil {
			logger.Error(fmt.Sprintf("custom-isp/%s: fetch %s failed: %v", tag, u, err))
			continue
		}
		logger.Info(fmt.Sprintf("custom-isp/%s: fetched %s rows=%d", tag, u, len(lines)))
		rows = append(rows, lines...)
	}
	if len(rows) == 0 {
		logger.Info(fmt.Sprintf("custom-isp/%s: no rows fetched, skipping", tag))
		return nil
	}
	rows = dedupe(rows)

	m := ikuai.Get()
	if m == nil || m.Client() == nil {
		return errNoClient
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// --- build existing chunk map: chunkIdx -> id ---
	data, err := m.ActionCall(ctx, "custom_isp", "show", map[string]string{
		"TYPE":  "total,data",
		"limit": "0,500",
	})
	if err != nil {
		return fmt.Errorf("custom-isp/%s: get existing: %w", tag, err)
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
			return fmt.Errorf("custom-isp/%s: decode list: %w", tag, err)
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
	logger.Info(fmt.Sprintf("custom-isp/%s: %d rows → %d chunks (max %d each)", tag, len(rows), len(chunks), chunkSize))

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
				return fmt.Errorf("custom-isp/%s: edit chunk %d (id=%d): %w", tag, chunkIdx, id, err)
			}
			delete(existing, chunkIdx)
			logger.Info(fmt.Sprintf("custom-isp/%s: edited chunk %d id=%d entries=%d", tag, chunkIdx, id, len(chunk)))
		} else {
			if _, err := m.ActionCall(ctx, "custom_isp", "add", map[string]any{
				"name":    tagName,
				"ipgroup": ipGroup,
				"comment": comment,
			}); err != nil {
				return fmt.Errorf("custom-isp/%s: add chunk %d: %w", tag, chunkIdx, err)
			}
			logger.Info(fmt.Sprintf("custom-isp/%s: added chunk %d entries=%d", tag, chunkIdx, len(chunk)))
		}
	}

	// --- delete stale chunks ---
	for idx, id := range existing {
		if _, err := m.ActionCall(ctx, "custom_isp", "del", map[string]any{
			"id": fmt.Sprintf("%d", id),
		}); err != nil {
			logger.Error(fmt.Sprintf("custom-isp/%s: delete stale chunk %d (id=%d): %v", tag, idx, id, err))
		} else {
			logger.Info(fmt.Sprintf("custom-isp/%s: deleted stale chunk %d id=%d", tag, idx, id))
		}
	}

	logger.Info(fmt.Sprintf("custom-isp/%s: done total=%d chunks=%d duration=%s",
		tag, len(rows), len(chunks), time.Since(start)))
	return nil
}
