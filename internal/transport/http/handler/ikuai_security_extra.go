package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ── Security extras: MAC filter rules ─────────────────────────────────────────

func ListMacRules(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Security().ListSecurityMacRules(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
func AddMacRules(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	id, err := api.Security().CreateSecurityMacRules(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}
func EditMacRules(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Security().UpdateSecurityMacRules(ctx, req); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
func DeleteMacRules(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Security().DeleteSecurityMacRules(ctx, id); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}

// ── Routing extras: load-balance / app-protocol rules ─────────────────────────

// Load balance (multi-WAN)
func ListLoadBalance(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Routing().ListRoutingLoadBalanceRules(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
func AddLoadBalance(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	id, err := api.Routing().CreateRoutingLoadBalanceRules(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}
func EditLoadBalance(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Routing().UpdateRoutingLoadBalanceRules(ctx, req); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
func DeleteLoadBalance(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Routing().DeleteRoutingLoadBalanceRules(ctx, id); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}

// App-protocol routing
func ListAppProtocols(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Routing().ListRoutingAppProtocols(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
func AddAppProtocols(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	id, err := api.Routing().CreateRoutingAppProtocols(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}
func EditAppProtocols(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Routing().UpdateRoutingAppProtocols(ctx, req); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
func DeleteAppProtocols(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Routing().DeleteRoutingAppProtocols(ctx, id); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
