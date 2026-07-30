package middleware

import (
	"context"
	"crypto/subtle"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
)

// CtxKeyActor is the request-context key under which the Auth middleware
// stores the authenticated actor for the Audit middleware to read.
const CtxKeyActor = "actor"

// publicPrefixes are request paths that never require authentication (health,
// docs, the web UI, metrics, and the auth/login endpoint itself).
var publicPrefixes = []string{
	"/health", "/live", "/ready", "/ping", "/version",
	"/ui", "/metrics", "/api/v1/auth/login",
}

// isPublic reports whether path is on the auth-free allowlist.
func isPublic(path string) bool {
	if path == "/" {
		return true
	}
	for _, p := range publicPrefixes {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}

// extractAPIKey pulls the key from either "Authorization: Bearer <key>" or
// "X-API-Key: <key>".
func extractAPIKey(c *app.RequestContext) string {
	if v := string(c.GetHeader("X-API-Key")); v != "" {
		return v
	}
	auth := string(c.GetHeader("Authorization"))
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// Auth returns middleware that enforces an API key when one is configured.
// When apiKey is empty the middleware is a no-op (auth disabled). On success
// it stores the actor ("api-key") in the request context for audit logging.
func Auth(apiKey string) app.HandlerFunc {
	enabled := apiKey != ""
	return func(ctx context.Context, c *app.RequestContext) {
		if !enabled || isPublic(string(c.URI().Path())) {
			c.Next(ctx)
			return
		}
		provided := extractAPIKey(c)
		if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
			resp.Unauthorized(c, "missing or invalid API key")
			c.Abort()
			return
		}
		c.Set(CtxKeyActor, "api-key")
		c.Next(ctx)
	}
}
