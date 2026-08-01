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

// ── Stream IP Port → routing/five-tuple-rules (v4) ────────────────────────────
// The v3 stream_ipport (/Action/call) is gone in v4; the v4 equivalent is the
// five-tuple rule (src/dst IP+port+protocol).

func ListStreamIPPort(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Routing().ListRoutingFiveTupleRules(ctx, nil)
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
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Routing().CreateRoutingFiveTupleRules(ctx, req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int64{"id": id})
}

func EditStreamIPPort(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Routing().UpdateRoutingFiveTupleRules(ctx, req); err != nil {
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
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	for _, id := range ids {
		if err := api.Routing().DeleteRoutingFiveTupleRules(ctx, int64(id)); err != nil {
			resp.InternalError(c, err.Error())
			return
		}
	}
	resp.Success(c, nil)
}

// ── Conn Limit → security/peerconn/rules (v4) ─────────────────────────────────

func ListConnLimit(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Security().ListSecurityPeerconnRules(ctx, nil)
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
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Security().CreateSecurityPeerconnRules(ctx, req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int64{"id": id})
}

func EditConnLimit(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Security().UpdateSecurityPeerconnRules(ctx, req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

func DeleteConnLimit(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Security().DeleteSecurityPeerconnRules(ctx, id); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}
