package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// ── System operations (backup / upgrade / reboot schedules / remote access) ───

// Backup
func ListBackup(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.System().ListSystemBackup(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func CreateBackup(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.System().CreateSystemBackup(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func DeleteBackup(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c)
	if api == nil { return }
	if err := api.System().DeleteSystemBackup(ctx, id); err != nil {
		resp.InternalError(c, err.Error()); return
	}
	resp.Success(c, nil)
}

func RestoreBackup(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c)
	if api == nil { return }
	_, err := api.System().RestoreSystemBackup(ctx, req)
	if err != nil {
		resp.InternalError(c, err.Error()); return
	}
	resp.Success(c, nil)
}

// Firmware upgrade
func ListUpgrade(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.System().ListSystemUpgrade(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func CheckUpgrade(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.System().CheckSystemUpgrade(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func StartUpgrade(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c)
	if api == nil { return }
	_, err := api.System().StartSystemUpgrade(ctx, req)
	if err != nil {
		resp.InternalError(c, err.Error()); return
	}
	resp.Success(c, nil)
}

func GetUpgradeStatus(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.System().GetSystemUpgradeStatus(ctx)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

// Reboot schedules (CRUD)
func ListRebootSchedules(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.System().ListSystemRebootSchedules(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func AddRebootSchedule(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c)
	if api == nil { return }
	id, err := api.System().CreateSystemRebootSchedules(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}

func EditRebootSchedule(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c)
	if api == nil { return }
	if err := api.System().UpdateSystemRebootSchedules(ctx, req); err != nil {
		resp.InternalError(c, err.Error()); return
	}
	resp.Success(c, nil)
}

func DeleteRebootSchedule(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c)
	if api == nil { return }
	if err := api.System().DeleteSystemRebootSchedules(ctx, id); err != nil {
		resp.InternalError(c, err.Error()); return
	}
	resp.Success(c, nil)
}

// Remote access (get/update)
func GetRemoteAccess(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.System().ListSystemRemoteAccess(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func UpdateRemoteAccess(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c)
	if api == nil { return }
	if err := api.System().UpdateSystemRemoteAccess(ctx, req); err != nil {
		resp.InternalError(c, err.Error()); return
	}
	resp.Success(c, nil)
}

// Disks (read-only)
func GetDisks(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.System().GetSystemDisks(ctx)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

// Web admin accounts (CRUD)
func ListWebAdminAccounts(ctx context.Context, c *app.RequestContext) {
	api := ikuaiAPI(c)
	if api == nil { return }
	data, err := api.System().ListSystemWebAdminAccounts(ctx, nil)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, data)
}

func AddWebAdminAccount(ctx context.Context, c *app.RequestContext) {
	var req map[string]any
	if err := c.BindAndValidate(&req); err != nil { resp.BadRequest(c, err.Error()); return }
	api := ikuaiAPI(c)
	if api == nil { return }
	id, err := api.System().CreateSystemWebAdminAccounts(ctx, req)
	if err != nil { resp.InternalError(c, err.Error()); return }
	resp.Success(c, map[string]int64{"id": id})
}

func DeleteWebAdminAccount(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(string(c.Param("id")), 10, 64)
	if err != nil { resp.BadRequest(c, "invalid id"); return }
	api := ikuaiAPI(c)
	if api == nil { return }
	if err := api.System().DeleteSystemWebAdminAccounts(ctx, id); err != nil {
		resp.InternalError(c, err.Error()); return
	}
	resp.Success(c, nil)
}
