package integration

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type testPublisher struct {
	events  chan []byte
	metrics chan []byte
}

func newTestPublisher() *testPublisher {
	return &testPublisher{
		events:  make(chan []byte, 16),
		metrics: make(chan []byte, 16),
	}
}

func (p *testPublisher) Publish(ctx context.Context, channel string, payload []byte) error {
	switch channel {
	case application.EventsChannel:
		select {
		case p.events <- payload:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	case application.MetricsChannel:
		select {
		case p.metrics <- payload:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return fmt.Errorf("unknown channel %s", channel)
}

func (p *testPublisher) Subscribe(ctx context.Context, channel string) (<-chan []byte, func(), error) {
	switch channel {
	case application.EventsChannel:
		return p.events, func() {}, nil
	case application.MetricsChannel:
		return p.metrics, func() {}, nil
	}
	return nil, nil, fmt.Errorf("unknown channel %s", channel)
}

func waitForLineContaining(t *testing.T, lines <-chan string, needle string, timeout time.Duration) string {
	t.Helper()

	deadline := time.After(timeout)
	var received []string
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				t.Fatalf("stream closed before finding %q; received %q", needle, received)
			}
			received = append(received, line)
			if strings.Contains(line, needle) {
				return line
			}
		case <-deadline:
			t.Fatalf("timed out waiting for line containing %q; received %q", needle, received)
		}
	}
}

func openEventStream(t *testing.T, r *gin.Engine, baseURL, token, path string) (<-chan string, func()) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, baseURL+path, nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		cancel()
		t.Fatalf("failed to open stream: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		cancel()
		_ = resp.Body.Close()
		t.Fatalf("expected 200 on stream open, got %d", resp.StatusCode)
	}

	lines := make(chan string, 64)
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	}()

	return lines, func() {
		cancel()
		_ = resp.Body.Close()
	}
}

func TestEventStreamDeliversIngestedEvents(t *testing.T) {
	publisher := newTestPublisher()
	r := setupRouterWith(t, publisher)

	apiKey := createUserAndGetApiKey(t, r)
	accessToken := loginAndGetAccessToken(t, r, "user@example.com")

	server := httptest.NewServer(r)
	defer server.Close()

	lines, cleanup := openEventStream(t, r, server.URL, accessToken, "/api/v1/events/stream")
	defer cleanup()

	waitForLineContaining(t, lines, "event:connected", 5*time.Second)
	waitForLineContaining(t, lines, "listening for events", 5*time.Second)

	eventID := uuid.NewString()
	ingestResp := doJSON(t, r, http.MethodPost, "/api/v1/events", []map[string]interface{}{
		{
			"id":        eventID,
			"type":      "deploy",
			"service":   "api-gateway",
			"message":   "deployed",
			"severity":  "info",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}, apiKey)
	if ingestResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on ingest, got %d: %s", ingestResp.Code, ingestResp.Body.String())
	}

	line := waitForLineContaining(t, lines, eventID, 5*time.Second)
	if !strings.HasPrefix(line, "data:") {
		t.Fatalf("expected SSE data line, got %q", line)
	}
}

func TestMetricStreamDeliversIngestedMetrics(t *testing.T) {
	publisher := newTestPublisher()
	r := setupRouterWith(t, publisher)

	apiKey := createUserAndGetApiKey(t, r)
	accessToken := loginAndGetAccessToken(t, r, "user@example.com")

	server := httptest.NewServer(r)
	defer server.Close()

	lines, cleanup := openEventStream(t, r, server.URL, accessToken, "/api/v1/metrics/stream")
	defer cleanup()

	waitForLineContaining(t, lines, "event:connected", 5*time.Second)
	waitForLineContaining(t, lines, "listening for metrics", 5*time.Second)

	metricID := uuid.NewString()
	ingestResp := doJSON(t, r, http.MethodPost, "/api/v1/metrics", []map[string]interface{}{
		{
			"id":        metricID,
			"name":      "cpu_usage",
			"service":   "api-gateway",
			"value":     "42.5",
			"unit":      "%",
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}, apiKey)
	if ingestResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on ingest, got %d: %s", ingestResp.Code, ingestResp.Body.String())
	}

	line := waitForLineContaining(t, lines, metricID, 5*time.Second)
	if !strings.HasPrefix(line, "data:") {
		t.Fatalf("expected SSE data line, got %q", line)
	}
}

func TestStreamRequiresAuth(t *testing.T) {
	r := setupRouter(t)

	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/events/stream")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", resp.StatusCode)
	}
}
