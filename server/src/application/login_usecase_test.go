package application

import (
	"testing"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/auth"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
)

func newLoginUsecase() (*LoginUsecase, *repositories.UserRepository, *auth.JwtService) {
	repo := repositories.NewUserRepository()
	jwtSvc := auth.NewJwtService("test-secret", time.Hour*24, time.Hour)
	uc := NewLoginUsecase(jwtSvc, repo)
	return uc, repo, jwtSvc
}

func TestLoginSuccess(t *testing.T) {
	uc, repo, _ := newLoginUsecase()

	user, err := (&domain.User{}).CreateUser("user@example.com", "Caique", "Senha#Segura1")
	if err != nil {
		t.Fatalf("unexpected error creating user: %v", err)
	}
	if err := repo.Save(user); err != nil {
		t.Fatalf("unexpected error saving user: %v", err)
	}

	result, err := uc.Execute("user@example.com", "Senha#Segura1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.JwtAccess == "" || result.JwtRefresh == "" {
		t.Fatal("expected non-empty tokens")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	uc, repo, _ := newLoginUsecase()

	user, err := (&domain.User{}).CreateUser("user@example.com", "Caique", "Senha#Segura1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := repo.Save(user); err != nil {
		t.Fatalf("unexpected error saving user: %v", err)
	}

	if _, err := uc.Execute("user@example.com", "WrongPass#1"); err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestLoginEmailNotFound(t *testing.T) {
	uc, _, _ := newLoginUsecase()

	if _, err := uc.Execute("missing@example.com", "Senha#Segura1"); err == nil {
		t.Fatal("expected error for missing email")
	}
}
