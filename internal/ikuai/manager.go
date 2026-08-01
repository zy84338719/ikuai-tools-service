package ikuai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	ikuaiapi "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/service"
	"github.com/zy84338719/ikuai-tools-service/internal/conf"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
)

// ErrRouterNotFound is returned when no manager is registered for a name.
var ErrRouterNotFound = errors.New("ikuai router not found")

// ErrNotConnected is returned when the named router has no usable client.
var ErrNotConnected = errors.New("ikuai client not connected")

// Manager wraps one v4-only SDK client for a single router.
type Manager struct {
	name   string
	client *ikuaiapi.Client
	api    *service.APIClient
}

// Registry holds the set of managers, one per configured router. It replaces
// the old global single-instance Manager. The zero value is not usable; build
// one with NewRegistry and populate via Add/Reload.
type Registry struct {
	mu       sync.RWMutex
	managers map[string]*Manager
	default_ string // name used when a caller doesn't specify one (legacy Get())
}

var globalRegistry = &Registry{managers: map[string]*Manager{}}

// GetRegistry returns the process-wide registry.
func GetRegistry() *Registry { return globalRegistry }

// Get returns the default manager (the first registered one), or nil. Kept for
// backward compatibility with code that does not (yet) address a specific
// router, e.g. the metrics collector.
func Get() *Manager {
	return globalRegistry.Default()
}

// Init bootstraps the registry from config. It seeds a "default" router from
// the legacy single-router ikuai config block (if a token is set) so existing
// deployments keep working, then loads any ikuai.routers entries.
func Init(cfg *conf.IKuaiConfig) error {
	if cfg.Token != "" {
		m, err := buildManager("default", cfg.BaseURL, cfg.Token, cfg.Insecure, cfg.Timeout)
		if err != nil {
			return fmt.Errorf("init default router: %w", err)
		}
		globalRegistry.Add("default", m)
		logger.Info("iKuai default router connected (v4 token mode)")
	}
	// Routers configured in the DB or via ikuai.routers are loaded later by
	// ReloadFromDB; config-file routers are handled by the caller.
	return nil
}

// buildManager constructs a Manager, validates the token with one probe, and
// starts nothing in the background (v4 is stateless).
func buildManager(name, baseURL, token string, insecure bool, timeoutSec int) (*Manager, error) {
	timeout := time.Duration(timeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	client, err := ikuaiapi.NewClient(baseURL,
		ikuaiapi.WithToken(token),
		ikuaiapi.WithTimeout(timeout),
		ikuaiapi.WithInsecureSkipVerify(insecure),
	)
	if err != nil {
		return nil, err
	}
	m := &Manager{name: name, client: client, api: service.NewAPIClient(client)}
	// Best-effort probe; don't fail construction if the router is temporarily
	// unreachable — the registry still serves other routers.
	pctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := m.api.Monitoring().GetMonitoringSystem(pctx); err != nil {
		logger.Error(fmt.Sprintf("ikuai router %q probe failed (continuing): %v", name, err))
	} else {
		logger.Info(fmt.Sprintf("ikuai router %q connected", name))
	}
	return m, nil
}

// Add registers (or replaces) a manager under the given name.
func (r *Registry) Add(name string, m *Manager) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.default_ == "" {
		r.default_ = name
	}
	r.managers[name] = m
}

// Remove unregisters a manager and closes its client.
func (r *Registry) Remove(name string) {
	r.mu.Lock()
	m := r.managers[name]
	delete(r.managers, name)
	if r.default_ == name {
		// pick a new default
		r.default_ = ""
		for n := range r.managers {
			r.default_ = n
			break
		}
	}
	r.mu.Unlock()
	if m != nil {
		// Manager.Close is nil-safe (guards a client that was never built),
		// unlike a direct m.client.Close() which would panic.
		m.Close()
	}
}

// Get returns the manager for name, or nil.
func (r *Registry) Get(name string) *Manager {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.managers[name]
}

// Default returns the default manager (first registered), or nil.
func (r *Registry) Default() *Manager {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.default_ == "" {
		return nil
	}
	return r.managers[r.default_]
}

// Names returns the registered router names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.managers))
	for n := range r.managers {
		out = append(out, n)
	}
	return out
}

// Close closes every registered client.
func (r *Registry) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range r.managers {
		m.client.Close()
	}
	r.managers = map[string]*Manager{}
	r.default_ = ""
}

// --- per-manager accessors ---

func (m *Manager) Name() string { return m.name }

// API returns the typed APIClient.
func (m *Manager) API() *service.APIClient { return m.api }

// Client returns the underlying SDK client for escape-hatch calls.
func (m *Manager) Client() *ikuaiapi.Client { return m.client }

// IsConnected reports whether the manager has a usable client.
func (m *Manager) IsConnected() bool { return m != nil && m.api != nil }

// Close releases the underlying transport.
func (m *Manager) Close() {
	if m != nil && m.client != nil {
		m.client.Close()
	}
}

// ActionCall posts a legacy /Action/call RPC request (func_name/action/param).
//
// A handful of iKuai features (custom_isp, stream_domain, stream_ipport,
// conn_limit, dns static) have no v4 REST endpoint. The firmware still serves
// the /Action/call RPC on v4, so callers can fall back to this method.
func (m *Manager) ActionCall(ctx context.Context, funcName, action string, param any) (json.RawMessage, error) {
	client := m.Client()
	if client == nil {
		return nil, ErrNotConnected
	}
	var paramRaw json.RawMessage
	if param != nil {
		b, err := json.Marshal(param)
		if err != nil {
			return nil, fmt.Errorf("marshal param: %w", err)
		}
		paramRaw = b
	}
	body := struct {
		FuncName string          `json:"func_name"`
		Action   string          `json:"action"`
		Param    json.RawMessage `json:"param"`
	}{FuncName: funcName, Action: action, Param: paramRaw}
	raw, err := client.Post(ctx, "/Action/call", body)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Result int             `json:"Result"`
		ErrMSG string          `json:"ErrMsg"`
		Data   json.RawMessage `json:"Data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode /Action/call response: %w", err)
	}
	if resp.Result != 30000 && resp.Result != 10000 {
		return nil, fmt.Errorf("/Action/call %s/%s failed: Result=%d %s", funcName, action, resp.Result, resp.ErrMSG)
	}
	return resp.Data, nil
}

// EnsureRouterByName resolves a manager by name (the :router_id path value),
// returning a clear error when it is missing/disabled. Handlers use this.
func EnsureRouterByName(name string) (*Manager, error) {
	m := GetRegistry().Get(name)
	if m == nil {
		return nil, fmt.Errorf("%w: %q", ErrRouterNotFound, name)
	}
	return m, nil
}

// RouterSpec carries the fields needed to build a Manager, decoupling the
// ikuai package from the DB model.
type RouterSpec struct {
	Name     string
	BaseURL  string
	Token    string
	Insecure bool
	Timeout  int
}

// BuildFromRouter constructs a Manager from a RouterSpec (used by handlers when
// adding/reloading routers from the DB).
func BuildFromRouter(s RouterSpec) (*Manager, error) {
	return buildManager(s.Name, s.BaseURL, s.Token, s.Insecure, s.Timeout)
}

// LoadAll loads every spec into the registry (replacing existing entries with
// the same name). Used at startup to hydrate from the DB.
func (r *Registry) LoadAll(specs []RouterSpec) {
	for _, s := range specs {
		m, err := buildManager(s.Name, s.BaseURL, s.Token, s.Insecure, s.Timeout)
		if err != nil {
			logger.Error(fmt.Sprintf("load router %q failed: %v", s.Name, err))
			continue
		}
		r.Add(s.Name, m)
	}
}
