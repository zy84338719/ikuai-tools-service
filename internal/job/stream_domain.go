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

func SyncStreamDomain(cfg *conf.CronStreamDomainConfig, jobsCfg *conf.JobsConfig) error {
	tag := cfg.GetTag()
	name := "stream-domain/" + tag
	markRunning(name)
	err := syncStreamDomain(cfg, jobsCfg)
	markDone(name, err)
	return err
}

func syncStreamDomain(cfg *conf.CronStreamDomainConfig, jobsCfg *conf.JobsConfig) error {
	start := time.Now()
	tag := cfg.GetTag()

	// --- fetch ---
	var rows []string
	for _, u := range cfg.Url {
		lines, err := fetchLines(u, jobsCfg.GhProxy)
		if err != nil {
			logger.Error(fmt.Sprintf("stream-domain/%s: fetch %s failed: %v", tag, u, err))
			continue
		}
		logger.Info(fmt.Sprintf("stream-domain/%s: fetched %s rows=%d", tag, u, len(lines)))
		rows = append(rows, lines...)
	}
	if len(rows) == 0 {
		logger.Info(fmt.Sprintf("stream-domain/%s: no rows fetched, skipping", tag))
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

	iface := strings.Join(cfg.Interface, ",")
	srcAddr := cfg.SrcAddr

	// --- build existing chunk map: chunkIdx -> id ---
	items, err := fw.GetStreamDomain(ctx)
	if err != nil {
		return fmt.Errorf("stream-domain/%s: get existing: %w", tag, err)
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
	logger.Info(fmt.Sprintf("stream-domain/%s: %d rows → %d chunks (max %d each)", tag, len(rows), len(chunks), chunkSize))

	for i, chunk := range chunks {
		chunkIdx := i + 1
		comment := buildStreamDomainComment(tag, chunkIdx)
		domains := strings.Join(chunk, ",")

		if id, ok := existing[chunkIdx]; ok {
			if err := fw.EditStreamDomain(ctx, id, iface, srcAddr, domains, comment); err != nil {
				return fmt.Errorf("stream-domain/%s: edit chunk %d (id=%d): %w", tag, chunkIdx, id, err)
			}
			delete(existing, chunkIdx)
			logger.Info(fmt.Sprintf("stream-domain/%s: edited chunk %d id=%d entries=%d", tag, chunkIdx, id, len(chunk)))
		} else {
			if _, err := fw.AddStreamDomain(ctx, cfg.Interface, chunk, srcAddr, comment); err != nil {
				return fmt.Errorf("stream-domain/%s: add chunk %d: %w", tag, chunkIdx, err)
			}
			logger.Info(fmt.Sprintf("stream-domain/%s: added chunk %d entries=%d", tag, chunkIdx, len(chunk)))
		}
	}

	// --- delete stale chunks ---
	for idx, id := range existing {
		if err := fw.DelStreamDomain(ctx, []int{id}); err != nil {
			logger.Error(fmt.Sprintf("stream-domain/%s: delete stale chunk %d (id=%d): %v", tag, idx, id, err))
		} else {
			logger.Info(fmt.Sprintf("stream-domain/%s: deleted stale chunk %d id=%d", tag, idx, id))
		}
	}

	logger.Info(fmt.Sprintf("stream-domain/%s: done total=%d chunks=%d duration=%s",
		tag, len(rows), len(chunks), time.Since(start)))
	return nil
}
