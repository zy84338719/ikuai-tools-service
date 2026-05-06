package ikuai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/service"
	"github.com/zy84338719/ikuai-tools-service/internal/conf"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
)

var globalManager *Manager

type Manager struct {
	mu     sync.RWMutex
	client *ikuaisdk.Client // stable pointer — never replaced after Init
	api    service.APIClient
	cfg    *conf.IKuaiConfig
	stopCh chan struct{}
}

// Init creates the Manager with a stable *Client and attempts initial login.
// The same *Client is reused across reconnects so that code holding the pointer
// (e.g. the metrics exporter) always sees the current session state.
func Init(cfg *conf.IKuaiConfig) error {
	if globalManager != nil {
		return nil
	}
	client := ikuaisdk.NewClient(
		cfg.BaseURL,
		cfg.Username,
		cfg.Password,
		ikuaisdk.WithTimeout(time.Duration(cfg.Timeout)*time.Second),
	)
	m := &Manager{cfg: cfg, client: client, stopCh: make(chan struct{})}
	if err := m.login(); err != nil {
		logger.Error(fmt.Sprintf("ikuai initial login failed (will retry): %v", err))
	}
	globalManager = m
	go m.sessionKeeper()
	return nil
}

func Get() *Manager {
	return globalManager
}

// login calls Login on the existing *Client and (re)creates the APIClient wrapper.
// Caller must hold m.mu for write, or call before globalManager is published.
func (m *Manager) login() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := m.client.Login(ctx); err != nil {
		m.api = nil
		return fmt.Errorf("login to ikuai: %w", err)
	}
	m.api = service.NewAPIClient(m.client)
	logger.Info("iKuai client connected")
	return nil
}

// sessionKeeper probes the router every 10 minutes and re-logs-in on failure.
// It exits when m.stopCh is closed (via Close()).
func (m *Manager) sessionKeeper() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.mu.RLock()
			api := m.api
			m.mu.RUnlock()

			var probeErr error
			if api != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				_, probeErr = api.System().GetHomepage(ctx)
				cancel()
			}
			if api == nil || isAuthError(probeErr) {
				logger.Info("iKuai session expired — re-logging in")
				m.mu.Lock()
				if err := m.login(); err != nil {
					logger.Error(fmt.Sprintf("ikuai re-login failed: %v", err))
				}
				m.mu.Unlock()
			}
		}
	}
}

// API returns a live APIClient, lazy-reconnecting if needed.
func (m *Manager) API() service.APIClient {
	m.mu.RLock()
	api := m.api
	m.mu.RUnlock()
	if api != nil {
		return api
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.api == nil {
		if err := m.login(); err != nil {
			logger.Error(fmt.Sprintf("ikuai lazy login failed: %v", err))
		}
	}
	return m.api
}

// Client returns the stable *Client pointer.
// Callers that hold this pointer (e.g. the metrics exporter) benefit from the
// exporter's own ensureLoggedIn() check on every scrape.
func (m *Manager) Client() *ikuaisdk.Client {
	return m.client // immutable after Init; no lock needed
}

// Reconnect forces a new login on the existing client object.
func (m *Manager) Reconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.login()
}

// APIWithReconnect returns the API client, retrying login once on auth errors.
func (m *Manager) APIWithReconnect(ctx context.Context, prevErr error) service.APIClient {
	if isAuthError(prevErr) {
		logger.Info("auth error detected — reconnecting iKuai session")
		m.mu.Lock()
		if err := m.login(); err != nil {
			m.mu.Unlock()
			logger.Error(fmt.Sprintf("ikuai reconnect failed: %v", err))
			return nil
		}
		m.mu.Unlock()
	}
	return m.API()
}

func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.api != nil
}

// Close stops the background session keeper and closes the underlying client.
func (m *Manager) Close() {
	close(m.stopCh)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client != nil {
		m.client.Close()
		m.api = nil
	}
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, kw := range []string{"not logged in", "unauthorized", "10000", "session expired", "please login"} {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}

var ErrNotConnected = errors.New("ikuai client not connected")
