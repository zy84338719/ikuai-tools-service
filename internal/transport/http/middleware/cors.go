package middleware

import (
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/cors"
)

// CORS builds the hertz-contrib CORS middleware (the stack oekit/middleware
// recommends) from a configured origin allowlist:
//   - empty or contains "*" → allow any origin (default; trusted LAN only)
//   - otherwise → hertz-contrib reflects the request Origin only when it is in
//     AllowOrigins (exact match), keeping the policy predictable.
//
// AllowHeaders always includes Authorization and X-API-Key so API-key auth
// works cross-origin.
func CORS(origins []string) app.HandlerFunc {
	allowAny := len(origins) == 0
	for _, o := range origins {
		if o == "*" {
			allowAny = true
		}
	}
	allowOrigins := origins
	if allowAny {
		allowOrigins = []string{"*"}
	}
	return cors.New(cors.Config{
		AllowOrigins: allowOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization", "X-Requested-With", "X-API-Key"},
		ExposeHeaders: []string{"Content-Length", "Content-Type"},
		MaxAge:        24 * time.Hour,
	})
}
