package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

// CORS returns middleware that applies Access-Control-Allow-Origin based on an
// allowlist. origins is the configured allowlist:
//   - empty or contains "*" → allow any origin (default; trusted LAN only)
//   - otherwise → reflect the request Origin only if it is in the list
//
// Allow-Headers always includes Authorization and X-API-Key so API-key auth
// works cross-origin.
func CORS(origins []string) app.HandlerFunc {
	allowAny := len(origins) == 0 || contains(origins, "*")
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		allowed[strings.TrimSpace(o)] = struct{}{}
	}
	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))
		switch {
		case allowAny:
			c.Header("Access-Control-Allow-Origin", "*")
		case origin != "" && originAllowed(origin, allowed):
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With, X-API-Key")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Max-Age", "86400")

		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next(ctx)
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// originAllowed does a case-insensitive match against the allowlist. Exact
// match is preferred; this intentionally does NOT support wildcard subdomains
// to keep the policy predictable.
func originAllowed(origin string, allowed map[string]struct{}) bool {
	_, ok := allowed[origin]
	if !ok {
		low := strings.ToLower(origin)
		for o := range allowed {
			if strings.ToLower(o) == low {
				ok = true
				break
			}
		}
	}
	return ok
}
