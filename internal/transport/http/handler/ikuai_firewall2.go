package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ── IP Group (objects group, IPv4) ────────────────────────────────────────────

func ListIPGroup(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Objects().ListObjectsIpObjects(ctx, nil)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

func ListIPv6Group(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Objects().ListObjectsIpv6Objects(ctx, nil)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

func AddIPGroup(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	if name, _ := req["group_name"].(string); name == "" {
		resp.BadRequest(c, "group_name is required")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Objects().CreateObjectsIpObjects(ctx, req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int64{"id": id})
}

func EditIPGroup(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Objects().UpdateObjectsIpObjects(ctx, req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

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
	// v4 Delete takes a single id; loop over the comma-separated list.
	for _, id := range ids {
		if err := api.Objects().DeleteObjectsIpObjects(ctx, int64(id)); err != nil {
			resp.InternalError(c, err.Error())
			return
		}
	}
	resp.Success(c, nil)
}

// ── Stream IP Port ────────────────────────────────────────────────────────────
// stream_ipport has no v4 REST endpoint; falls back to /Action/call.

func ListStreamIPPort(ctx context.Context, c *app.RequestContext) {
	m := managerFor(c)
	if m == nil || m.Client() == nil {
		resp.InternalError(c, "ikuai client not connected")
		return
	}
	data, err := m.ActionCall(ctx, "stream_ipport", "show", map[string]string{
		"TYPE": "total,data", "limit": "0,500",
	})
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

func AddStreamIPPort(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	if req["enabled"] == nil {
		req["enabled"] = "yes"
	}
	m := managerFor(c)
	if m == nil || m.Client() == nil {
		resp.InternalError(c, "ikuai client not connected")
		return
	}
	if _, err := m.ActionCall(ctx, "stream_ipport", "add", req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

func EditStreamIPPort(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	m := managerFor(c)
	if m == nil || m.Client() == nil {
		resp.InternalError(c, "ikuai client not connected")
		return
	}
	if _, err := m.ActionCall(ctx, "stream_ipport", "edit", req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

func DeleteStreamIPPort(ctx context.Context, c *app.RequestContext) {
	ids, err := parseIDs(string(c.Param("ids")))
	if err != nil {
		resp.BadRequest(c, "invalid ids: "+err.Error())
		return
	}
	m := managerFor(c)
	if m == nil || m.Client() == nil {
		resp.InternalError(c, "ikuai client not connected")
		return
	}
	if _, err := m.ActionCall(ctx, "stream_ipport", "del", map[string]any{"id": joinInts(ids)}); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── Conn Limit ────────────────────────────────────────────────────────────────
// conn_limit has no v4 REST endpoint; falls back to /Action/call.
// TODO(router-verify): v4 equivalent may be security/peerconn/rules.

func ListConnLimit(ctx context.Context, c *app.RequestContext) {
	m := managerFor(c)
	if m == nil || m.Client() == nil {
		resp.InternalError(c, "ikuai client not connected")
		return
	}
	data, err := m.ActionCall(ctx, "conn_limit", "show", map[string]string{
		"TYPE": "total,data", "limit": "0,500",
	})
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

func AddConnLimit(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	m := managerFor(c)
	if m == nil || m.Client() == nil {
		resp.InternalError(c, "ikuai client not connected")
		return
	}
	if _, err := m.ActionCall(ctx, "conn_limit", "add", req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

func EditConnLimit(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	m := managerFor(c)
	if m == nil || m.Client() == nil {
		resp.InternalError(c, "ikuai client not connected")
		return
	}
	if _, err := m.ActionCall(ctx, "conn_limit", "edit", req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

func DeleteConnLimit(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.Atoi(string(c.Param("id")))
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	m := managerFor(c)
	if m == nil || m.Client() == nil {
		resp.InternalError(c, "ikuai client not connected")
		return
	}
	if _, err := m.ActionCall(ctx, "conn_limit", "del", map[string]any{"id": strconv.Itoa(id)}); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}
