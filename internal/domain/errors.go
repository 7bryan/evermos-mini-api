package domain

import "errors"

var (
	ErrNotFound       = errors.New("record not found")
	ErrDuplicatePhone = errors.New("phone number already exists")
	ErrDuplicateEmail = errors.New("email already exists")
)
