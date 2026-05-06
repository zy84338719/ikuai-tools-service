package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-api/types"
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
	items, err := api.VPN().GetPPTPClients(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// AddPPTPClient creates a PPTP VPN client.
// @router /api/v1/ikuai/vpn/pptp [POST]
func AddPPTPClient(ctx context.Context, c *app.RequestContext) {
	var req types.PPTPClientAddRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.VPN().AddPPTPClient(ctx, &req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int{"id": id})
}

// EditPPTPClient updates a PPTP VPN client.
// @router /api/v1/ikuai/vpn/pptp [PUT]
func EditPPTPClient(ctx context.Context, c *app.RequestContext) {
	var req types.PPTPClientEditRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.VPN().EditPPTPClient(ctx, &req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// DeletePPTPClient removes a PPTP VPN client by ID.
// @router /api/v1/ikuai/vpn/pptp/:id [DELETE]
func DeletePPTPClient(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.Atoi(string(c.Param("id")))
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.VPN().DelPPTPClient(ctx, id); err != nil {
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
	items, err := api.VPN().GetL2TPClients(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// AddL2TPClient creates an L2TP VPN client.
// @router /api/v1/ikuai/vpn/l2tp [POST]
func AddL2TPClient(ctx context.Context, c *app.RequestContext) {
	var req types.L2TPClientAddRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.VPN().AddL2TPClient(ctx, &req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int{"id": id})
}

// EditL2TPClient updates an L2TP VPN client.
// @router /api/v1/ikuai/vpn/l2tp [PUT]
func EditL2TPClient(ctx context.Context, c *app.RequestContext) {
	var req types.L2TPClientEditRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.VPN().EditL2TPClient(ctx, &req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// DeleteL2TPClient removes an L2TP VPN client by ID.
// @router /api/v1/ikuai/vpn/l2tp/:id [DELETE]
func DeleteL2TPClient(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.Atoi(string(c.Param("id")))
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.VPN().DelL2TPClient(ctx, id); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}
