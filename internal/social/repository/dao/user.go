package dao

import (
	"time"

	"otus-highload-arh-homework/internal/social/entity"
)

// UserDAO - структура для работы с PostgreSQL
type UserDAO struct {
	ID        int       `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	BirthDate time.Time `db:"birth_date"`
	Gender    string    `db:"gender"`
	Interests []string  `db:"interests"`
	City      string    `db:"city"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// ToEntity преобразует DAO в бизнес-сущность
func ToEntity(dao UserDAO) entity.User {
	return entity.User{
		ID:        dao.ID,
		FirstName: dao.FirstName,
		LastName:  dao.LastName,
		BirthDate: dao.BirthDate,
		Gender:    entity.Gender(dao.Gender),
		Interests: dao.Interests,
		City:      dao.City,
	}
}

// FromEntity преобразует бизнес-сущность в DAO
func FromEntity(user entity.User) UserDAO {
	return UserDAO{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		BirthDate: user.BirthDate,
		Gender:    string(user.Gender),
		Interests: user.Interests,
		City:      user.City,
	}
}
