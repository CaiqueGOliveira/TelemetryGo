package valueobjects

import (
	"errors"
	"net/mail"
	"strings"
)

type Email struct {
	text string
}

func CreateEmail(email string) (*Email, error) {
	email = strings.TrimSpace(email)

	if !validateEmail(email) {
		return nil, errors.New("Invalid Email Address")
	}

	return &Email{
		text: email,
	}, nil
}

func (e *Email) Text() string {
	return e.text
}

func validateEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	if addr.Name != "" || addr.Address != email {
		return false
	}

	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 {
		return false
	}

	local, domain := parts[0], parts[1]
	if len(local) == 0 || len(domain) == 0 {
		return false
	}

	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}
