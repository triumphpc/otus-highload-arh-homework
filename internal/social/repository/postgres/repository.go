package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"otus-highload-arh-homework/internal/social/repository"
	"otus-highload-arh-homework/internal/social/repository/dao"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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

// Search - поиск по имени и фамилии
func (r *UserRepository) Search(ctx context.Context, firstName, lastName string) ([]*entity.User, error) {
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
        WHERE first_name ILIKE $1 || '%' 
        AND last_name ILIKE $2 || '%'
        LIMIT 100
    `

	rows, err := r.pool.Query(ctx, query, firstName, lastName)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		var interests []sql.NullString

		err := rows.Scan(
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
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		// Преобразование интересов
		for _, interest := range interests {
			if interest.Valid {
				user.Interests = append(user.Interests, interest.String)
			}
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if len(users) == 0 {
		return nil, repository.ErrUserNotFound
	}

	return users, nil
}

// AddFriend добавляет друга для пользователя
func (r *UserRepository) AddFriend(ctx context.Context, userID, friendID int) error {
	const query = `
		INSERT INTO friends (user_id, friend_id)
		VALUES ($1, $2), ($2, $1)
		ON CONFLICT (user_id, friend_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, userID, friendID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503": // foreign_key_violation
				return fmt.Errorf("one of the users doesn't exist: %w", err)
			case "23505": // unique_violation
				return nil // дружба уже существует, это не ошибка
			}
		}
		return fmt.Errorf("failed to add friend: %w", err)
	}

	return nil
}

// RemoveFriend удаляет друга у пользователя (взаимное удаление)
func (r *UserRepository) RemoveFriend(ctx context.Context, userID, friendID int) error {
	const query = `
		DELETE FROM friends
		WHERE (user_id = $1 AND friend_id = $2)
		OR (user_id = $2 AND friend_id = $1)
	`

	result, err := r.pool.Exec(ctx, query, userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to remove friend: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("friendship not found: %w", sql.ErrNoRows)
	}

	return nil
}

// CheckFriendship проверяет существование дружбы между пользователями
func (r *UserRepository) CheckFriendship(ctx context.Context, userID, friendID int) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1 FROM friends
			WHERE (user_id = $1 AND friend_id = $2)
			OR (user_id = $2 AND friend_id = $1)
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, friendID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check friendship: %w", err)
	}

	return exists, nil
}

// GetFriendsIDs возвращает список друзей пользователя
func (r *UserRepository) GetFriendsIDs(ctx context.Context, userID int) ([]int, error) {
	const query = `
        SELECT friend_id FROM friends WHERE user_id = $1
    `

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query friends IDs: %w", err)
	}
	defer rows.Close()

	var friendsIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan friend ID: %w", err)
		}
		friendsIDs = append(friendsIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return friendsIDs, nil
}

// GetOrCreateDialog получает или создает диалог между пользователями
func (r *UserRepository) getOrCreateDialog(ctx context.Context, user1ID, user2ID int64) (int64, error) {
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}

	// Сначала попробуем найти существующий диалог
	var dialogID int64
	err := r.pool.QueryRow(ctx,
		"SELECT dialog_id FROM dialogs WHERE user1_id = $1 AND user2_id = $2",
		user1ID, user2ID).Scan(&dialogID)

	if err == nil {
		return dialogID, nil // Диалог найден
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("failed to find dialog: %w", err)
	}

	// Диалога нет - создаем новый
	err = r.pool.QueryRow(ctx,
		"INSERT INTO dialogs (user1_id, user2_id) VALUES ($1, $2) RETURNING dialog_id",
		user1ID, user2ID).Scan(&dialogID)

	if err != nil {
		return 0, fmt.Errorf("failed to create dialog: %w", err)
	}

	return dialogID, nil
}

// StoreDialogMessage сохраняет сообщение в диалоге
func (r *UserRepository) StoreDialogMessage(ctx context.Context, senderID, recipientID int64, content string) (int64, error) {

	dialogID, err := r.getOrCreateDialog(ctx, senderID, recipientID)
	if err != nil {
		return 0, fmt.Errorf("failed to get or create dialog: %w", err)
	}

	const query = `
		INSERT INTO messages (dialog_id, sender_id, recipient_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING message_id
	`

	var messageID int64
	err = r.pool.QueryRow(ctx, query, dialogID, senderID, recipientID, content).Scan(&messageID)
	if err != nil {
		return 0, fmt.Errorf("failed to store message: %w", err)
	}

	return messageID, nil
}

// GetDialogMessages возвращает все сообщения между двумя пользователями, отсортированные по времени
func (r *UserRepository) GetDialogMessages(ctx context.Context, senderID, recipientID int64) ([]*entity.DialogMessage, error) {
	const query = `
        SELECT 
            message_id::text,
            sender_id::text,
            recipient_id::text,
            content as text,
            created_at as sent_at,
            read_at IS NOT NULL as is_read
        FROM messages
        WHERE (sender_id = $1 AND recipient_id = $2)
           OR (sender_id = $2 AND recipient_id = $1)
        ORDER BY created_at ASC
    `

	rows, err := r.pool.Query(ctx, query, senderID, recipientID)
	if err != nil {
		return nil, fmt.Errorf("failed to query dialog messages: %w", err)
	}
	defer rows.Close()

	var messages []*entity.DialogMessage
	for rows.Next() {
		var msg entity.DialogMessage
		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Text,
			&msg.SentAt,
			&msg.IsRead,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if len(messages) == 0 {
		return nil, repository.ErrNoMessagesFound
	}

	return messages, nil
}
