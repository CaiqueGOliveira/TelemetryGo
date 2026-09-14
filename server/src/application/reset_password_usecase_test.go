package application

import (
	"log"
	"testing"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/auth"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
)

func newResetPasswordUsecase() (*ResetPasswordUsecase, *repositories.UserRepository, *auth.JwtService) {
	repo := repositories.NewUserRepository()
	jwtSvc, err := auth.NewJwtService("test-secret", time.Hour*24, time.Hour)
	if err != nil {
		log.Fatalf("failed to initialize jwt service: %v", err)
	}
	return NewResetPasswordUsecase(repo, jwtSvc), repo, jwtSvc
}

func TestResetPasswordSuccess(t *testing.T) {
	uc, repo, jwtSvc := newResetPasswordUsecase()

	user, err := createUserInRepo(repo, "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error creating user: %v", err)
	}
	originalHash := user.Hash

	token, err := jwtSvc.GenerateResetToken(user.Id, 30*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	if err := uc.Execute(token, "NovaSenha#Segura1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := repo.FindByEmail("user@example.com")
	if err != nil {
		t.Fatalf("unexpected error finding user: %v", err)
	}

	if updated.Hash == originalHash {
		t.Fatal("expected password hash to change after reset")
	}

	loginUc := newLoginUsecaseWithRepo(repo)
	if _, err := loginUc.Execute("user@example.com", "NovaSenha#Segura1"); err != nil {
		t.Fatalf("expected login with new password to succeed: %v", err)
	}
	if _, err := loginUc.Execute("user@example.com", "Senha#Segura1"); err == nil {
		t.Fatal("expected login with old password to fail")
	}
}

func TestResetPasswordAccessTokenRejected(t *testing.T) {
	uc, repo, jwtSvc := newResetPasswordUsecase()

	user, err := createUserInRepo(repo, "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error creating user: %v", err)
	}

	token, err := jwtSvc.GenerateToken(user.Id, "access")
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	if err := uc.Execute(token, "NovaSenha#Segura1"); err == nil {
		t.Fatal("expected error for access token")
	}
}

func TestResetPasswordInvalidToken(t *testing.T) {
	uc, _, _ := newResetPasswordUsecase()

	if err := uc.Execute("not.a.valid.token", "NovaSenha#Segura1"); err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestResetPasswordWeakPassword(t *testing.T) {
	uc, repo, jwtSvc := newResetPasswordUsecase()

	user, err := createUserInRepo(repo, "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error creating user: %v", err)
	}

	token, err := jwtSvc.GenerateResetToken(user.Id, 30*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	if err := uc.Execute(token, "fraca"); err == nil {
		t.Fatal("expected error for weak password")
	}
}
