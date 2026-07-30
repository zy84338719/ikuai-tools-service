package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	routerapp "github.com/zy84338719/ikuai-tools-service/internal/app/router"
	"github.com/zy84338719/ikuai-tools-service/internal/ikuai"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
	routermodel "github.com/zy84338719/ikuai-tools-service/internal/repo/db/model"
)

// ── Router instance management ────────────────────────────────────────────────
//
// These endpoints manage the set of routers the service talks to. Adding /
// editing / removing a router live-updates the connection registry without a
// restart. Note: these routes are NOT under /:router_id (they manage routers
// themselves), so they use their own service helper rather than ikuaiAPI.

func routerSvc() *routerapp.Service { return routerapp.NewService() }

// ListRouters lists all configured routers (tokens redacted).
// @router /api/v1/routers [GET]
func ListRouters(ctx context.Context, c *app.RequestContext) {
	rs, err := routerSvc().List(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, rs)
}

// GetRouter returns one router by name.
// @router /api/v1/routers/:name [GET]
func GetRouter(ctx context.Context, c *app.RequestContext) {
	name := string(c.Param("name"))
	r, err := routerSvc().GetByName(ctx, name)
	if err != nil {
		resp.NotFound(c, err.Error())
		return
	}
	resp.Success(c, r)
}

// CreateRouter adds a new router and connects it.
// @router /api/v1/routers [POST]
func CreateRouter(ctx context.Context, c *app.RequestContext) {
	var req routerapp.CreateRouterReq
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	r, err := routerSvc().Create(ctx, &req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	// Live-register the connection.
	registerRouter(ctx, r.Name)
	resp.Success(c, r)
}

// UpdateRouter edits a router (token/base_url/...). Reconnects on change.
// @router /api/v1/routers/:name [PUT]
func UpdateRouter(ctx context.Context, c *app.RequestContext) {
	name := string(c.Param("name"))
	var req routerapp.UpdateRouterReq
	if err := c.BindAndValidate(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	r, err := routerSvc().Update(ctx, name, &req)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	// Reconnect: drop the old client and build a fresh one.
	ikuai.GetRegistry().Remove(name)
	registerRouter(ctx, name)
	resp.Success(c, r)
}

// DeleteRouter removes a router and closes its connection.
// @router /api/v1/routers/:name [DELETE]
func DeleteRouter(ctx context.Context, c *app.RequestContext) {
	name := string(c.Param("name"))
	if err := routerSvc().Delete(ctx, name); err != nil {
		resp.NotFound(c, err.Error())
		return
	}
	ikuai.GetRegistry().Remove(name)
	resp.Success(c, nil)
}

// registerRouter (re)builds a Manager for the named router from the DB and
// adds it to the registry. Failures are logged but not fatal.
func registerRouter(ctx context.Context, name string) {
	rs, err := routerSvc().AllForManager(ctx)
	if err != nil {
		return
	}
	for i := range rs {
		if rs[i].Name != name {
			continue
		}
		m, err := ikuai.BuildFromRouter(ikuai.RouterSpec{
			Name:     rs[i].Name,
			BaseURL:  rs[i].BaseURL,
			Token:    rs[i].Token,
			Insecure: rs[i].Insecure,
			Timeout:  rs[i].Timeout,
		})
		if err != nil {
			return
		}
		ikuai.GetRegistry().Add(name, m)
		return
	}
}

// itoa is a small helper for handlers that need string ids.
func itoa(n int) string { return strconv.Itoa(n) }

// (routermodel is imported to keep the generate-friendly model alias alive)
var _ = routermodel.Router{}
