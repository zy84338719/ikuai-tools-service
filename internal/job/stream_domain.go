package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zy84338719/ikuai-tools-service/internal/conf"
	"github.com/zy84338719/ikuai-tools-service/internal/ikuai"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
)

// stream_domain is implemented on v4 routing/domain-rules. The v3 /Action/call
// RPC is gone (404 on iKuai v4). Each chunk becomes a domain-rules record:
// tagname=<tagName>-<idx>, domain.custom=[domains], interface=<egress>.
// Verified against iKuai 4.0.303.

// domainRule mirrors one row of /routing/domain-rules[data].
type domainRule struct {
	ID        int64          `json:"id"`
	Tagname   string         `json:"tagname"`
	Interface string         `json:"interface"`
	Comment   string         `json:"comment"`
	Domain    domainRuleSet  `json:"domain"`
	Enabled   string         `json:"enabled"`
}
// domainRuleSet is the {custom:[],object:{}} shape used by domain/src_addr.
type domainRuleSet struct {
	Custom []string `json:"custom"`
}

// SyncStreamDomain runs the stream-domain sync against one router, mapping
// onto v4 domain-rules.
func SyncStreamDomain(routerID string, cfg *conf.CronStreamDomainConfig, jobsCfg *conf.JobsConfig) error {
	tag := cfg.GetTag()
	name := "stream-domain/" + tag
	markRunning(routerID, name)
	start := time.Now()
	changed, failed, err := syncStreamDomain(routerID, cfg, jobsCfg)
	markDone(routerID, name, err)

	status := "success"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
	} else if failed > 0 {
		status = "partial"
	}
	recordRun(routerID, "stream-domain", tag, status, changed, failed, time.Since(start), errMsg)
	return err
}

func syncStreamDomain(routerID string, cfg *conf.CronStreamDomainConfig, jobsCfg *conf.JobsConfig) (changed, failed int, err error) {
	tag := cfg.GetTag()
	tagName := buildTagName(tag)

	// --- fetch ---
	var rows []string
	for _, u := range cfg.Url {
		lines, ferr := fetchLines(u, jobsCfg.GhProxy)
		if ferr != nil {
			logger.Error(fmt.Sprintf("stream-domain/%s[%s]: fetch %s failed: %v", tag, routerID, u, ferr))
			continue
		}
		logger.Info(fmt.Sprintf("stream-domain/%s[%s]: fetched %s rows=%d", tag, routerID, u, len(lines)))
		rows = append(rows, lines...)
	}
	if len(rows) == 0 {
		logger.Info(fmt.Sprintf("stream-domain/%s[%s]: no rows fetched, skipping", tag, routerID))
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

	// v4 domain-rules interface is a single egress iface string.
	iface := ""
	if len(cfg.Interface) > 0 {
		iface = cfg.Interface[0]
	}

	// --- existing chunk map: chunkIdx -> domain-rules id ---
	raw, err := api.Routing().ListRoutingDomainRules(ctx, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("stream-domain/%s[%s]: list domain-rules: %w", tag, routerID, err)
	}
	var list struct {
		Data []domainRule `json:"data"`
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		return 0, 0, fmt.Errorf("stream-domain/%s[%s]: decode domain-rules: %w", tag, routerID, err)
	}
	existing := make(map[int]int64) // chunkIdx -> id
	for _, r := range list.Data {
		idx := chunkIndexOfName(r.Tagname, tagName)
		if idx <= 0 {
			idx = parseStreamDomainComment(r.Comment, tag)
		}
		if idx > 0 {
			existing[idx] = r.ID
		}
	}

	// --- chunk and sync ---
	chunkSize := jobsCfg.MaxPerRecord.DomainLimit()
	chunks := splitChunks(rows, chunkSize)
	logger.Info(fmt.Sprintf("stream-domain/%s[%s]: %d rows → %d chunks (max %d each)", tag, routerID, len(rows), len(chunks), chunkSize))

	for i, chunk := range chunks {
		chunkIdx := i + 1
		tagname := fmt.Sprintf("%s-%d", tagName, chunkIdx)
		comment := buildStreamDomainComment(tag, chunkIdx)
		body := map[string]any{
			"enabled":   "yes",
			"tagname":   tagname,
			"interface": iface,
			"src_addr":  cfg.SrcAddr,
			"domain":    map[string]any{"custom": chunk, "object": map[string]any{}},
			"prio":      31,
			"time":      "",
			"comment":   comment,
		}

		if id, ok := existing[chunkIdx]; ok {
			body["id"] = id
			if err := api.Routing().UpdateRoutingDomainRules(ctx, body); err != nil {
				failed++
				logger.Error(fmt.Sprintf("stream-domain/%s[%s]: edit chunk %d (id=%d): %v", tag, routerID, chunkIdx, id, err))
				continue
			}
			delete(existing, chunkIdx)
			changed++
			logger.Info(fmt.Sprintf("stream-domain/%s[%s]: edited chunk %d id=%d entries=%d", tag, routerID, chunkIdx, id, len(chunk)))
		} else {
			if _, err := api.Routing().CreateRoutingDomainRules(ctx, body); err != nil {
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
		if err := api.Routing().DeleteRoutingDomainRules(ctx, id); err != nil {
			failed++
			logger.Error(fmt.Sprintf("stream-domain/%s[%s]: delete stale chunk %d (id=%d): %v", tag, routerID, idx, id, err))
		} else {
			changed++
			logger.Info(fmt.Sprintf("stream-domain/%s[%s]: deleted stale chunk %d id=%d", tag, routerID, idx, id))
		}
	}

	logger.Info(fmt.Sprintf("stream-domain/%s[%s]: done rows=%d chunks=%d changed=%d failed=%d",
		tag, routerID, len(rows), len(chunks), changed, failed))
	return changed, failed, nil
}
