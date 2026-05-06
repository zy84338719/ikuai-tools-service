package job

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zy84338719/ikuai-tools-service/internal/conf"
	"github.com/zy84338719/ikuai-tools-service/internal/ikuai"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
)

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
	if m == nil {
		return errNoClient
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	fw := m.API().Firewall()

	// --- build existing chunk map: chunkIdx -> id ---
	items, err := fw.GetCustomISP(ctx)
	if err != nil {
		return fmt.Errorf("custom-isp/%s: get existing: %w", tag, err)
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

		if id, ok := existing[chunkIdx]; ok {
			ipGroup := strings.Join(chunk, ",")
			if err := fw.EditCustomISP(ctx, id, tagName, ipGroup, comment); err != nil {
				return fmt.Errorf("custom-isp/%s: edit chunk %d (id=%d): %w", tag, chunkIdx, id, err)
			}
			delete(existing, chunkIdx)
			logger.Info(fmt.Sprintf("custom-isp/%s: edited chunk %d id=%d entries=%d", tag, chunkIdx, id, len(chunk)))
		} else {
			// AddCustomISP handles exactly one SDK call when chunk size <= ISPLimit.
			if _, err := fw.AddCustomISP(ctx, tagName, chunk, comment); err != nil {
				return fmt.Errorf("custom-isp/%s: add chunk %d: %w", tag, chunkIdx, err)
			}
			logger.Info(fmt.Sprintf("custom-isp/%s: added chunk %d entries=%d", tag, chunkIdx, len(chunk)))
		}
	}

	// --- delete stale chunks ---
	for idx, id := range existing {
		if err := fw.DelCustomISP(ctx, []int{id}); err != nil {
			logger.Error(fmt.Sprintf("custom-isp/%s: delete stale chunk %d (id=%d): %v", tag, idx, id, err))
		} else {
			logger.Info(fmt.Sprintf("custom-isp/%s: deleted stale chunk %d id=%d", tag, idx, id))
		}
	}

	logger.Info(fmt.Sprintf("custom-isp/%s: done total=%d chunks=%d duration=%s",
		tag, len(rows), len(chunks), time.Since(start)))
	return nil
}

