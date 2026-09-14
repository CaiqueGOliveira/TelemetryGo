package application

import (
	"log"
	"testing"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/auth"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
)

func newForgotPasswordUsecase() (*ForgotPasswordUsecase, *repositories.UserRepository) {
	repo := repositories.NewUserRepository()
	jwtSvc, err := auth.NewJwtService("test-secret", time.Hour*24, time.Hour)
	if err != nil {
		log.Fatalf("failed to initialize jwt service: %v", err)
	}
	return NewForgotPasswordUsecase(repo, jwtSvc, 30*time.Minute), repo
}

func TestForgotPasswordExistingEmailReturnsToken(t *testing.T) {
	uc, repo := newForgotPasswordUsecase()

	user, err := domain.CreateUser("user@example.com", "Caique", "Senha#Segura1")
	if err != nil {
		t.Fatalf("unexpected error creating user: %v", err)
	}
	if err := repo.Save(user); err != nil {
		t.Fatalf("unexpected error saving user: %v", err)
	}

	token, err := uc.Execute("user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty reset token")
	}
}

func TestForgotPasswordMissingEmailReturnsEmptyToken(t *testing.T) {
	uc, _ := newForgotPasswordUsecase()

	token, err := uc.Execute("missing@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "" {
		t.Fatalf("expected empty token for missing email, got %s", token)
	}
}

func createUserInRepo(repo *repositories.UserRepository, email string) (*domain.User, error) {
	user, err := domain.CreateUser(email, "Caique", "Senha#Segura1")
	if err != nil {
		return nil, err
	}
	if err := repo.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

func newLoginUsecaseWithRepo(repo *repositories.UserRepository) *LoginUsecase {
	jwtSvc, err := auth.NewJwtService("test-secret", time.Hour*24, time.Hour)
	if err != nil {
		log.Fatalf("failed to initialize jwt service: %v", err)
	}
	return NewLoginUsecase(jwtSvc, repo)
}
