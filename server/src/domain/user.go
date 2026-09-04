package domain

import (
	"errors"

	vo "github.com/CaiqueGOliveira/TelemetryGo/src/domain/valueobjects"
	"github.com/google/uuid"
)

type User struct {
	Id    uuid.UUID
	Email *vo.Email
	Name  string
	Hash  *vo.PasswordHashed
}

func NewUser(id uuid.UUID, email string, name string, password string) (*User, error) {
	valid_email, err := vo.CreateEmail(email)
	if err != nil {
		return nil, errors.New("Invalid Email Address")
	}

	passwordHash := &vo.PasswordHashed{}
	hashed, err := passwordHash.CreateHash(password)
	if err != nil {

		return nil, err
	}

	return &User{
		Id:    id,
		Email: valid_email,
		Name:  name,
		Hash:  hashed,
	}, nil
}

func (u *User) CreateUser(email string, name string, password string) (*User, error) {
	valid_email, err := vo.CreateEmail(email)
	if err != nil {
		return nil, errors.New("Invalid Email Address")
	}

	passwordHash := &vo.PasswordHashed{}
	hashed, err := passwordHash.CreateHash(password)
	if err != nil {
		return nil, err
	}

	return &User{
		Id:    uuid.New(),
		Email: valid_email,
		Name:  name,
		Hash:  hashed,
	}, nil
}
