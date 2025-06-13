package postgres

import (
	"context"
	"errors"
	"fmt"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/repository/dao"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPostNotFound = errors.New("post not found")
)

type PostRepository struct {
	pool *pgxpool.Pool
}

// NewPostRepository создает новый экземпляр PostRepository
func NewPostRepository(pool *pgxpool.Pool) *PostRepository {
	return &PostRepository{pool: pool}
}

func (r *PostRepository) Create(ctx context.Context, post *entity.Post) (string, error) {
	const query = `
		INSERT INTO posts (
			author_id, text, created_at, updated_at
		) VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	dao := dao.FromPostEntity(*post)

	var postID string
	err := r.pool.QueryRow(ctx, query,
		dao.AuthorID,
		dao.Text,
		dao.CreatedAt,
		dao.UpdatedAt,
	).Scan(&postID)

	if err != nil {
		if isForeignKeyViolation(err) {
			return "", fmt.Errorf("author not found: %w", err)
		}
		return "", fmt.Errorf("failed to create post: %w", err)
	}

	post.ID = postID
	return postID, nil
}

func (r *PostRepository) Update(ctx context.Context, post *entity.Post) error {
	const query = `
		UPDATE posts
		SET text = $1, updated_at = $2
		WHERE id = $3
	`

	dao := dao.FromPostEntity(*post)

	res, err := r.pool.Exec(ctx, query,
		dao.Text,
		dao.UpdatedAt,
		dao.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrPostNotFound
	}

	return nil
}

func (r *PostRepository) Delete(ctx context.Context, postID string) error {
	const query = `
		DELETE FROM posts
		WHERE id = $1
	`

	res, err := r.pool.Exec(ctx, query, postID)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrPostNotFound
	}

	return nil
}

func (r *PostRepository) Get(ctx context.Context, postID string) (*entity.Post, error) {
	const query = `
		SELECT 
			id, author_id, text, created_at, updated_at
		FROM posts
		WHERE id = $1
	`

	var dao dao.Post
	err := r.pool.QueryRow(ctx, query, postID).Scan(
		&dao.ID,
		&dao.AuthorID,
		&dao.Text,
		&dao.CreatedAt,
		&dao.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	post := dao.ToEntity()
	return &post, nil
}

func (r *PostRepository) GetFeed(ctx context.Context, userID, offset, limit int) ([]*entity.Post, error) {
	const query = `
		SELECT 
			p.id, p.author_id, p.text, p.created_at, p.updated_at
		FROM posts p
		JOIN friends f ON p.author_id = f.friend_id
		WHERE f.user_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query feed: %w", err)
	}
	defer rows.Close()

	var posts []*entity.Post
	for rows.Next() {
		var dao dao.Post
		err := rows.Scan(
			&dao.ID,
			&dao.AuthorID,
			&dao.Text,
			&dao.CreatedAt,
			&dao.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		post := dao.ToEntity()
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return posts, nil
}

// Вспомогательная функция для проверки нарушения внешнего ключа
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
