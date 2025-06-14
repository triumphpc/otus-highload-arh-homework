package repository

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

var (
	ErrPostNotFound = errors.New("post not found")
)
