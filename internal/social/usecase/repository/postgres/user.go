package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/usecase/repository"
	"otus-highload-arh-homework/internal/social/usecase/repository/dto"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	const query = `
		INSERT INTO users (
			first_name, last_name, birth_date, gender, interests, city
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	dao := dto.FromEntity(*user)

	err := r.pool.QueryRow(ctx, query,
		dao.FirstName,
		dao.LastName,
		dao.BirthDate,
		dao.Gender,
		dao.Interests,
		dao.City,
	).Scan(&dao.ID, &dao.CreatedAt, &dao.UpdatedAt)

	if err != nil {
		if isDuplicateKeyError(err) {
			return repository.ErrUserAlreadyExists
		}
		return err
	}

	// Обновляем entity из DAO
	user.ID = dao.ID
	user.CreatedAt = dao.CreatedAt
	user.UpdatedAt = dao.UpdatedAt

	return nil
}

// Вспомогательная функция для проверки ошибки дубликата
func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*entity.User, error) {
	//TODO implement me
	panic("implement me")
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	//TODO implement me
	panic("implement me")
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	//TODO implement me
	panic("implement me")
}
