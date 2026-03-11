package projectkeepalive

import (
	"context"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"sync"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/util"
	sdkconfig "github.com/router-for-me/CLIProxyAPI/v6/sdk/config"
	log "github.com/sirupsen/logrus"
)

const requestUserAgent = "CLIProxyAPI-ProjectKeepAlive"

type runtimeConfig struct {
	url      string
	interval time.Duration
	proxyURL string
}

// Manager manages the background project keep-alive worker.
type Manager struct {
	mu      sync.Mutex
	baseCtx context.Context
	cancel  context.CancelFunc
}

// NewManager creates a new project keep-alive manager.
func NewManager() *Manager {
	return &Manager{}
}

// Start binds the manager to the service lifecycle context and applies the current config.
func (m *Manager) Start(ctx context.Context, cfg *config.Config) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}
	m.baseCtx = ctx
	m.restartLocked(cfg)
}

// Update reapplies configuration and restarts the worker when needed.
func (m *Manager) Update(cfg *config.Config) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.baseCtx == nil {
		m.baseCtx = context.Background()
	}
	m.restartLocked(cfg)
}

// Stop stops the active keep-alive worker.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
}

func (m *Manager) restartLocked(cfg *config.Config) {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}

	runtimeCfg, ok := buildRuntimeConfig(cfg)
	if !ok {
		return
	}

	ctx, cancel := context.WithCancel(m.baseCtx)
	m.cancel = cancel

	go m.run(ctx, runtimeCfg)
}

func buildRuntimeConfig(cfg *config.Config) (runtimeConfig, bool) {
	if cfg == nil || !cfg.ProjectKeepAlive.Enabled {
		return runtimeConfig{}, false
	}

	url := strings.TrimSpace(cfg.ProjectKeepAlive.URL)
	if url == "" || cfg.ProjectKeepAlive.IntervalSeconds <= 0 {
		return runtimeConfig{}, false
	}

	parsed, err := neturl.ParseRequestURI(url)
	scheme := ""
	host := ""
	if parsed != nil {
		scheme = strings.ToLower(parsed.Scheme)
		host = parsed.Host
	}
	if err != nil || (scheme != "http" && scheme != "https") || host == "" {
		log.Warnf("project keep-alive disabled due to invalid url %q: %v", url, err)
		return runtimeConfig{}, false
	}

	if _, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil); err != nil {
		log.Warnf("project keep-alive disabled due to invalid url %q: %v", url, err)
		return runtimeConfig{}, false
	}

	return runtimeConfig{
		url:      url,
		interval: time.Duration(cfg.ProjectKeepAlive.IntervalSeconds) * time.Second,
		proxyURL: strings.TrimSpace(cfg.ProxyURL),
	}, true
}

func (m *Manager) run(ctx context.Context, cfg runtimeConfig) {
	client := &http.Client{
		Timeout: requestTimeout(cfg.interval),
	}
	util.SetProxy(&sdkconfig.SDKConfig{ProxyURL: cfg.proxyURL}, client)

	log.Infof("project keep-alive started: %s every %s", cfg.url, cfg.interval)
	defer log.Infof("project keep-alive stopped: %s", cfg.url)

	m.fire(ctx, client, cfg.url)

	ticker := time.NewTicker(cfg.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.fire(ctx, client, cfg.url)
		}
	}
}

func (m *Manager) fire(ctx context.Context, client *http.Client, targetURL string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		log.Warnf("project keep-alive request build failed for %s: %v", targetURL, err)
		return
	}
	req.Header.Set("User-Agent", requestUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == nil {
			log.Warnf("project keep-alive request failed for %s: %v", targetURL, err)
		}
		return
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= http.StatusBadRequest {
		log.Warnf("project keep-alive request returned status %d for %s", resp.StatusCode, targetURL)
		return
	}

	log.Debugf("project keep-alive request completed for %s with status %d", targetURL, resp.StatusCode)
}

func requestTimeout(interval time.Duration) time.Duration {
	if interval <= 0 {
		return 10 * time.Second
	}
	if interval > 30*time.Second {
		return 30 * time.Second
	}
	return interval
}
