package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// GetSystemStatus returns router monitoring stats (CPU, memory, uptime, version).
// @router /api/v1/ikuai/system/status [GET]
func GetSystemStatus(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Monitoring().GetMonitoringSystem(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

// GetInterfaces returns WAN/LAN interface status.
// @router /api/v1/ikuai/system/interfaces [GET]
func GetInterfaces(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Monitoring().GetMonitoringInterfacesStatus(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}

// GetLanDevices returns connected LAN devices.
// @router /api/v1/ikuai/system/devices [GET]
func GetLanDevices(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil {
		return
	}
	data, err := api.Monitoring().GetMonitoringClientsOnline(ctx)
	if err != nil {
		resp.InternalError(c, err.Error())
		return
	}
	resp.Success(c, data)
}
