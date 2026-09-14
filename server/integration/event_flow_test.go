package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestEventIngestAndList(t *testing.T) {
	r := setupRouter(t)
	apiKey := createUserAndGetApiKey(t, r)

	events := []map[string]string{
		{"type": "deploy", "service": "api-gateway", "message": "Deploy realizado com sucesso", "severity": "info"},
		{"type": "threshold", "service": "worker", "message": "Uso de memória acima do limite", "severity": "warning"},
		{"type": "error", "service": "auth-service", "message": "Erro ao autenticar usuário", "severity": "critical"},
	}

	createResp := doJSON(t, r, http.MethodPost, "/api/v1/events", events, apiKey)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on ingest, got %d: %s", createResp.Code, createResp.Body.String())
	}

	var ack map[string]any
	if err := json.Unmarshal(createResp.Body.Bytes(), &ack); err != nil {
		t.Fatalf("unexpected error decoding ingest response: %v", err)
	}
	if float64(len(events)) != ack["accepted"] {
		t.Fatalf("expected %d events accepted, got %v", len(events), ack["accepted"])
	}

	token := loginAndGetAccessToken(t, r, "user@example.com")

	listResp := doJSON(t, r, http.MethodGet, "/api/v1/events", nil, token)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on list, got %d: %s", listResp.Code, listResp.Body.String())
	}

	var listed []map[string]any
	if err := json.Unmarshal(listResp.Body.Bytes(), &listed); err != nil {
		t.Fatalf("unexpected error decoding list response: %v", err)
	}
	if len(listed) != len(events) {
		t.Fatalf("expected %d events listed, got %d", len(events), len(listed))
	}
}

func TestEventTenantIsolation(t *testing.T) {
	r := setupRouter(t)

	apiKeyA := createUserAndGetApiKeyWithEmail(t, r, "alice@example.com")
	events := []map[string]string{
		{"type": "deploy", "service": "api-gateway", "message": "evento da alice", "severity": "info"},
	}
	createResp := doJSON(t, r, http.MethodPost, "/api/v1/events", events, apiKeyA)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on ingest, got %d: %s", createResp.Code, createResp.Body.String())
	}

	tokenA := loginAndGetAccessToken(t, r, "alice@example.com")
	listA := doJSON(t, r, http.MethodGet, "/api/v1/events", nil, tokenA)
	var listedA []map[string]any
	if err := json.Unmarshal(listA.Body.Bytes(), &listedA); err != nil {
		t.Fatalf("unexpected error decoding list response: %v", err)
	}
	if len(listedA) != 1 {
		t.Fatalf("expected 1 event for alice, got %d", len(listedA))
	}

	createUserAndGetApiKeyWithEmail(t, r, "bob@example.com")
	tokenB := loginAndGetAccessToken(t, r, "bob@example.com")
	listB := doJSON(t, r, http.MethodGet, "/api/v1/events", nil, tokenB)
	var listedB []map[string]any
	if err := json.Unmarshal(listB.Body.Bytes(), &listedB); err != nil {
		t.Fatalf("unexpected error decoding list response: %v", err)
	}
	if len(listedB) != 0 {
		t.Fatalf("expected 0 events for bob, got %d", len(listedB))
	}
}

func TestEventListRequiresToken(t *testing.T) {
	r := setupRouter(t)

	resp := doJSON(t, r, http.MethodGet, "/api/v1/events", nil, "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on list without token, got %d", resp.Code)
	}
}

func TestEventIngestInvalidSeverity(t *testing.T) {
	r := setupRouter(t)
	apiKey := createUserAndGetApiKey(t, r)

	events := []map[string]string{
		{"type": "deploy", "service": "api-gateway", "message": "msg", "severity": "bogus"},
	}

	resp := doJSON(t, r, http.MethodPost, "/api/v1/events", events, apiKey)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on invalid severity, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestEventIngestInvalidTimestamp(t *testing.T) {
	r := setupRouter(t)
	apiKey := createUserAndGetApiKey(t, r)

	events := []map[string]string{
		{"type": "deploy", "service": "api-gateway", "message": "msg", "severity": "info", "timestamp": "not-a-timestamp"},
	}

	resp := doJSON(t, r, http.MethodPost, "/api/v1/events", events, apiKey)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on invalid timestamp, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestEventIngestMissingField(t *testing.T) {
	r := setupRouter(t)
	apiKey := createUserAndGetApiKey(t, r)

	events := []map[string]string{
		{"type": "deploy", "service": "api-gateway", "severity": "info"},
	}

	resp := doJSON(t, r, http.MethodPost, "/api/v1/events", events, apiKey)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on missing message, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestEventIngestWithoutApiKey(t *testing.T) {
	r := setupRouter(t)

	events := []map[string]string{
		{"type": "deploy", "service": "api-gateway", "message": "msg", "severity": "info"},
	}

	resp := doJSON(t, r, http.MethodPost, "/api/v1/events", events, "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without api key, got %d: %s", resp.Code, resp.Body.String())
	}
}
