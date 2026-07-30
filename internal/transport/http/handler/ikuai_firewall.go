package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ── Custom ISP ────────────────────────────────────────────────────────────────
// custom_isp has no v4 REST endpoint; falls back to /Action/call (see
// ikuai.Manager.ActionCall). TODO(router-verify).

func ListCustomISP(ctx context.Context, c *app.RequestContext) {
	m := managerFor(c)
	if m == nil || m.Client() == nil {
		resp.InternalError(c, "ikuai client not connected")
		return
	}
	data, err := m.ActionCall(ctx, "custom_isp", "show", map[string]string{
		"TYPE": "total,data", "limit": "0,500",
	})
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

type AddCustomISPReq struct {
	Name    string   `json:"name"`
	IPGroup []string `json:"ip_group"`
	Comment string   `json:"comment"`
}

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
	m := managerFor(c)
	if m == nil || m.Client() == nil {
		resp.InternalError(c, "ikuai client not connected")
		return
	}
	if _, err := m.ActionCall(ctx, "custom_isp", "add", map[string]any{
		"name": req.Name, "ipgroup": strings.Join(req.IPGroup, ","), "comment": req.Comment,
	}); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

func DeleteCustomISP(ctx context.Context, c *app.RequestContext) {
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
	if _, err := m.ActionCall(ctx, "custom_isp", "del", map[string]any{"id": joinInts(ids)}); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── Stream Domain ─────────────────────────────────────────────────────────────
// stream_domain has no v4 REST endpoint; falls back to /Action/call.

func ListStreamDomain(ctx context.Context, c *app.RequestContext) {
	m := managerFor(c)
	if m == nil || m.Client() == nil {
		resp.InternalError(c, "ikuai client not connected")
		return
	}
	data, err := m.ActionCall(ctx, "stream_domain", "show", map[string]string{
		"TYPE": "total,data", "limit": "0,500",
	})
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

type AddStreamDomainReq struct {
	Interface []string `json:"interface"`
	Domains   []string `json:"domains"`
	SrcAddr   string   `json:"src_addr"`
	Comment   string   `json:"comment"`
}

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
	m := managerFor(c)
	if m == nil || m.Client() == nil {
		resp.InternalError(c, "ikuai client not connected")
		return
	}
	if _, err := m.ActionCall(ctx, "stream_domain", "add", map[string]any{
		"enabled": "yes", "interface": strings.Join(req.Interface, ","),
		"src_addr": req.SrcAddr, "domain": strings.Join(req.Domains, ","), "comment": req.Comment,
	}); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

func DeleteStreamDomain(ctx context.Context, c *app.RequestContext) {
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
	idStr := joinInts(ids)
	if _, err := m.ActionCall(ctx, "stream_domain", "del", map[string]any{"id": idStr}); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── ACL (security group) ──────────────────────────────────────────────────────

func ListACL(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Security().ListSecurityAclRules(ctx, nil)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

func AddACL(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Security().CreateSecurityAclRules(ctx, req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int64{"id": id})
}

func EditACL(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Security().UpdateSecurityAclRules(ctx, req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

func DeleteACL(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Security().DeleteSecurityAclRules(ctx, id); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

// ── DNAT (network group) ──────────────────────────────────────────────────────

func ListDNAT(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Network().ListNetworkDnatRules(ctx, nil)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

func AddDNAT(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Network().CreateNetworkDnatRules(ctx, req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int64{"id": id})
}

func EditDNAT(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Network().UpdateNetworkDnatRules(ctx, req); err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, nil)
}

func DeleteDNAT(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid id")
		return
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Network().DeleteNetworkDnatRules(ctx, id); err != nil {
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

// joinInts returns a comma-separated string of ints, e.g. [1,2,3] -> "1,2,3".
func joinInts(ids []int) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.Itoa(id)
	}
	return strings.Join(parts, ",")
}
