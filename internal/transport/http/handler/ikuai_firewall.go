package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ── Custom ISP → objects/ip-objects (v4) ──────────────────────────────────────
// v3 custom_isp ("自定义运营商": name + IP list) maps onto v4 IP object groups:
// group_name ← name, group_value:[{ip,comment}] ← the IP list. Verified against
// iKuai 4.0.303 (the v3 /Action/call RPC returns 404; /ip-objects CRUD works).

// ipObject mirrors one row of /ip-objects[ip_data].
type ipObject struct {
	ID         int64                `json:"id"`
	GroupName  string               `json:"group_name"`
	GroupValue []ipObjectEntry      `json:"group_value"`
	RefCount   int                  `json:"ref_count"`
}
type ipObjectEntry struct {
	IP      string `json:"ip"`
	Comment string `json:"comment"`
}

func ListCustomISP(ctx context.Context, c *app.RequestContext) {
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
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	entries := make([]ipObjectEntry, 0, len(req.IPGroup))
	for _, ip := range req.IPGroup {
		entries = append(entries, ipObjectEntry{IP: ip})
	}
	id, err := api.Objects().CreateObjectsIpObjects(ctx, map[string]any{
		"group_name":  req.Name,
		"group_value": entries,
		"comment":     req.Comment,
	})
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, map[string]int64{"id": id})
}

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
	// v4 ip-objects delete takes a single ?id=; loop the comma list.
	for _, id := range ids {
		if err := api.Objects().DeleteObjectsIpObjects(ctx, int64(id)); err != nil {
			resp.InternalError(c, err.Error())
			return
		}
	}
	resp.Success(c, nil)
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
	// v4 domain-rules field shape (verified on iKuai 4.0.303):
	// interface is a single egress iface string; domain is {custom:[...],object:{}};
	// prio defaults to 31; comment is required on update.
	iface := req.Interface[0]
	if req.Comment == "" {
		req.Comment = "ikuai-tools"
	}
	id, err := api.Routing().CreateRoutingDomainRules(ctx, map[string]any{
		"enabled":   "yes",
		"tagname":   "ikb_" + iface,
		"interface": iface,
		"src_addr":  req.SrcAddr,
		"domain":    map[string]any{"custom": req.Domains, "object": map[string]any{}},
		"prio":      31,
		"time":      "",
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
