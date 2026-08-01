package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// Device logs (read-only list). Each endpoint returns the latest entries of one
// iKuai log type.

func ListLogNotice(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.Log().ListLogNotice(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func ListLogSystem(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.Log().ListLogSystem(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func ListLogAuth(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.Log().ListLogAuth(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func ListLogDhcp(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.Log().ListLogDhcp(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func ListLogPppoe(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.Log().ListLogPppoe(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func ListLogArp(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.Log().ListLogArp(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func ListLogDdns(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.Log().ListLogDdns(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func ListLogWebActivity(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.Log().ListLogWebActivity(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func ListLogWireless(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.Log().ListLogWireless(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
