package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ── Network extras: DMZ / NAT / QoS(IP,MAC) / VLAN ────────────────────────────
// All follow the same CRUD pattern. VLAN has no DELETE (PATCH only).

// DMZ rules
func ListDMZ(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Network().ListNetworkDmzRules(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
func AddDMZ(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	id, err := api.Network().CreateNetworkDmzRules(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}
func EditDMZ(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Network().UpdateNetworkDmzRules(ctx, req); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
func DeleteDMZ(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Network().DeleteNetworkDmzRules(ctx, id); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}

// NAT rules
func ListNAT(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Network().ListNetworkNatRules(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
func AddNAT(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	id, err := api.Network().CreateNetworkNatRules(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}
func EditNAT(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Network().UpdateNetworkNatRules(ctx, req); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
func DeleteNAT(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Network().DeleteNetworkNatRules(ctx, id); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}

// QoS by IP
func ListQosIP(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Network().ListNetworkQosIp(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
func AddQosIP(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	id, err := api.Network().CreateNetworkQosIp(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}
func EditQosIP(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Network().UpdateNetworkQosIp(ctx, req); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
func DeleteQosIP(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Network().DeleteNetworkQosIp(ctx, id); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}

// QoS by MAC
func ListQosMac(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Network().ListNetworkQosMac(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
func AddQosMac(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	id, err := api.Network().CreateNetworkQosMac(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}
func EditQosMac(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Network().UpdateNetworkQosMac(ctx, req); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
func DeleteQosMac(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Network().DeleteNetworkQosMac(ctx, id); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}

// VLAN (list/add/edit; v4 has no DELETE, PATCH used to disable)
func ListVLAN(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Network().ListNetworkVlan(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
func AddVLAN(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	id, err := api.Network().CreateNetworkVlan(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}
func EditVLAN(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Network().UpdateNetworkVlan(ctx, req); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
