package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-api/types"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ── IP Group ──────────────────────────────────────────────────────────────────

// ListIPGroup lists all IPv4 IP group objects.
// @router /api/v1/ikuai/firewall/ip-group [GET]
func ListIPGroup(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Firewall().GetIPGroup(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// ListIPv6Group lists all IPv6 group objects.
// @router /api/v1/ikuai/firewall/ipv6-group [GET]
func ListIPv6Group(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Firewall().GetIPv6Group(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// AddIPGroup creates a new IP group.
// @router /api/v1/ikuai/firewall/ip-group [POST]
func AddIPGroup(ctx context.Context, c *app.RequestContext) {
	var req types.IPGroupAddRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	if req.GroupName == "" {
		resp.BadRequest(c, "group_name is required")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Firewall().AddIPGroup(ctx, &req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int{"id": id})
}

// EditIPGroup updates an existing IP group.
// @router /api/v1/ikuai/firewall/ip-group [PUT]
func EditIPGroup(ctx context.Context, c *app.RequestContext) {
	var req types.IPGroupEditRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Firewall().EditIPGroup(ctx, &req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// DeleteIPGroup deletes IP groups by comma-separated IDs.
// @router /api/v1/ikuai/firewall/ip-group/:ids [DELETE]
func DeleteIPGroup(ctx context.Context, c *app.RequestContext) {
	ids, err := parseIDs(string(c.Param("ids")))
	if err != nil {
		resp.BadRequest(c, "invalid ids: "+err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Firewall().DelIPGroup(ctx, ids); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── Stream IP Port ────────────────────────────────────────────────────────────

// ListStreamIPPort lists all stream IP/port routing rules.
// @router /api/v1/ikuai/firewall/stream-ipport [GET]
func ListStreamIPPort(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Firewall().GetStreamIPPort(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// AddStreamIPPort creates a new stream IP/port routing rule.
// @router /api/v1/ikuai/firewall/stream-ipport [POST]
func AddStreamIPPort(ctx context.Context, c *app.RequestContext) {
	var req types.StreamIPPortAddRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	if req.Enabled == "" {
		req.Enabled = "yes"
	}
	if req.Week == "" {
		req.Week = "1234567"
	}
	if req.Time == "" {
		req.Time = "00:00-23:59"
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Firewall().AddStreamIPPort(ctx, &req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int{"id": id})
}

// EditStreamIPPort updates an existing stream IP/port rule.
// @router /api/v1/ikuai/firewall/stream-ipport [PUT]
func EditStreamIPPort(ctx context.Context, c *app.RequestContext) {
	var req types.StreamIPPortEditRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Firewall().EditStreamIPPort(ctx, &req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// DeleteStreamIPPort deletes stream IP/port rules by comma-separated IDs.
// @router /api/v1/ikuai/firewall/stream-ipport/:ids [DELETE]
func DeleteStreamIPPort(ctx context.Context, c *app.RequestContext) {
	ids, err := parseIDs(string(c.Param("ids")))
	if err != nil {
		resp.BadRequest(c, "invalid ids: "+err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Firewall().DelStreamIPPort(ctx, ids); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── Conn Limit ────────────────────────────────────────────────────────────────

// ListConnLimit lists all connection limit rules.
// @router /api/v1/ikuai/firewall/conn-limit [GET]
func ListConnLimit(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Firewall().GetConnLimit(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// AddConnLimit creates a connection limit rule.
// @router /api/v1/ikuai/firewall/conn-limit [POST]
func AddConnLimit(ctx context.Context, c *app.RequestContext) {
	var req types.ConnLimitAddRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Firewall().AddConnLimit(ctx, &req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int{"id": id})
}

// EditConnLimit updates an existing connection limit rule.
// @router /api/v1/ikuai/firewall/conn-limit [PUT]
func EditConnLimit(ctx context.Context, c *app.RequestContext) {
	var req types.ConnLimitEditRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Firewall().EditConnLimit(ctx, &req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// DeleteConnLimit deletes a connection limit rule by ID.
// @router /api/v1/ikuai/firewall/conn-limit/:id [DELETE]
func DeleteConnLimit(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.Atoi(string(c.Param("id")))
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Firewall().DelConnLimit(ctx, id); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}
