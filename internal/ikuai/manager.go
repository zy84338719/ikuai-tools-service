package ikuai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	ikuaiapi "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/service"
	"github.com/zy84338719/ikuai-tools-service/internal/conf"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
)

var globalManager *Manager

// Manager wraps the v4-only SDK client.
//
// The v4 REST API is stateless: a personal Bearer token authorizes every
// request, so there is no login / session-keepalive lifecycle (the old v3
// username/password flow has been removed by the SDK). Init validates the
// token with one probe request and constructs the typed APIClient once.
type Manager struct {
	mu     sync.RWMutex
	client *ikuaiapi.Client // stable pointer — never replaced after Init
	api    *service.APIClient
	cfg    *conf.IKuaiConfig
	stopCh chan struct{}
}

// Init creates the Manager with a stable *Client and validates the token.
func Init(cfg *conf.IKuaiConfig) error {
	if globalManager != nil {
		return nil
	}
	if cfg.Token == "" {
		// The v4-only SDK has no username/password login. Keep going so the
		// rest of the service (HTTP API, jobs) can still start; API() will
		// surface the misconfiguration on first use.
		logger.Error("ikuai.token is empty — v4 REST API requires a personal token (系统设置 → 个人令牌); ikuai client disabled until configured")
		globalManager = &Manager{cfg: cfg, stopCh: make(chan struct{})}
		return nil
	}

	client, err := ikuaiapi.NewClient(
		cfg.BaseURL,
		ikuaiapi.WithToken(cfg.Token),
		ikuaiapi.WithTimeout(time.Duration(cfg.Timeout)*time.Second),
		ikuaiapi.WithInsecureSkipVerify(cfg.Insecure),
	)
	if err != nil {
		return fmt.Errorf("create ikuai client: %w", err)
	}

	m := &Manager{cfg: cfg, client: client, api: service.NewAPIClient(client), stopCh: make(chan struct{})}
	if err := m.probe(context.Background()); err != nil {
		logger.Error(fmt.Sprintf("ikuai token probe failed (continuing, will retry): %v", err))
	} else {
		logger.Info("iKuai client connected (v4 token mode)")
	}
	globalManager = m
	go m.healthKeeper()
	return nil
}

func Get() *Manager {
	return globalManager
}

// probe sends a lightweight request to confirm the token works.
func (m *Manager) probe(ctx context.Context) error {
	if m.api == nil {
		return errors.New("ikuai client not configured (token empty?)")
	}
	pctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_, err := m.api.Monitoring().GetMonitoringSystem(pctx)
	return err
}

// healthKeeper periodically probes the router. The v4 API is stateless so
// there is no re-login; this only logs connectivity changes for observability.
func (m *Manager) healthKeeper() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			if m.api == nil {
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			if _, err := m.api.Monitoring().GetMonitoringSystem(ctx); err != nil {
				logger.Error(fmt.Sprintf("ikuai probe failed: %v", err))
			}
			cancel()
		}
	}
}

// API returns the typed APIClient, or nil if the manager was not configured
// (e.g. token was empty at startup).
func (m *Manager) API() *service.APIClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.api
}

// Client returns the stable *Client pointer. Callers that need an endpoint
// not covered by the typed service layer can use Client().Get/Post/...
// directly with an arbitrary path.
func (m *Manager) Client() *ikuaiapi.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client
}

// IsConnected reports whether the manager has a usable client.
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.api != nil
}

// Close stops the background health keeper and closes the underlying client.
func (m *Manager) Close() {
	select {
	case <-m.stopCh:
		// already closed
	default:
		close(m.stopCh)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client != nil {
		m.client.Close()
	}
}

// ActionCall posts a legacy /Action/call RPC request (func_name/action/param).
//
// A handful of iKuai features (custom_isp, stream_domain, stream_ipport,
// conn_limit, dns static) have no v4 REST endpoint. The firmware still serves
// the /Action/call RPC on v4, so callers can fall back to this method. It
// returns the inner Data payload of the response.
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

// isAuthError reports whether err looks like an auth/permission failure.
// The v4 SDK returns *ikuaiapi.APIError; we also keep a few string heuristics
// for the /Action/call escape hatch used by the custom_isp/stream_domain jobs.
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, kw := range []string{"unauthorized", "forbidden", "invalid token", "token", "10000", "session expired", "please login"} {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}

var ErrNotConnected = errors.New("ikuai client not connected")
