package entity

import (
	"fmt"
	"time"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type User struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	BirthDate time.Time
	Gender    Gender
	Interests []string
	City      string

	CreatedAt time.Time
	UpdatedAt time.Time

	PasswordHash string
}

func (u *User) IsAdult() bool {
	return time.Since(u.BirthDate).Hours()/24/365 >= 18
}

func (u *User) IsValid() (bool, error) {
	switch u.Gender {
	case GenderMale, GenderFemale, GenderOther:
		return true, nil
	default:
		return false, fmt.Errorf("invalid gender")
	}
}
