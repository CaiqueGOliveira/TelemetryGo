package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestPasswordResetFlow(t *testing.T) {
	router := setupRouter(t)
	createUserAndGetApiKeyWithEmail(t, router, "user@example.com")

	forgot := doJSON(t, router, http.MethodPost, "/api/v1/auth/forgot-password",
		map[string]string{"email": "user@example.com"}, "")
	if forgot.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", forgot.Code, forgot.Body.String())
	}

	var forgotBody struct {
		ResetToken string `json:"reset_token"`
	}
	if err := json.Unmarshal(forgot.Body.Bytes(), &forgotBody); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	if forgotBody.ResetToken == "" {
		t.Fatal("expected reset_token for existing email")
	}

	reset := doJSON(t, router, http.MethodPost, "/api/v1/auth/reset-password",
		map[string]string{"token": forgotBody.ResetToken, "new_password": "NovaSenha#Segura1"}, "")
	if reset.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", reset.Code, reset.Body.String())
	}

	login := doJSON(t, router, http.MethodPost, "/api/v1/login",
		map[string]string{"email": "user@example.com", "password": "NovaSenha#Segura1"}, "")
	if login.Code != http.StatusOK {
		t.Fatalf("expected login with new password to succeed, got %d", login.Code)
	}

	oldLogin := doJSON(t, router, http.MethodPost, "/api/v1/login",
		map[string]string{"email": "user@example.com", "password": "Senha#Segura1"}, "")
	if oldLogin.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with old password, got %d", oldLogin.Code)
	}
}

func TestForgotPasswordDoesNotRevealMissingEmail(t *testing.T) {
	router := setupRouter(t)

	resp := doJSON(t, router, http.MethodPost, "/api/v1/auth/forgot-password",
		map[string]string{"email": "missing@example.com"}, "")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected decode: %v", err)
	}
	if _, ok := body["reset_token"]; ok {
		t.Fatal("expected no reset_token for missing email")
	}
}

func TestResetPasswordWithExpiredTokenFails(t *testing.T) {
	router := setupRouter(t)
	createUserAndGetApiKeyWithEmail(t, router, "user@example.com")

	resp := doJSON(t, router, http.MethodPost, "/api/v1/auth/reset-password",
		map[string]string{"token": "expired-or-invalid-token", "new_password": "NovaSenha#Segura1"}, "")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}
