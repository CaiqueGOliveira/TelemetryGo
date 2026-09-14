package valueobjects

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Time    uint32 = 1
	argon2Memory  uint32 = 64 * 1024
	argon2Threads uint8  = 4
	argon2KeyLen  uint32 = 32
	argon2SaltLen uint32 = 16
	argon2Version uint32 = 0x13
	argon2Prefix         = "$argon2id$v=19$m=65536,t=1,p=4"
)

var (
	lowercaseRegex = regexp.MustCompile(`[a-z]`)
	uppercaseRegex = regexp.MustCompile(`[A-Z]`)
	specialRegex   = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

type PasswordHashed struct {
	password string
}

func (h *PasswordHashed) CreateHash(text string) (*PasswordHashed, error) {
	validations := h.validatePassword(text)
	if len(validations) > 0 {
		return nil, errors.New(strings.Join(validations, "\n"))
	}

	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	hash := argon2.IDKey(
		[]byte(text),
		salt,
		argon2Time,
		argon2Memory,
		argon2Threads,
		argon2KeyLen,
	)

	encoded := fmt.Sprintf("%s$%s$%s",
		argon2Prefix,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return &PasswordHashed{
		password: encoded,
	}, nil
}

func (h *PasswordHashed) GetPasswordHash() string {
	return h.password
}

func NewPasswordHashedFromHash(hash string) *PasswordHashed {
	return &PasswordHashed{password: hash}
}

func (h *PasswordHashed) Verify(text string) (bool, error) {
	parts := strings.Split(h.password, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	var version uint32
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, err
	}
	if version != argon2Version {
		return false, errors.New("incompatible argon2 version")
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey(
		[]byte(text),
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(hash)),
	)

	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func (h *PasswordHashed) validatePassword(text string) []string {
	var messages []string

	length := len([]byte(text))
	if length < 8 {
		messages = append(messages, "the password must be at least 8 characters long")
	}
	if length > 64 {
		messages = append(messages, "the password must be at most 64 characters long")
	}

	if !lowercaseRegex.MatchString(text) {
		messages = append(messages, "the password must contain at least 1 lowercase letter")
	}

	if !uppercaseRegex.MatchString(text) {
		messages = append(messages, "the password must contain at least 1 uppercase letter")
	}

	if !specialRegex.MatchString(text) {
		messages = append(messages, "the password must contain at least 1 special character")
	}

	return messages
}
