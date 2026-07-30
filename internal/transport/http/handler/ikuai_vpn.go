package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ── PPTP ──────────────────────────────────────────────────────────────────────

// ListPPTPClients lists all PPTP VPN client configurations.
// @router /api/v1/ikuai/vpn/pptp [GET]
func ListPPTPClients(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Vpn().ListVpnPptpClients(ctx, nil)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

// AddPPTPClient creates a PPTP VPN client.
// @router /api/v1/ikuai/vpn/pptp [POST]
func AddPPTPClient(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Vpn().CreateVpnPptpClients(ctx, req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int64{"id": id})
}

// EditPPTPClient updates a PPTP VPN client.
// @router /api/v1/ikuai/vpn/pptp [PUT]
func EditPPTPClient(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Vpn().UpdateVpnPptpClients(ctx, req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// DeletePPTPClient removes a PPTP VPN client by ID.
// @router /api/v1/ikuai/vpn/pptp/:id [DELETE]
func DeletePPTPClient(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Vpn().DeleteVpnPptpClients(ctx, id); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── L2TP ──────────────────────────────────────────────────────────────────────

// ListL2TPClients lists all L2TP VPN client configurations.
// @router /api/v1/ikuai/vpn/l2tp [GET]
func ListL2TPClients(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Vpn().ListVpnL2TpClients(ctx, nil)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

// AddL2TPClient creates an L2TP VPN client.
// @router /api/v1/ikuai/vpn/l2tp [POST]
func AddL2TPClient(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Vpn().CreateVpnL2TpClients(ctx, req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int64{"id": id})
}

// EditL2TPClient updates an L2TP VPN client.
// @router /api/v1/ikuai/vpn/l2tp [PUT]
func EditL2TPClient(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Vpn().UpdateVpnL2TpClients(ctx, req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// DeleteL2TPClient removes an L2TP VPN client by ID.
// @router /api/v1/ikuai/vpn/l2tp/:id [DELETE]
func DeleteL2TPClient(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Vpn().DeleteVpnL2TpClients(ctx, id); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}
