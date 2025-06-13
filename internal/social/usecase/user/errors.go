package user

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrSelfOperation  = errors.New("cannot perform operation on yourself")
	ErrAlreadyFriends = errors.New("users are already friends")
	ErrNotFriends     = errors.New("users are not friends")
)
