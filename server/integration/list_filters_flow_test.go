package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEventListFiltersAndPagination(t *testing.T) {
	r := setupRouter(t)
	apiKey := createUserAndGetApiKey(t, r)
	token := loginAndGetAccessToken(t, r, "user@example.com")

	ev1 := "00000000-0000-0000-0000-000000000001"
	ev2 := "00000000-0000-0000-0000-000000000002"
	ev3 := "00000000-0000-0000-0000-000000000003"
	ev4 := "00000000-0000-0000-0000-000000000004"

	events := []map[string]string{
		{"id": ev1, "type": "deploy", "service": "api-gateway", "message": "a", "severity": "critical", "timestamp": "2026-01-01T10:00:00Z"},
		{"id": ev2, "type": "deploy", "service": "auth-service", "message": "b", "severity": "warning", "timestamp": "2026-01-01T10:30:00Z"},
		{"id": ev3, "type": "rollback", "service": "api-gateway", "message": "c", "severity": "info", "timestamp": "2026-01-01T11:00:00Z"},
		{"id": ev4, "type": "deploy", "service": "api-gateway", "message": "d", "severity": "critical", "timestamp": "2026-01-01T11:30:00Z"},
	}

	resp := doJSON(t, r, http.MethodPost, "/api/v1/events", events, apiKey)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on ingest, got %d: %s", resp.Code, resp.Body.String())
	}

	listAndAssertIds(t, r, token, "/api/v1/events?severity=critical", []string{ev4, ev1})
	listAndAssertIds(t, r, token, "/api/v1/events?type=deploy", []string{ev4, ev2, ev1})
	listAndAssertIds(t, r, token, "/api/v1/events?service=api-gateway", []string{ev4, ev3, ev1})
	listAndAssertIds(t, r, token, "/api/v1/events?start=2026-01-01T10:30:00Z&end=2026-01-01T11:00:00Z", []string{ev3, ev2})
	listAndAssertIds(t, r, token, "/api/v1/events?limit=2", []string{ev4, ev3})
	listAndAssertIds(t, r, token, "/api/v1/events?limit=2&offset=2", []string{ev2, ev1})
	listAndAssertIds(t, r, token, "/api/v1/events?offset=10", []string{})

	badStart := doJSON(t, r, http.MethodGet, "/api/v1/events?start=not-a-time", nil, token)
	if badStart.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on invalid start, got %d", badStart.Code)
	}
}

func TestMetricListFiltersAndPagination(t *testing.T) {
	r := setupRouter(t)
	apiKey := createUserAndGetApiKey(t, r)
	token := loginAndGetAccessToken(t, r, "user@example.com")

	m1 := "00000000-0000-0000-0000-000000000101"
	m2 := "00000000-0000-0000-0000-000000000102"
	m3 := "00000000-0000-0000-0000-000000000103"

	metrics := []map[string]string{
		{"id": m1, "name": "cpu_usage", "service": "api-gateway", "value": "10", "unit": "%", "status": "ok", "timestamp": "2026-01-01T10:00:00Z"},
		{"id": m2, "name": "cpu_usage", "service": "auth-service", "value": "95", "unit": "%", "status": "crit", "timestamp": "2026-01-01T10:30:00Z"},
		{"id": m3, "name": "mem_usage", "service": "api-gateway", "value": "50", "unit": "%", "status": "warn", "timestamp": "2026-01-01T11:00:00Z"},
	}

	resp := doJSON(t, r, http.MethodPost, "/api/v1/metrics", metrics, apiKey)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on ingest, got %d: %s", resp.Code, resp.Body.String())
	}

	listAndAssertIds(t, r, token, "/api/v1/metrics?status=crit", []string{m2})
	listAndAssertIds(t, r, token, "/api/v1/metrics?name=cpu_usage", []string{m2, m1})
	listAndAssertIds(t, r, token, "/api/v1/metrics?service=api-gateway", []string{m3, m1})
	listAndAssertIds(t, r, token, "/api/v1/metrics?start=2026-01-01T10:30:00Z", []string{m3, m2})
	listAndAssertIds(t, r, token, "/api/v1/metrics?limit=1&offset=1", []string{m2})

	badEnd := doJSON(t, r, http.MethodGet, "/api/v1/metrics?end=not-a-time", nil, token)
	if badEnd.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on invalid end, got %d", badEnd.Code)
	}
}

func listAndAssertIds(t *testing.T, r *gin.Engine, token string, path string, want []string) {
	t.Helper()

	resp := doJSON(t, r, http.MethodGet, path, nil, token)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 on %s, got %d: %s", path, resp.Code, resp.Body.String())
	}

	var items []struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &items); err != nil {
		t.Fatalf("unexpected error decoding list body: %v", err)
	}

	if len(items) != len(want) {
		t.Fatalf("expected %d items on %s, got %d: %s", len(want), path, len(items), resp.Body.String())
	}

	for i, item := range items {
		if item.Id != want[i] {
			t.Errorf("on %s expected item[%d]=%s, got %s", path, i, want[i], fmt.Sprint(item.Id))
		}
	}
}
