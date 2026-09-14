package integration

import (
	"net/http"
	"testing"
)

func TestDeleteEvent(t *testing.T) {
	r := setupRouter(t)
	apiKeyA := createUserAndGetApiKey(t, r)
	token := loginAndGetAccessToken(t, r, "user@example.com")

	eventID := "00000000-0000-0000-0000-000000000011"

	events := []map[string]string{
		{"id": eventID, "type": "deploy", "service": "api-gateway", "message": "a", "severity": "info", "timestamp": "2026-01-01T10:00:00Z"},
	}
	ingestResp := doJSON(t, r, http.MethodPost, "/api/v1/events", events, apiKeyA)
	if ingestResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on ingest, got %d: %s", ingestResp.Code, ingestResp.Body.String())
	}

	del := doJSON(t, r, http.MethodDelete, "/api/v1/events/"+eventID, nil, token)
	if del.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on delete, got %d: %s", del.Code, del.Body.String())
	}

	listAndAssertIds(t, r, token, "/api/v1/events", []string{})

	again := doJSON(t, r, http.MethodDelete, "/api/v1/events/"+eventID, nil, token)
	if again.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on second delete, got %d: %s", again.Code, again.Body.String())
	}

	invalid := doJSON(t, r, http.MethodDelete, "/api/v1/events/not-a-uuid", nil, token)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on invalid id, got %d", invalid.Code)
	}
}

func TestDeleteMetric(t *testing.T) {
	r := setupRouter(t)
	apiKeyA := createUserAndGetApiKey(t, r)
	token := loginAndGetAccessToken(t, r, "user@example.com")

	metricID := "00000000-0000-0000-0000-000000000111"

	metrics := []map[string]string{
		{"id": metricID, "name": "cpu_usage", "service": "api-gateway", "value": "10", "unit": "%", "status": "ok", "timestamp": "2026-01-01T10:00:00Z"},
	}
	ingestResp := doJSON(t, r, http.MethodPost, "/api/v1/metrics", metrics, apiKeyA)
	if ingestResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on ingest, got %d: %s", ingestResp.Code, ingestResp.Body.String())
	}

	del := doJSON(t, r, http.MethodDelete, "/api/v1/metrics/"+metricID, nil, token)
	if del.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on delete, got %d: %s", del.Code, del.Body.String())
	}

	listAndAssertIds(t, r, token, "/api/v1/metrics", []string{})

	again := doJSON(t, r, http.MethodDelete, "/api/v1/metrics/"+metricID, nil, token)
	if again.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on second delete, got %d: %s", again.Code, again.Body.String())
	}
}

func TestDeleteEventTenantIsolation(t *testing.T) {
	r := setupRouter(t)
	apiKeyA := createUserAndGetApiKey(t, r)
	tokenA := loginAndGetAccessToken(t, r, "user@example.com")

	eventID := "00000000-0000-0000-0000-000000000012"

	events := []map[string]string{
		{"id": eventID, "type": "deploy", "service": "api-gateway", "message": "a", "severity": "info", "timestamp": "2026-01-01T10:00:00Z"},
	}
	ingestResp := doJSON(t, r, http.MethodPost, "/api/v1/events", events, apiKeyA)
	if ingestResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on ingest, got %d: %s", ingestResp.Code, ingestResp.Body.String())
	}

	createUserAndGetApiKeyWithEmail(t, r, "outro@example.com")
	tokenB := loginAndGetAccessToken(t, r, "outro@example.com")

	del := doJSON(t, r, http.MethodDelete, "/api/v1/events/"+eventID, nil, tokenB)
	if del.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when deleting another tenant's event, got %d: %s", del.Code, del.Body.String())
	}

	listAndAssertIds(t, r, tokenB, "/api/v1/events", []string{})
	listAndAssertIds(t, r, tokenA, "/api/v1/events", []string{eventID})
}

func TestDeleteWithoutAuth(t *testing.T) {
	r := setupRouter(t)
	createUserAndGetApiKey(t, r)

	resp := doJSON(t, r, http.MethodDelete, "/api/v1/events/anything-here", nil, "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", resp.Code)
	}
}
