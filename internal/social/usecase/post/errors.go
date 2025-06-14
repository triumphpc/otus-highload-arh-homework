package usecase

import (
	"errors"
)

var (
	ErrNotPostOwner      = errors.New("user is not the owner of the post")
	ErrPostNotFound      = errors.New("post not found")
	ErrDatabaseOperation = errors.New("database operation failed")
)
