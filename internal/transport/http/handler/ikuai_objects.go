package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ── Object groups: MAC / Domain / Port / Protocol / Time ──────────────────────
// These complement the existing IP/IPv6 object groups and serve as the
// building blocks referenced by ACL / routing rules.

// MAC objects
func ListMacObjects(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Objects().ListObjectsMacObjects(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
func AddMacObjects(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	id, err := api.Objects().CreateObjectsMacObjects(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}
func EditMacObjects(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Objects().UpdateObjectsMacObjects(ctx, req); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
func DeleteMacObjects(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Objects().DeleteObjectsMacObjects(ctx, id); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}

// Domain objects
func ListDomainObjects(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Objects().ListObjectsDomainObjects(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
func AddDomainObjects(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	id, err := api.Objects().CreateObjectsDomainObjects(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}
func EditDomainObjects(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Objects().UpdateObjectsDomainObjects(ctx, req); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
func DeleteDomainObjects(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Objects().DeleteObjectsDomainObjects(ctx, id); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}

// Port objects
func ListPortObjects(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Objects().ListObjectsPortObjects(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
func AddPortObjects(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	id, err := api.Objects().CreateObjectsPortObjects(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}
func EditPortObjects(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Objects().UpdateObjectsPortObjects(ctx, req); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
func DeletePortObjects(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Objects().DeleteObjectsPortObjects(ctx, id); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}

// Protocol objects
func ListProtocolObjects(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Objects().ListObjectsProtocolObjects(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
func AddProtocolObjects(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	id, err := api.Objects().CreateObjectsProtocolObjects(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}
func EditProtocolObjects(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Objects().UpdateObjectsProtocolObjects(ctx, req); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
func DeleteProtocolObjects(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Objects().DeleteObjectsProtocolObjects(ctx, id); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}

// Time objects
func ListTimeObjects(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Objects().ListObjectsTimeObjects(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
func AddTimeObjects(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	id, err := api.Objects().CreateObjectsTimeObjects(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}
func EditTimeObjects(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Objects().UpdateObjectsTimeObjects(ctx, req); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
func DeleteTimeObjects(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c); if api == nil { return }
	if err := api.Objects().DeleteObjectsTimeObjects(ctx, id); err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, nil)
}
