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

// stream_domain has no v4 REST endpoint. Like custom_isp it falls back to the
// legacy /Action/call RPC (func_name="stream_domain"). See the note in
// custom_isp.go.
// TODO(router-verify): if /Action/call is unreachable on your firmware, rework
// this onto routing/domain-rules.

// streamDomainItem mirrors one row returned by stream_domain/show.
type streamDomainItem struct {
	ID      int    `json:"id"`
	Comment string `json:"comment"`
	Domain  string `json:"domain"`
}

// SyncStreamDomain runs the stream-domain sync against one router.
func SyncStreamDomain(routerID string, cfg *conf.CronStreamDomainConfig, jobsCfg *conf.JobsConfig) error {
	tag := cfg.GetTag()
	name := "stream-domain/" + tag
	markRunning(routerID, name)
	err := syncStreamDomain(routerID, cfg, jobsCfg)
	markDone(routerID, name, err)
	return err
}

func syncStreamDomain(routerID string, cfg *conf.CronStreamDomainConfig, jobsCfg *conf.JobsConfig) error {
	start := time.Now()
	tag := cfg.GetTag()

	// --- fetch ---
	var rows []string
	for _, u := range cfg.Url {
		lines, err := fetchLines(u, jobsCfg.GhProxy)
		if err != nil {
			logger.Error(fmt.Sprintf("stream-domain/%s[%s]: fetch %s failed: %v", tag, routerID, u, err))
			continue
		}
		logger.Info(fmt.Sprintf("stream-domain/%s[%s]: fetched %s rows=%d", tag, routerID, u, len(lines)))
		rows = append(rows, lines...)
	}
	if len(rows) == 0 {
		logger.Info(fmt.Sprintf("stream-domain/%s[%s]: no rows fetched, skipping", tag, routerID))
		return nil
	}
	rows = dedupe(rows)

	m := ikuai.GetRegistry().Get(routerID)
	if m == nil || m.Client() == nil {
		return errNoClient
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	iface := strings.Join(cfg.Interface, ",")
	srcAddr := cfg.SrcAddr

	// --- build existing chunk map: chunkIdx -> id ---
	data, err := m.ActionCall(ctx, "stream_domain", "show", map[string]string{
		"TYPE":  "total,data",
		"limit": "0,500",
	})
	if err != nil {
		return fmt.Errorf("stream-domain/%s[%s]: get existing: %w", tag, routerID, err)
	}
	var items []streamDomainItem
	if err := json.Unmarshal(data, &items); err != nil {
		var wrap struct {
			Data []streamDomainItem `json:"data"`
		}
		if jerr := json.Unmarshal(data, &wrap); jerr == nil && len(wrap.Data) > 0 {
			items = wrap.Data
		} else {
			return fmt.Errorf("stream-domain/%s[%s]: decode list: %w", tag, routerID, err)
		}
	}

	existing := make(map[int]int) // chunkIdx -> id
	for _, item := range items {
		if idx := parseStreamDomainComment(item.Comment, tag); idx > 0 {
			existing[idx] = item.ID
		}
	}

	// --- chunk and sync ---
	chunkSize := jobsCfg.MaxPerRecord.DomainLimit()
	chunks := splitChunks(rows, chunkSize)
	logger.Info(fmt.Sprintf("stream-domain/%s[%s]: %d rows → %d chunks (max %d each)", tag, routerID, len(rows), len(chunks), chunkSize))

	var changed, failed int
	for i, chunk := range chunks {
		chunkIdx := i + 1
		comment := buildStreamDomainComment(tag, chunkIdx)
		domains := strings.Join(chunk, ",")

		if id, ok := existing[chunkIdx]; ok {
			if _, err := m.ActionCall(ctx, "stream_domain", "edit", map[string]any{
				"id":        id,
				"enabled":   "yes",
				"interface": iface,
				"src_addr":  srcAddr,
				"domain":    domains,
				"comment":   comment,
			}); err != nil {
				failed++
				logger.Error(fmt.Sprintf("stream-domain/%s[%s]: edit chunk %d (id=%d): %v", tag, routerID, chunkIdx, id, err))
				continue
			}
			delete(existing, chunkIdx)
			changed++
			logger.Info(fmt.Sprintf("stream-domain/%s[%s]: edited chunk %d id=%d entries=%d", tag, routerID, chunkIdx, id, len(chunk)))
		} else {
			if _, err := m.ActionCall(ctx, "stream_domain", "add", map[string]any{
				"enabled":   "yes",
				"interface": iface,
				"src_addr":  srcAddr,
				"domain":    domains,
				"comment":   comment,
			}); err != nil {
				failed++
				logger.Error(fmt.Sprintf("stream-domain/%s[%s]: add chunk %d: %v", tag, routerID, chunkIdx, err))
				continue
			}
			changed++
			logger.Info(fmt.Sprintf("stream-domain/%s[%s]: added chunk %d entries=%d", tag, routerID, chunkIdx, len(chunk)))
		}
	}

	// --- delete stale chunks ---
	for idx, id := range existing {
		if _, err := m.ActionCall(ctx, "stream_domain", "del", map[string]any{
			"id": fmt.Sprintf("%d", id),
		}); err != nil {
			failed++
			logger.Error(fmt.Sprintf("stream-domain/%s[%s]: delete stale chunk %d (id=%d): %v", tag, routerID, idx, id, err))
		} else {
			changed++
			logger.Info(fmt.Sprintf("stream-domain/%s[%s]: deleted stale chunk %d id=%d", tag, routerID, idx, id))
		}
	}

	dur := time.Since(start)
	status := "success"
	if failed > 0 {
		status = "partial"
	}
	recordRun(routerID, "stream-domain", tag, status, changed, failed, dur, "")
	logger.Info(fmt.Sprintf("stream-domain/%s[%s]: done total=%d chunks=%d changed=%d failed=%d duration=%s",
		tag, routerID, len(rows), len(chunks), changed, failed, dur))
	return nil
}
