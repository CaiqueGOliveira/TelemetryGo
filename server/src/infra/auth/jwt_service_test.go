package auth

import (
	"log"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func newTestService() *JwtService {
	token, err := NewJwtService(
		"test-secret",
		time.Hour*24*7,
		time.Hour/6,
	)
	if err != nil {
		log.Fatalf("failed to initialize jwt service: %v", err)
	}

	return token
}

func TestGenerateTokenAccess(t *testing.T) {
	svc := newTestService()
	userID := uuid.New()

	tokenString, err := svc.GenerateToken(userID, "access")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokenString == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := svc.VerifyToken(tokenString)
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}

	sub := claims["sub"]
	idStr, ok := sub.(string)
	if !ok {
		t.Fatalf("expected sub to be string, got %T", sub)
	}
	if idStr != userID.String() {
		t.Errorf("expected sub %s, got %s", userID.String(), idStr)
	}

	if typ := claims["typ"]; typ != "access" {
		t.Errorf("expected typ access, got %v", typ)
	}
}

func TestGenerateTokenRefresh(t *testing.T) {
	svc := newTestService()
	userID := uuid.New()

	tokenString, err := svc.GenerateToken(userID, "refresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := svc.VerifyToken(tokenString)
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}

	if typ := claims["typ"]; typ != "refresh" {
		t.Errorf("expected typ refresh, got %v", typ)
	}
}

func TestGenerateTokenUnknownType(t *testing.T) {
	svc := newTestService()

	_, err := svc.GenerateToken(uuid.New(), "unknown")
	if err == nil {
		t.Fatal("expected error for unknown token type")
	}
}

func TestGenerateResetToken(t *testing.T) {
	svc := newTestService()
	userID := uuid.New()

	tokenString, err := svc.GenerateResetToken(userID, 30*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := svc.VerifyToken(tokenString)
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}

	if typ := claims["typ"]; typ != "reset-password" {
		t.Errorf("expected typ reset-password, got %v", typ)
	}

	sub, _ := claims["sub"].(string)
	if sub != userID.String() {
		t.Errorf("expected sub %s, got %s", userID.String(), sub)
	}
}

func TestVerifyTokenInvalid(t *testing.T) {
	svc := newTestService()

	if _, err := svc.VerifyToken("not.a.jwt"); err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestVerifyTokenWrongSecret(t *testing.T) {
	svc := newTestService()
	other, err := NewJwtService("other-secret", time.Hour, time.Hour)
	if err != nil {
		log.Fatalf("failed to initialize jwt service: %v", err)
	}

	tokenString, err := svc.GenerateToken(uuid.New(), "access")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := other.VerifyToken(tokenString); err == nil {
		t.Fatal("expected error verifying token with wrong secret")
	}
}

func TestVerifyTokenExpired(t *testing.T) {
	claims := &jwt.MapClaims{
		"sub": uuid.New().String(),
		"exp": time.Now().Add(-time.Hour).Unix(),
		"typ": "access",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc := newTestService()
	if _, err := svc.VerifyToken(tokenString); err == nil {
		t.Fatal("expected error for expired token")
	}
}
