package handler

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-api/service"
	"github.com/zy84338719/ikuai-tools-service/internal/ikuai"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ikuaiAPI returns the live API client or writes an error response and returns nil.
// Pattern: api := ikuaiAPI(c); if api == nil { return }
func ikuaiAPI(c *app.RequestContext) *service.APIClient {
	m := ikuai.Get()
	if m == nil {
		resp.InternalError(c, "ikuai client not initialized")
		return nil
	}
	api := m.API()
	if api == nil {
		resp.InternalError(c, "ikuai client not connected (check ikuai.token)")
		return nil
	}
	return api
}
