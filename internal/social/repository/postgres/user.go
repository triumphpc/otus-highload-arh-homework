package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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
            id, first_name, last_name, email, 
            birth_date, gender, interests, city,
            created_at, updated_at
        FROM users 
        WHERE id = $1
    `

	var user entity.User
	var interests []string
	var birthDate time.Time

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&birthDate,
		&user.Gender,
		pq.Array(&interests),
		&user.City,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", repository.ErrUserNotFound, err)
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	user.BirthDate = birthDate
	user.Interests = interests

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	const query = `
        UPDATE users 
        SET 
            first_name = $1,
            last_name = $2,
            email = $3,
            birth_date = $4,
            gender = $5,
            interests = $6,
            city = $7,
            updated_at = NOW()
        WHERE id = $8
        RETURNING updated_at
    `

	dao := dao.FromEntity(*user)
	var updatedAt time.Time

	err := r.pool.QueryRow(ctx, query,
		dao.FirstName,
		dao.LastName,
		dao.Email,
		dao.BirthDate,
		dao.Gender,
		pq.Array(dao.Interests),
		dao.City,
		dao.ID,
	).Scan(&updatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrUserNotFound
		}
		if isDuplicateKeyError(err) {
			return repository.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	user.UpdatedAt = updatedAt
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	const query = `DELETE FROM users WHERE id = $1`

	cmd, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
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
            created_at
        FROM users 
        WHERE email = $1
    `

	var user entity.User
	var interests []sql.NullString
	var birthDateStr string

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&birthDateStr,
		&user.Gender,
		pq.Array(&interests),
		&user.City,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with email %s not found: %w", email, repository.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Преобразование даты рождения
	user.BirthDate, err = time.Parse(time.RFC3339, birthDateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse birth date: %w", err)
	}

	// Преобразование интересов (фильтрация NULL значений)
	for _, interest := range interests {
		if interest.Valid {
			user.Interests = append(user.Interests, interest.String)
		}
	}

	return &user, nil
}
