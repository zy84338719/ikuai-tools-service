package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ── WAN / LAN (interfaces group) ──────────────────────────────────────────────

func ListWan(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Interfaces().GetInterfacesWanConfig(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

func ListLan(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Interfaces().GetInterfacesLanConfig(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

// ── DHCP (network group) ──────────────────────────────────────────────────────

func ListDHCPLeases(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Network().GetNetworkDhcpClients(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

func ListDHCPStatic(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Network().ListNetworkDhcpStatic(ctx, nil)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

func AddDHCPStatic(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Network().CreateNetworkDhcpStatic(ctx, req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int64{"id": id})
}

func EditDHCPStatic(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Network().UpdateNetworkDhcpStatic(ctx, req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// DeleteDHCPStatic removes a DHCP static binding.
// v4's dhcp-static endpoint has no DELETE verb, so we use PATCH to disable it.
func DeleteDHCPStatic(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Network().PatchNetworkDhcpStatic(ctx, map[string]any{"id": id, "enabled": "no"}); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── DNS Static → network/dns/proxy/rules (v4) ─────────────────────────────────
// v3 dns_static (/Action/call) is gone in v4; DNS static mappings use the
// dns-proxy-rules endpoint (domain → dns_addr).

func ListDNSStatic(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Network().ListNetworkDnsProxyRules(ctx, nil)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

func AddDNSStatic(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Network().CreateNetworkDnsProxyRules(ctx, req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int64{"id": id})
}

func EditDNSStatic(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Network().UpdateNetworkDnsProxyRules(ctx, req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

func DeleteDNSStatic(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Network().DeleteNetworkDnsProxyRules(ctx, id); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── Route Static (routing group) ──────────────────────────────────────────────

func ListRouteStatic(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Routing().ListRoutingStaticRoutes(ctx, nil)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

func AddRouteStatic(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Routing().CreateRoutingStaticRoutes(ctx, req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int64{"id": id})
}

func EditRouteStatic(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Routing().UpdateRoutingStaticRoutes(ctx, req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

func DeleteRouteStatic(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Routing().DeleteRoutingStaticRoutes(ctx, id); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}
