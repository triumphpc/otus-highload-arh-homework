package service

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrAlreadyFriends    = errors.New("users are already friends")
	ErrNotFriends        = errors.New("users are not friends")
	ErrSelfOperation     = errors.New("cannot perform operation on yourself")
	ErrInvalidFriendID   = errors.New("invalid friend id format")
	ErrDatabaseOperation = errors.New("database operation failed")
)

var (
	ErrPostNotFound            = errors.New("post not found")
	ErrNotPostOwner            = errors.New("not post owner")
	ErrEmptyPostText           = errors.New("empty post text")
	ErrInvalidPostID           = errors.New("invalid post ID")
	ErrInvalidPaginationParams = errors.New("invalid pagination parameters")
	ErrValidation              = errors.New("validation failed")
)
