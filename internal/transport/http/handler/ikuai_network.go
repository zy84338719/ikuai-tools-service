package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-api/types"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ── WAN / LAN ─────────────────────────────────────────────────────────────────

// ListWan returns all WAN interface configurations.
// @router /api/v1/ikuai/network/wan [GET]
func ListWan(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Network().GetWan(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// ListLan returns all LAN interface configurations.
// @router /api/v1/ikuai/network/lan [GET]
func ListLan(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Network().GetLan(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// ── DHCP ──────────────────────────────────────────────────────────────────────

// ListDHCPLeases lists all active DHCP leases.
// @router /api/v1/ikuai/network/dhcp/leases [GET]
func ListDHCPLeases(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Network().GetLeases(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// ListDHCPStatic lists all DHCP static IP bindings.
// @router /api/v1/ikuai/network/dhcp/static [GET]
func ListDHCPStatic(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Network().GetStaticBindings(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// AddDHCPStatic creates a DHCP static binding.
// @router /api/v1/ikuai/network/dhcp/static [POST]
func AddDHCPStatic(ctx context.Context, c *app.RequestContext) {
	var req types.DHCPStaticAddRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Network().AddStaticBinding(ctx, &req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int{"id": id})
}

// EditDHCPStatic updates a DHCP static binding.
// @router /api/v1/ikuai/network/dhcp/static [PUT]
func EditDHCPStatic(ctx context.Context, c *app.RequestContext) {
	var req types.DHCPStaticEditRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Network().EditStaticBinding(ctx, &req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// DeleteDHCPStatic removes a DHCP static binding by ID.
// @router /api/v1/ikuai/network/dhcp/static/:id [DELETE]
func DeleteDHCPStatic(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.Atoi(string(c.Param("id")))
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Network().DelStaticBinding(ctx, id); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── DNS Static ────────────────────────────────────────────────────────────────

// ListDNSStatic lists all DNS static records.
// @router /api/v1/ikuai/network/dns/static [GET]
func ListDNSStatic(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Network().GetDNSStatic(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// AddDNSStatic creates a DNS static record.
// @router /api/v1/ikuai/network/dns/static [POST]
func AddDNSStatic(ctx context.Context, c *app.RequestContext) {
	var req types.DNSStaticAddRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Network().AddDNSStatic(ctx, &req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int{"id": id})
}

// EditDNSStatic updates a DNS static record.
// @router /api/v1/ikuai/network/dns/static [PUT]
func EditDNSStatic(ctx context.Context, c *app.RequestContext) {
	var req types.DNSStaticEditRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Network().EditDNSStatic(ctx, &req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// DeleteDNSStatic removes a DNS static record by ID.
// @router /api/v1/ikuai/network/dns/static/:id [DELETE]
func DeleteDNSStatic(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.Atoi(string(c.Param("id")))
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Network().DelDNSStatic(ctx, id); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── Route Static ──────────────────────────────────────────────────────────────

// ListRouteStatic lists all static routes.
// @router /api/v1/ikuai/network/route/static [GET]
func ListRouteStatic(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Network().GetRouteStatic(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// AddRouteStatic creates a static route.
// @router /api/v1/ikuai/network/route/static [POST]
func AddRouteStatic(ctx context.Context, c *app.RequestContext) {
	var req types.RouteStaticAddRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Network().AddRouteStatic(ctx, &req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int{"id": id})
}

// EditRouteStatic updates a static route.
// @router /api/v1/ikuai/network/route/static [PUT]
func EditRouteStatic(ctx context.Context, c *app.RequestContext) {
	var req types.RouteStaticEditRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Network().EditRouteStatic(ctx, &req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// DeleteRouteStatic removes a static route by ID.
// @router /api/v1/ikuai/network/route/static/:id [DELETE]
func DeleteRouteStatic(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.Atoi(string(c.Param("id")))
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Network().DelRouteStatic(ctx, id); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}
