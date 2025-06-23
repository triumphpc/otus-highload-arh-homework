package repository

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrNoMessagesFound   = errors.New("no messages found")
)
