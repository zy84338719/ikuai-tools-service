package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
	"github.com/zy84338719/ikuai-tools-service/internal/repo/db"
	"github.com/zy84338719/ikuai-tools-service/internal/repo/db/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// writeMethods are the HTTP verbs that mutate state and therefore get audited.
var writeMethods = map[string]bool{
	"POST": true, "PUT": true, "PATCH": true, "DELETE": true,
}

// Audit records every mutating request to the audit_logs table. It runs after
// the handler completes so it can capture the response status. Failures to
// write an audit record are logged but never break the request.
func Audit() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		method := string(c.Method())
		if !writeMethods[method] {
			c.Next(ctx)
			return
		}

		path := string(c.URI().Path())
		reqID := string(c.GetHeader(headerRequestID))
		actor := actorOf(c)
		routerID := routerIDOf(path)

		c.Next(ctx)

		// Record asynchronously so it never blocks the response. The values
		// are captured by value above, safe to use after c.Next.
		status := c.Response.StatusCode()
		go recordAudit(actor, method, path, routerID, status, reqID, c.ClientIP())
	}
}

// actorOf reads the actor set by the Auth middleware, defaulting to anonymous.
func actorOf(c *app.RequestContext) string {
	if v, exists := c.Get(CtxKeyActor); exists {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return "anonymous"
}

// routerIDOf extracts a :router_id path segment if present (used once the
// multi-router routing lands). Returns "" otherwise.
func routerIDOf(path string) string {
	// Path shapes like /api/v1/ikuai/<router_id>/... — best-effort extraction.
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// /api/v1/ikuai/<router_id>/system/...
	if len(parts) >= 4 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "ikuai" {
		seg := parts[3]
		// Only treat as router id if it's not a known top-level resource group.
		if !isResourceGroup(seg) {
			return seg
		}
	}
	return ""
}

// isResourceGroup lists the resource groups that sit directly under
// /api/v1/ikuai/ (so routerIDOf does not mistake them for a router id).
var resourceGroups = map[string]bool{
	"system": true, "firewall": true, "network": true,
	"vpn": true, "sync": true,
}

func isResourceGroup(seg string) bool {
	return resourceGroups[seg]
}

func recordAudit(actor, method, path, routerID string, status int, reqID, ip string) {
	if db.DB == nil {
		return
	}
	entry := &model.AuditLog{
		Actor:    actor,
		Method:   method,
		Path:     path,
		RouterID: routerID,
		Status:   status,
		ReqID:    reqID,
		IP:       ip,
	}
	// Use a fresh short-timeout context for the background write.
	gctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.DB.WithContext(gctx).Create(entry).Error; err != nil {
		// Never fail the request over an audit write; just log.
		if err != gorm.ErrRecordNotFound {
			logger.Error("audit log write failed", zap.Error(err), zap.String("path", path))
		}
	}
}
