package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestMetricIngestAndList(t *testing.T) {
	r := setupRouter(t)
	apiKey := createUserAndGetApiKey(t, r)

	metrics := []map[string]string{
		{"name": "Latency", "service": "auth-service", "value": "42", "unit": "ms", "status": "ok"},
		{"name": "CPU", "service": "api-gateway", "value": "38", "unit": "%", "status": "ok"},
		{"name": "Memory", "service": "worker", "value": "1.2", "unit": "GB", "status": "warn"},
	}

	createResp := doJSON(t, r, http.MethodPost, "/api/v1/metrics", metrics, apiKey)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on ingest, got %d: %s", createResp.Code, createResp.Body.String())
	}

	var ack map[string]any
	if err := json.Unmarshal(createResp.Body.Bytes(), &ack); err != nil {
		t.Fatalf("unexpected error decoding ingest response: %v", err)
	}
	if float64(len(metrics)) != ack["accepted"] {
		t.Fatalf("expected %d metrics accepted, got %v", len(metrics), ack["accepted"])
	}

	token := loginAndGetAccessToken(t, r, "user@example.com")

	listResp := doJSON(t, r, http.MethodGet, "/api/v1/metrics", nil, token)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on list, got %d: %s", listResp.Code, listResp.Body.String())
	}

	var listed []map[string]any
	if err := json.Unmarshal(listResp.Body.Bytes(), &listed); err != nil {
		t.Fatalf("unexpected error decoding list response: %v", err)
	}
	if len(listed) != len(metrics) {
		t.Fatalf("expected %d metrics listed, got %d", len(metrics), len(listed))
	}
}

func TestMetricTenantIsolation(t *testing.T) {
	r := setupRouter(t)

	apiKeyA := createUserAndGetApiKeyWithEmail(t, r, "alice@example.com")
	metrics := []map[string]string{
		{"name": "CPU", "service": "api-gateway", "value": "38", "unit": "%", "status": "ok"},
	}
	createResp := doJSON(t, r, http.MethodPost, "/api/v1/metrics", metrics, apiKeyA)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 on ingest, got %d: %s", createResp.Code, createResp.Body.String())
	}

	tokenA := loginAndGetAccessToken(t, r, "alice@example.com")
	listA := doJSON(t, r, http.MethodGet, "/api/v1/metrics", nil, tokenA)
	var listedA []map[string]any
	if err := json.Unmarshal(listA.Body.Bytes(), &listedA); err != nil {
		t.Fatalf("unexpected error decoding list response: %v", err)
	}
	if len(listedA) != 1 {
		t.Fatalf("expected 1 metric for alice, got %d", len(listedA))
	}

	createUserAndGetApiKeyWithEmail(t, r, "bob@example.com")
	tokenB := loginAndGetAccessToken(t, r, "bob@example.com")
	listB := doJSON(t, r, http.MethodGet, "/api/v1/metrics", nil, tokenB)
	var listedB []map[string]any
	if err := json.Unmarshal(listB.Body.Bytes(), &listedB); err != nil {
		t.Fatalf("unexpected error decoding list response: %v", err)
	}
	if len(listedB) != 0 {
		t.Fatalf("expected 0 metrics for bob, got %d", len(listedB))
	}
}

func TestMetricListRequiresToken(t *testing.T) {
	r := setupRouter(t)

	resp := doJSON(t, r, http.MethodGet, "/api/v1/metrics", nil, "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on list without token, got %d", resp.Code)
	}
}

func TestMetricIngestInvalidStatus(t *testing.T) {
	r := setupRouter(t)
	apiKey := createUserAndGetApiKey(t, r)

	metrics := []map[string]string{
		{"name": "CPU", "service": "api-gateway", "value": "38", "unit": "%", "status": "bogus"},
	}

	resp := doJSON(t, r, http.MethodPost, "/api/v1/metrics", metrics, apiKey)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on invalid status, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestMetricIngestInvalidValue(t *testing.T) {
	r := setupRouter(t)
	apiKey := createUserAndGetApiKey(t, r)

	metrics := []map[string]string{
		{"name": "CPU", "service": "api-gateway", "value": "not-a-number", "unit": "%", "status": "ok"},
	}

	resp := doJSON(t, r, http.MethodPost, "/api/v1/metrics", metrics, apiKey)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on invalid value, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestMetricIngestMissingValue(t *testing.T) {
	r := setupRouter(t)
	apiKey := createUserAndGetApiKey(t, r)

	metrics := []map[string]string{
		{"name": "CPU", "service": "api-gateway", "unit": "%", "status": "ok"},
	}

	resp := doJSON(t, r, http.MethodPost, "/api/v1/metrics", metrics, apiKey)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on missing value, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestMetricIngestWithoutApiKey(t *testing.T) {
	r := setupRouter(t)

	metrics := []map[string]string{
		{"name": "CPU", "service": "api-gateway", "value": "38", "unit": "%", "status": "ok"},
	}

	resp := doJSON(t, r, http.MethodPost, "/api/v1/metrics", metrics, "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without api key, got %d: %s", resp.Code, resp.Body.String())
	}
}
