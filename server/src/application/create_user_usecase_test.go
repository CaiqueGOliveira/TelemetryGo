package application

import (
	"log"
	"strings"
	"testing"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/auth"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/repositories"
)

func newCreateUserUsecase() (*CreateUserUsecase, *repositories.UserRepository) {
	repo := repositories.NewUserRepository()
	jwtSvc, err := auth.NewJwtService("test-secret", time.Hour*24, time.Hour)
	if err != nil {
		log.Fatalf("failed to initialize jwt service: %v", err)
	}
	return NewCreateUserUsecase(repo, jwtSvc), repo
}

func TestCreateUserSuccess(t *testing.T) {
	uc, repo := newCreateUserUsecase()

	in := &dtos.UserCreateRequestDto{
		Name:     "Caique",
		Email:    "user@example.com",
		Password: "Senha#Segura1",
	}

	result, err := uc.Execute(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user := result.User
	if user.Name != in.Name {
		t.Errorf("expected name %s, got %s", in.Name, user.Name)
	}
	if user.Email.Text() != in.Email {
		t.Errorf("expected email %s, got %s", in.Email, user.Email.Text())
	}
	if result.JwtAccess == "" || result.JwtRefresh == "" {
		t.Fatal("expected non-empty tokens")
	}
	if user.ApiKey == "" {
		t.Fatal("expected non-empty api key")
	}
	if !strings.HasPrefix(user.ApiKey, "tg_") {
		t.Errorf("expected api key to start with tg_, got %s", user.ApiKey)
	}

	saved, err := repo.FindByEmail(in.Email)
	if err != nil {
		t.Fatalf("expected user to be saved, got: %v", err)
	}
	if saved.Id != user.Id {
		t.Errorf("expected saved user id to match, got %s != %s", saved.Id, user.Id)
	}
	if saved.ApiKey != user.ApiKey {
		t.Errorf("expected saved api key to match, got %s != %s", saved.ApiKey, user.ApiKey)
	}

	found, err := repo.FindByApiKey(user.ApiKey)
	if err != nil {
		t.Fatalf("expected user to be found by api key, got: %v", err)
	}
	if found.Id != user.Id {
		t.Errorf("expected api key lookup to return same user, got %s != %s", found.Id, user.Id)
	}
}

func TestCreateUserInvalidEmail(t *testing.T) {
	uc, _ := newCreateUserUsecase()

	in := &dtos.UserCreateRequestDto{
		Name:     "Caique",
		Email:    "not-an-email",
		Password: "Senha#Segura1",
	}

	if _, err := uc.Execute(in); err == nil {
		t.Fatal("expected error for invalid email")
	}
}
