package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setup(provider *auth.JwtService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	protected := r.Group("/")
	protected.Use(Auth(provider))
	protected.GET("/me", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})

	return r
}

func TestAuthValidToken(t *testing.T) {
	svc, err := auth.NewJwtService("test-secret", time.Hour, time.Hour)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	r := setup(svc)

	tokenString, err := svc.GenerateToken(uuid.New(), "access")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthMissingHeader(t *testing.T) {
	svc, err := auth.NewJwtService("test-secret", time.Hour, time.Hour)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	r := setup(svc)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthMalformedHeader(t *testing.T) {
	svc, err := auth.NewJwtService("test-secret", time.Hour, time.Hour)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	r := setup(svc)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Token abc")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthInvalidToken(t *testing.T) {
	svc, err := auth.NewJwtService("test-secret", time.Hour, time.Hour)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	r := setup(svc)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer not.a.jwt")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
