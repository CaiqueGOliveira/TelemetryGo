package repository

import "errors"

var (
	ErrDuplicateEmail = errors.New("user with this email already exists")
	ErrNotFound       = errors.New("not found")
)
