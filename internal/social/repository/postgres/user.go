package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"otus-highload-arh-homework/internal/social/repository"
	"otus-highload-arh-homework/internal/social/repository/dao"

	"github.com/lib/pq"

	"otus-highload-arh-homework/internal/social/entity"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User, passwordHash string) error {
	const query = `
        INSERT INTO users (
            first_name, last_name, email, birth_date, gender, interests, city, password_hash
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at, updated_at
    `

	dao := dao.FromEntity(*user)

	err := r.pool.QueryRow(ctx, query,
		dao.FirstName,
		dao.LastName,
		dao.Email,
		dao.BirthDate,
		dao.Gender,
		pq.Array(dao.Interests),
		dao.City,
		passwordHash,
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
	const query = `
        SELECT 
            id, 
            first_name, 
            last_name, 
            email, 
            birth_date, 
            gender, 
            interests, 
            city, 
            created_at
        FROM users 
        WHERE id = $1
    `

	var user entity.User
	var interests []sql.NullString

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.BirthDate,
		&user.Gender,
		&interests,
		&user.City,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id %d not found: %w", id, repository.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	// Преобразование интересов (фильтрация NULL значений)
	for _, interest := range interests {
		if interest.Valid {
			user.Interests = append(user.Interests, interest.String)
		}
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	const query = `
        SELECT 
            id, 
            first_name, 
            last_name, 
            email, 
            birth_date, 
            gender, 
            interests, 
            city, 
            created_at,
			password_hash
        FROM users 
        WHERE email = $1
    `

	var user entity.User
	var interests []sql.NullString

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.BirthDate,
		&user.Gender,
		&interests,
		&user.City,
		&user.CreatedAt,
		&user.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with email %s not found: %w", email, repository.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Преобразование интересов (фильтрация NULL значений)
	for _, interest := range interests {
		if interest.Valid {
			user.Interests = append(user.Interests, interest.String)
		}
	}

	return &user, nil
}
