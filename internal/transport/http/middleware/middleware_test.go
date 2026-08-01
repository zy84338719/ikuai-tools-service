package middleware

import "testing"

func TestRouterIDOf(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		// router_id sits as the 4th segment under /api/v1/ikuai/<id>/...
		{"/api/v1/ikuai/office/system/status", "office"},
		{"/api/v1/ikuai/home/firewall/acl", "home"},
		// resource groups are NOT treated as router ids.
		{"/api/v1/ikuai/system/status", ""},
		{"/api/v1/ikuai/firewall/acl", ""},
		{"/api/v1/ikuai/vpn/pptp", ""},
		{"/api/v1/ikuai/network/wan", ""},
		// non-ikuai paths.
		{"/api/v1/routers", ""},
		{"/api/v1/routers/office", ""},
		{"/health", ""},
	}
	for _, tt := range tests {
		if got := routerIDOf(tt.path); got != tt.want {
			t.Errorf("routerIDOf(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestIsPublic(t *testing.T) {
	for _, p := range []string{"/", "/health", "/live", "/ready", "/ping", "/version", "/ui", "/metrics", "/api/v1/auth/login"} {
		if !isPublic(p) {
			t.Errorf("isPublic(%q) = false, want true", p)
		}
	}
	for _, p := range []string{"/api/v1/ikuai/office/system/status", "/api/v1/routers", "/api/v1/ikuai/office/firewall/acl"} {
		if isPublic(p) {
			t.Errorf("isPublic(%q) = true, want false", p)
		}
	}
}
