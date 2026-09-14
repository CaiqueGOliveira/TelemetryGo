package integration

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestUpdateProfile(t *testing.T) {
	r := setupRouter(t)
	createUserAndGetApiKey(t, r)

	token := loginAndGetAccessToken(t, r, "user@example.com")

	resp := doJSON(t, r, http.MethodPut, "/api/v1/me", map[string]string{
		"name":  "Caique Atualizado",
		"email": "novo@example.com",
	}, token)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d: %s", resp.Code, resp.Body.String())
	}

	meResp := doJSON(t, r, http.MethodGet, "/api/v1/me", nil, token)
	if meResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on /me, got %d: %s", meResp.Code, meResp.Body.String())
	}

	var me struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(meResp.Body.Bytes(), &me); err != nil {
		t.Fatalf("unexpected error decoding me body: %v", err)
	}
	if me.Name != "Caique Atualizado" {
		t.Errorf("expected updated name, got %q", me.Name)
	}
	if me.Email != "novo@example.com" {
		t.Errorf("expected updated email, got %q", me.Email)
	}
}

func TestUpdateProfileDuplicateEmail(t *testing.T) {
	r := setupRouter(t)
	createUserAndGetApiKeyWithEmail(t, r, "alice@example.com")

	token := loginAndGetAccessToken(t, r, "alice@example.com")
	createUserAndGetApiKeyWithEmail(t, r, "bob@example.com")

	resp := doJSON(t, r, http.MethodPut, "/api/v1/me", map[string]string{
		"name":  "Bob",
		"email": "bob@example.com",
	}, token)
	if resp.Code != http.StatusConflict {
		t.Fatalf("expected 409 on duplicate email update, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestChangePassword(t *testing.T) {
	r := setupRouter(t)
	createUserAndGetApiKey(t, r)

	token := loginAndGetAccessToken(t, r, "user@example.com")

	wrong := doJSON(t, r, http.MethodPost, "/api/v1/auth/change-password", map[string]string{
		"current_password": "Senha#Errada1",
		"new_password":     "NovaSenha#1",
	}, token)
	if wrong.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on wrong current password, got %d: %s", wrong.Code, wrong.Body.String())
	}

	ok := doJSON(t, r, http.MethodPost, "/api/v1/auth/change-password", map[string]string{
		"current_password": "Senha#Segura1",
		"new_password":     "NovaSenha#1",
	}, token)
	if ok.Code != http.StatusOK {
		t.Fatalf("expected 200 on change password, got %d: %s", ok.Code, ok.Body.String())
	}

	oldLogin := doJSON(t, r, http.MethodPost, "/api/v1/login", map[string]string{
		"email":    "user@example.com",
		"password": "Senha#Segura1",
	}, "")
	if oldLogin.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with old password, got %d", oldLogin.Code)
	}

	newLogin := doJSON(t, r, http.MethodPost, "/api/v1/login", map[string]string{
		"email":    "user@example.com",
		"password": "NovaSenha#1",
	}, "")
	if newLogin.Code != http.StatusOK {
		t.Fatalf("expected 200 with new password, got %d: %s", newLogin.Code, newLogin.Body.String())
	}
}

func TestRotateApiKey(t *testing.T) {
	r := setupRouter(t)
	apiKey := createUserAndGetApiKey(t, r)

	events := []map[string]string{
		{"type": "deploy", "service": "api-gateway", "message": "msg", "severity": "info"},
	}
	first := doJSON(t, r, http.MethodPost, "/api/v1/events", events, apiKey)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201 before rotation, got %d: %s", first.Code, first.Body.String())
	}

	token := loginAndGetAccessToken(t, r, "user@example.com")

	rotateResp := doJSON(t, r, http.MethodPost, "/api/v1/auth/rotate-api-key", nil, token)
	if rotateResp.Code != http.StatusOK {
		t.Fatalf("expected 200 on rotate, got %d: %s", rotateResp.Code, rotateResp.Body.String())
	}

	var body struct {
		ApiKey string `json:"api_key"`
	}
	if err := json.Unmarshal(rotateResp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected error decoding rotate body: %v", err)
	}
	if body.ApiKey == "" || body.ApiKey == apiKey {
		t.Fatalf("expected a new api key, got %q", body.ApiKey)
	}

	after := doJSON(t, r, http.MethodPost, "/api/v1/events", events, body.ApiKey)
	if after.Code != http.StatusCreated {
		t.Fatalf("expected 201 with new key, got %d: %s", after.Code, after.Body.String())
	}

	oldKey := doJSON(t, r, http.MethodPost, "/api/v1/events", events, apiKey)
	if oldKey.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with old key, got %d", oldKey.Code)
	}
}

func TestDeleteAccount(t *testing.T) {
	r := setupRouter(t)
	createUserAndGetApiKey(t, r)

	token := loginAndGetAccessToken(t, r, "user@example.com")

	delResp := doJSON(t, r, http.MethodDelete, "/api/v1/users/me", nil, token)
	if delResp.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on delete account, got %d: %s", delResp.Code, delResp.Body.String())
	}

	after := doJSON(t, r, http.MethodPost, "/api/v1/login", map[string]string{
		"email":    "user@example.com",
		"password": "Senha#Segura1",
	}, "")
	if after.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after account deletion, got %d", after.Code)
	}
}

func TestNotFoundHandler(t *testing.T) {
	r := setupRouter(t)

	resp := doJSON(t, r, http.MethodGet, "/api/v1/nonexistent", nil, "")
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), "route not found") {
		t.Fatalf("expected JSON error body, got: %s", resp.Body.String())
	}
}
