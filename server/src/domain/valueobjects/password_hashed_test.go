package valueobjects

import (
	"strings"
	"testing"
)

func TestCreateHashAndVerify(t *testing.T) {
	pw := &PasswordHashed{}
	hashed, err := pw.CreateHash("Senha#Segura1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.HasPrefix(hashed.password, "$argon2id$v=19$") == false {
		t.Fatalf("expected argon2id PHC prefix, got: %s", hashed.password)
	}

	ok, err := hashed.Verify("Senha#Segura1")
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify")
	}

	ok, err = hashed.Verify("SENHA#segura2")
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}
	if ok {
		t.Fatal("expected wrong password to fail verification")
	}
}

func TestValidatePasswordMessages(t *testing.T) {
	pw := &PasswordHashed{}

	_, err := pw.CreateHash("12345678")
	if err == nil {
		t.Fatal("expected error for password without required characters")
	}
	for _, msg := range []string{
		"lowercase letter",
		"uppercase letter",
		"special character",
	} {
		if !strings.Contains(err.Error(), msg) {
			t.Errorf("expected error to contain %q, got: %v", msg, err)
		}
	}
}

func TestValidatePasswordPartial(t *testing.T) {
	pw := &PasswordHashed{}

	_, err := pw.CreateHash("abcdefgh")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "lowercase letter") {
		t.Error("lowercase present, should not complain about lowercase")
	}
	if !strings.Contains(err.Error(), "uppercase letter") {
		t.Error("missing uppercase, should complain")
	}
	if !strings.Contains(err.Error(), "special character") {
		t.Error("missing special, should complain")
	}
}
