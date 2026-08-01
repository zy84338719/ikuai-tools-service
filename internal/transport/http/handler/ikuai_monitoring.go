package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ── Monitoring: traffic / topology (read-only) ────────────────────────────────

func GetInterfacesTraffic(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Monitoring().GetMonitoringInterfacesTraffic(ctx)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func GetDownstream(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Monitoring().GetMonitoringDownstream(ctx)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func GetSwitch(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Monitoring().GetMonitoringSwitch(ctx)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func GetNetworkTopology(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Monitoring().GetMonitoringNetwork(ctx)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func GetAppTrafficSummary(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Monitoring().GetMonitoringAppTrafficSummary(ctx)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func GetClientsTrafficSummary(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c); if api == nil { return }
	data, err := api.Monitoring().GetMonitoringClientsTrafficSummary(ctx)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}
