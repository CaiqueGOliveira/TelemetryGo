package valueobjects

import "testing"

func TestCreateEmailValid(t *testing.T) {
	email, err := CreateEmail("user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if email.Text() != "user@example.com" {
		t.Errorf("unexpected text %q", email.Text())
	}
}

func TestCreateEmailTrimsWhitespace(t *testing.T) {
	email, err := CreateEmail("  user@example.com  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if email.Text() != "user@example.com" {
		t.Errorf("expected trimmed email, got %q", email.Text())
	}
}

func TestCreateEmailInvalid(t *testing.T) {
	for _, candidate := range []string{
		"",
		"not-an-email",
		"user@localhost",
		"Name <user@example.com>",
		"user@example",
		"@example.com",
		"user@",
	} {
		if _, err := CreateEmail(candidate); err == nil {
			t.Errorf("expected error for %q", candidate)
		}
	}
}
