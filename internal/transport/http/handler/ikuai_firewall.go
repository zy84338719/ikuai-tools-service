package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-api/types"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ── Custom ISP ────────────────────────────────────────────────────────────────

// ListCustomISP lists all custom ISP rules.
// @router /api/v1/ikuai/firewall/custom-isp [GET]
func ListCustomISP(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Firewall().GetCustomISP(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

type AddCustomISPReq struct {
	Name    string   `json:"name"`
	IPGroup []string `json:"ip_group"`
	Comment string   `json:"comment"`
}

// AddCustomISP creates a new custom ISP rule.
// @router /api/v1/ikuai/firewall/custom-isp [POST]
func AddCustomISP(ctx context.Context, c *app.RequestContext) {
	var req AddCustomISPReq
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	if req.Name == "" || len(req.IPGroup) == 0 {
		resp.BadRequest(c, "name and ip_group are required")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	count, err := api.Firewall().AddCustomISP(ctx, req.Name, req.IPGroup, req.Comment)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int{"count": count})
}

// DeleteCustomISP deletes custom ISP rules by comma-separated IDs.
// @router /api/v1/ikuai/firewall/custom-isp/:ids [DELETE]
func DeleteCustomISP(ctx context.Context, c *app.RequestContext) {
	ids, err := parseIDs(string(c.Param("ids")))
	if err != nil {
		resp.BadRequest(c, "invalid ids: "+err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Firewall().DelCustomISP(ctx, ids); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── Stream Domain ─────────────────────────────────────────────────────────────

// ListStreamDomain lists all stream domain routing rules.
// @router /api/v1/ikuai/firewall/stream-domain [GET]
func ListStreamDomain(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Firewall().GetStreamDomain(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

type AddStreamDomainReq struct {
	Interface []string `json:"interface"`
	Domains   []string `json:"domains"`
	SrcAddr   string   `json:"src_addr"`
	Comment   string   `json:"comment"`
}

// AddStreamDomain creates stream domain routing rules.
// @router /api/v1/ikuai/firewall/stream-domain [POST]
func AddStreamDomain(ctx context.Context, c *app.RequestContext) {
	var req AddStreamDomainReq
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	if len(req.Interface) == 0 || len(req.Domains) == 0 {
		resp.BadRequest(c, "interface and domains are required")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	count, err := api.Firewall().AddStreamDomain(ctx, req.Interface, req.Domains, req.SrcAddr, req.Comment)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int{"count": count})
}

// DeleteStreamDomain deletes stream domain rules by comma-separated IDs.
// @router /api/v1/ikuai/firewall/stream-domain/:ids [DELETE]
func DeleteStreamDomain(ctx context.Context, c *app.RequestContext) {
	ids, err := parseIDs(string(c.Param("ids")))
	if err != nil {
		resp.BadRequest(c, "invalid ids: "+err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Firewall().DelStreamDomain(ctx, ids); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── ACL ───────────────────────────────────────────────────────────────────────

// ListACL lists all ACL rules.
// @router /api/v1/ikuai/firewall/acl [GET]
func ListACL(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Firewall().GetACL(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// AddACL creates a new ACL rule.
// @router /api/v1/ikuai/firewall/acl [POST]
func AddACL(ctx context.Context, c *app.RequestContext) {
	var req types.ACLAddRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Firewall().AddACL(ctx, &req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int{"id": id})
}

// EditACL updates an existing ACL rule.
// @router /api/v1/ikuai/firewall/acl [PUT]
func EditACL(ctx context.Context, c *app.RequestContext) {
	var req types.ACLEditRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Firewall().EditACL(ctx, &req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// DeleteACL deletes an ACL rule by ID.
// @router /api/v1/ikuai/firewall/acl/:id [DELETE]
func DeleteACL(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.Atoi(string(c.Param("id")))
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Firewall().DelACL(ctx, id); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── DNAT ──────────────────────────────────────────────────────────────────────

// ListDNAT lists all DNAT (port forwarding) rules.
// @router /api/v1/ikuai/firewall/dnat [GET]
func ListDNAT(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	items, err := api.Firewall().GetDNAT(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, items)
}

// AddDNAT creates a new DNAT rule.
// @router /api/v1/ikuai/firewall/dnat [POST]
func AddDNAT(ctx context.Context, c *app.RequestContext) {
	var req types.DNATAddRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Firewall().AddDNAT(ctx, &req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int{"id": id})
}

// EditDNAT updates an existing DNAT rule.
// @router /api/v1/ikuai/firewall/dnat [PUT]
func EditDNAT(ctx context.Context, c *app.RequestContext) {
	var req types.DNATEditRequest
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Firewall().EditDNAT(ctx, &req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// DeleteDNAT deletes a DNAT rule by ID.
// @router /api/v1/ikuai/firewall/dnat/:id [DELETE]
func DeleteDNAT(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.Atoi(string(c.Param("id")))
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Firewall().DelDNAT(ctx, id); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func parseIDs(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	ids := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		ids = append(ids, n)
	}
	return ids, nil
}
