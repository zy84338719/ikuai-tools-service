package handler

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-api/service"
	"github.com/zy84338719/ikuai-tools-service/internal/ikuai"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// routerIDFromCtx reads the :router_id path parameter.
func routerIDFromCtx(c *app.RequestContext) string {
	return string(c.Param("router_id"))
}

// managerFor returns the Manager for the request's :router_id, writing an error
// response and returning nil when it is missing/disabled.
// Pattern: m := managerFor(c); if m == nil { return }
func managerFor(c *app.RequestContext) *ikuai.Manager {
	name := routerIDFromCtx(c)
	if name == "" {
		// Fall back to the default router for routes that don't carry a id.
		if m := ikuai.Get(); m != nil {
			return m
		}
		resp.InternalError(c, "ikuai client not initialized")
		return nil
	}
	m, err := ikuai.EnsureRouterByName(name)
	if err != nil {
		resp.NotFound(c, err.Error())
		return nil
	}
	return m
}

// ikuaiAPI returns the live API client for the request's :router_id, or writes
// an error response and returns nil.
// Pattern: api := ikuaiAPI(c); if api == nil { return }
func ikuaiAPI(c *app.RequestContext) *service.APIClient {
	m := managerFor(c)
	if m == nil {
		return nil
	}
	api := m.API()
	if api == nil {
		resp.InternalError(c, "ikuai client not connected (check router token)")
		return nil
	}
	return api
}
