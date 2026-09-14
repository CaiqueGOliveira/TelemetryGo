package domain

import (
	"strings"
	"testing"
)

func TestCreateUser(t *testing.T) {
	user, err := CreateUser("user@example.com", "Caique", "Senha#Segura1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Id.String() == "" {
		t.Fatal("expected non-empty id")
	}
	if user.Email.Text() != "user@example.com" {
		t.Errorf("unexpected email %s", user.Email.Text())
	}
	if user.Name != "Caique" {
		t.Errorf("unexpected name %s", user.Name)
	}
	if !strings.HasPrefix(user.ApiKey, "tg_") {
		t.Errorf("expected api key to start with tg_, got %s", user.ApiKey)
	}
}

func TestCreateUserInvalidEmail(t *testing.T) {
	if _, err := CreateUser("not-an-email", "Caique", "Senha#Segura1"); err == nil {
		t.Fatal("expected error for invalid email")
	}
}

func TestCreateUserWeakPassword(t *testing.T) {
	if _, err := CreateUser("user@example.com", "Caique", "weak"); err == nil {
		t.Fatal("expected error for weak password")
	}
}

func TestUpdateProfile(t *testing.T) {
	user, _ := CreateUser("user@example.com", "Caique", "Senha#Segura1")

	if err := user.UpdateProfile("Novo Nome", "novo@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Name != "Novo Nome" {
		t.Errorf("expected new name, got %s", user.Name)
	}
	if user.Email.Text() != "novo@example.com" {
		t.Errorf("expected new email, got %s", user.Email.Text())
	}
}

func TestUpdateProfileEmptyName(t *testing.T) {
	user, _ := CreateUser("user@example.com", "Caique", "Senha#Segura1")

	if err := user.UpdateProfile("", "user@example.com"); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestUpdateProfileInvalidEmail(t *testing.T) {
	user, _ := CreateUser("user@example.com", "Caique", "Senha#Segura1")

	if err := user.UpdateProfile("Caique", "not-an-email"); err == nil {
		t.Fatal("expected error for invalid email")
	}
}

func TestChangePassword(t *testing.T) {
	user, _ := CreateUser("user@example.com", "Caique", "Senha#Segura1")

	if err := user.ChangePassword("wrong-password", "NovaSenha#1"); err == nil {
		t.Fatal("expected error for wrong current password")
	}

	if err := user.ChangePassword("Senha#Segura1", "NovaSenha#1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ok, err := user.Hash.Verify("NovaSenha#1")
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}
	if !ok {
		t.Fatal("expected new password to verify")
	}
}

func TestChangePasswordInvalidNewPassword(t *testing.T) {
	user, _ := CreateUser("user@example.com", "Caique", "Senha#Segura1")

	if err := user.ChangePassword("Senha#Segura1", "short"); err == nil {
		t.Fatal("expected error for weak new password")
	}
}

func TestRegenerateApiKey(t *testing.T) {
	user, _ := CreateUser("user@example.com", "Caique", "Senha#Segura1")
	oldKey := user.ApiKey

	user.RegenerateApiKey()

	if user.ApiKey == oldKey {
		t.Fatal("expected api key to change")
	}
	if !strings.HasPrefix(user.ApiKey, "tg_") {
		t.Errorf("expected api key to start with tg_, got %s", user.ApiKey)
	}
}
