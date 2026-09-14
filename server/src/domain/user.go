package domain

import (
	"errors"

	vo "github.com/CaiqueGOliveira/TelemetryGo/src/domain/valueobjects"
	"github.com/google/uuid"
)

type User struct {
	Id     uuid.UUID
	Email  *vo.Email
	Name   string
	Hash   *vo.PasswordHashed
	ApiKey string
}

func newApiKey() string {
	return "tg_" + uuid.New().String()
}

func NewUser(id uuid.UUID, email string, name string, password string) (*User, error) {
	validEmail, err := vo.CreateEmail(email)
	if err != nil {
		return nil, err
	}

	passwordHash := &vo.PasswordHashed{}
	hashed, err := passwordHash.CreateHash(password)
	if err != nil {
		return nil, err
	}

	return &User{
		Id:     id,
		Email:  validEmail,
		Name:   name,
		Hash:   hashed,
		ApiKey: newApiKey(),
	}, nil
}

func CreateUser(email string, name string, password string) (*User, error) {
	return NewUser(uuid.New(), email, name, password)
}

func (u *User) UpdateProfile(name string, email string) error {
	if name == "" {
		return errors.New("name is required")
	}

	validEmail, err := vo.CreateEmail(email)
	if err != nil {
		return err
	}

	u.Name = name
	u.Email = validEmail
	return nil
}

func (u *User) ChangePassword(currentPassword string, newPassword string) error {
	match, err := u.Hash.Verify(currentPassword)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("current password does not match")
	}

	newHash, err := (&vo.PasswordHashed{}).CreateHash(newPassword)
	if err != nil {
		return err
	}

	u.Hash = newHash
	return nil
}

func (u *User) ResetPassword(newPassword string) error {
	newHash, err := (&vo.PasswordHashed{}).CreateHash(newPassword)
	if err != nil {
		return err
	}

	u.Hash = newHash
	return nil
}

func (u *User) RegenerateApiKey() {
	u.ApiKey = newApiKey()
}
