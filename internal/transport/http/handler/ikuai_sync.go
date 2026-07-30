package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/conf"
	"github.com/zy84338719/ikuai-tools-service/internal/job"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// GetSyncStatus returns the last execution status of all sync jobs plus recent
// persisted history.
// @router /api/v1/ikuai/:router_id/sync/status [GET]
func GetSyncStatus(ctx context.Context, c *app.RequestContext) {
	history, _ := job.ListHistory(50)
	resp.Success(c, map[string]any{
		"current": job.GetAllStatuses(),
		"history": history,
	})
}

type TriggerCustomISPReq struct {
	Tag     string   `json:"tag"`
	Name    string   `json:"name"`    // legacy alias for Tag
	Url     []string `json:"url"`
	Comment string   `json:"comment"`
	GhProxy string   `json:"gh_proxy"`
}

// TriggerCustomISPSync manually triggers a custom ISP sync from provided URLs.
// @router /api/v1/ikuai/sync/custom-isp [POST]
func TriggerCustomISPSync(ctx context.Context, c *app.RequestContext) {
	var req TriggerCustomISPReq
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	tag := req.Tag
	if tag == "" {
		tag = req.Name
	}
	if tag == "" || len(req.Url) == 0 {
		resp.BadRequest(c, "tag and url are required")
		return
	}

	cfg := &conf.CronCustomISPConfig{
		Tag:     tag,
		Url:     req.Url,
		Comment: req.Comment,
	}
	jobsCfg := globalJobsConfig(req.GhProxy)
	routerID := routerIDFromCtx(c)

	go func() {
		_ = job.SyncCustomISP(routerID, cfg, jobsCfg)
	}()

	resp.SuccessWithMessage(c, "custom ISP sync triggered", nil)
}

type TriggerStreamDomainReq struct {
	Tag       string   `json:"tag"`
	Interface []string `json:"interface"`
	Url       []string `json:"url"`
	SrcAddr   string   `json:"src_addr"`
	Comment   string   `json:"comment"` // legacy alias for Tag
	GhProxy   string   `json:"gh_proxy"`
}

// TriggerStreamDomainSync manually triggers a stream domain sync from provided URLs.
// @router /api/v1/ikuai/sync/stream-domain [POST]
func TriggerStreamDomainSync(ctx context.Context, c *app.RequestContext) {
	var req TriggerStreamDomainReq
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	if len(req.Interface) == 0 || len(req.Url) == 0 {
		resp.BadRequest(c, "interface and url are required")
		return
	}

	cfg := &conf.CronStreamDomainConfig{
		Tag:       req.Tag,
		Interface: req.Interface,
		Url:       req.Url,
		SrcAddr:   req.SrcAddr,
		Comment:   req.Comment,
	}
	jobsCfg := globalJobsConfig(req.GhProxy)
	routerID := routerIDFromCtx(c)

	go func() {
		_ = job.SyncStreamDomain(routerID, cfg, jobsCfg)
	}()

	resp.SuccessWithMessage(c, "stream domain sync triggered", nil)
}

// globalJobsConfig returns the global jobs config merged with an optional ghProxy override.
func globalJobsConfig(ghProxyOverride string) *conf.JobsConfig {
	var base conf.JobsConfig
	if conf.GlobalConfig != nil {
		base = conf.GlobalConfig.Jobs
	}
	if ghProxyOverride != "" {
		base.GhProxy = ghProxyOverride
	}
	return &base
}
