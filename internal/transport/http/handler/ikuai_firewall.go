package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// notSupportedV4 reports that a feature has no v4 equivalent. Used for the
// legacy /Action/call features the v4 firmware dropped (custom_isp).
func notSupportedV4(c *app.RequestContext, feature string) {
	resp.BadRequest(c, feature+" is not available on iKuai v4 (the v3 /Action/call RPC was removed and there is no v4 equivalent)")
}

// ── Custom ISP ────────────────────────────────────────────────────────────────
// custom_isp ("自定义运营商") has no v4 REST equivalent and the v3 /Action/call
// RPC returns 404 on iKuai v4. Tracked for reimplementation on ip-objects +
// routing rules; for now these endpoints report the limitation clearly.

func ListCustomISP(ctx context.Context, c *app.RequestContext) {
	notSupportedV4(c, "custom_isp")
}

type AddCustomISPReq struct {
	Name    string   `json:"name"`
	IPGroup []string `json:"ip_group"`
	Comment string   `json:"comment"`
}

func AddCustomISP(ctx context.Context, c *app.RequestContext) {
	notSupportedV4(c, "custom_isp")
}

func DeleteCustomISP(ctx context.Context, c *app.RequestContext) {
	notSupportedV4(c, "custom_isp")
}

// ── Stream Domain → routing/domain-rules (v4) ─────────────────────────────────

func ListStreamDomain(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Routing().ListRoutingDomainRules(ctx, nil)
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
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	id, err := api.Routing().CreateRoutingDomainRules(ctx, map[string]any{
		"enabled":   "yes",
		"interface": strings.Join(req.Interface, ","),
		"src_addr":  req.SrcAddr,
		"domain":    strings.Join(req.Domains, ","),
		"comment":   req.Comment,
	})
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int64{"id": id})
}

func DeleteStreamDomain(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(strings.TrimSpace(string(c.Param("ids"))), 10, 64)
	if err != nil {
		// The route passes comma-separated ids; v4 Delete takes one id, so take
		// the first and loop is overkill for the UI's single-row delete.
		first := strings.SplitN(string(c.Param("ids")), ",", 2)[0]
		id, err = strconv.ParseInt(strings.TrimSpace(first), 10, 64)
		if err != nil {
			resp.BadRequest(c, "invalid ids")
			return
		}
	}
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	if err := api.Routing().DeleteRoutingDomainRules(ctx, id); err != nil {
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
