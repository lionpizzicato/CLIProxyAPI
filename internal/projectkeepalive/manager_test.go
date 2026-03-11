package projectkeepalive

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lionpizzicato/CLIProxyAPI/v6/internal/config"
)

func TestManagerRequestsConfiguredURL(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mgr := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr.Start(ctx, &config.Config{
		ProjectKeepAlive: config.ProjectKeepAlive{
			Enabled:         true,
			URL:             server.URL,
			IntervalSeconds: 1,
		},
	})
	defer mgr.Stop()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if hits.Load() > 0 {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("expected keep-alive request, got %d", hits.Load())
}

func TestManagerUpdateSwitchesTargetURL(t *testing.T) {
	var firstHits atomic.Int32
	firstServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		firstHits.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer firstServer.Close()

	var secondHits atomic.Int32
	secondServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondHits.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer secondServer.Close()

	mgr := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr.Start(ctx, &config.Config{
		ProjectKeepAlive: config.ProjectKeepAlive{
			Enabled:         true,
			URL:             firstServer.URL,
			IntervalSeconds: 1,
		},
	})
	defer mgr.Stop()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if firstHits.Load() > 0 {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if firstHits.Load() == 0 {
		t.Fatalf("expected first keep-alive request before update")
	}

	mgr.Update(&config.Config{
		ProjectKeepAlive: config.ProjectKeepAlive{
			Enabled:         true,
			URL:             secondServer.URL,
			IntervalSeconds: 1,
		},
	})

	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if secondHits.Load() > 0 {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("expected keep-alive request for updated target, got %d", secondHits.Load())
}

func TestManagerDisabledConfigDoesNotRequest(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mgr := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr.Start(ctx, &config.Config{
		ProjectKeepAlive: config.ProjectKeepAlive{
			Enabled:         false,
			URL:             server.URL,
			IntervalSeconds: 1,
		},
	})
	defer mgr.Stop()

	time.Sleep(250 * time.Millisecond)

	if hits.Load() != 0 {
		t.Fatalf("expected no keep-alive requests when disabled, got %d", hits.Load())
	}
}
