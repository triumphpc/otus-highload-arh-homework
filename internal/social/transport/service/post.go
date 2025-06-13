package service

import (
	"context"
	"errors"
	"fmt"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/transport/dto"
	postUC "otus-highload-arh-homework/internal/social/usecase/post"
)

type postUseCase interface {
	Create(ctx context.Context, authorID int, text string) (string, error)
	Update(ctx context.Context, postID string, authorID int, text string) error
	Delete(ctx context.Context, postID string, authorID int) error
	Get(ctx context.Context, postID string) (*entity.Post, error)
	GetFeed(ctx context.Context, userID, offset, limit int) ([]*entity.Post, error)
}

type PostService struct {
	postUC postUseCase
}

func NewPostService(
	postUC postUseCase,
) *PostService {
	return &PostService{
		postUC: postUC,
	}
}

// CreatePost создает новый пост
func (s *PostService) CreatePost(ctx context.Context, authorID int, text string) (string, error) {
	// Валидация текста поста
	if len(text) == 0 {
		return "", ErrEmptyPostText
	}
	if len(text) > 1000 {
		return "", fmt.Errorf("%w: post text too long", ErrValidation)
	}

	postID, err := s.postUC.Create(ctx, authorID, text)
	if err != nil {
		if errors.Is(err, postUC.ErrDatabaseOperation) {
			return "", fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
		}
		return "", fmt.Errorf("failed to create post: %w", err)
	}

	return postID, nil
}

// UpdatePost обновляет существующий пост
func (s *PostService) UpdatePost(ctx context.Context, authorID int, postID, text string) error {
	// Валидация
	if len(text) == 0 {
		return ErrEmptyPostText
	}
	if len(text) > 1000 {
		return fmt.Errorf("%w: post text too long", ErrValidation)
	}
	if len(postID) == 0 {
		return ErrInvalidPostID
	}

	err := s.postUC.Update(ctx, postID, authorID, text)
	if err != nil {
		switch {
		case errors.Is(err, postUC.ErrPostNotFound):
			return ErrPostNotFound
		case errors.Is(err, postUC.ErrNotPostOwner):
			return ErrNotPostOwner
		case errors.Is(err, postUC.ErrDatabaseOperation):
			return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
		default:
			return fmt.Errorf("failed to update post: %w", err)
		}
	}

	return nil
}

// DeletePost удаляет пост
func (s *PostService) DeletePost(ctx context.Context, authorID int, postID string) error {
	if len(postID) == 0 {
		return ErrInvalidPostID
	}

	err := s.postUC.Delete(ctx, postID, authorID)
	if err != nil {
		switch {
		case errors.Is(err, postUC.ErrPostNotFound):
			return ErrPostNotFound
		case errors.Is(err, postUC.ErrNotPostOwner):
			return ErrNotPostOwner
		case errors.Is(err, postUC.ErrDatabaseOperation):
			return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
		default:
			return fmt.Errorf("failed to delete post: %w", err)
		}
	}

	return nil
}

// GetPost возвращает пост по ID
func (s *PostService) GetPost(ctx context.Context, postID string) (*dto.PostResponse, error) {
	if len(postID) == 0 {
		return nil, ErrInvalidPostID
	}

	post, err := s.postUC.Get(ctx, postID)
	if err != nil {
		switch {
		case errors.Is(err, postUC.ErrPostNotFound):
			return nil, ErrPostNotFound
		case errors.Is(err, postUC.ErrDatabaseOperation):
			return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
		default:
			return nil, fmt.Errorf("failed to get post: %w", err)
		}
	}

	response := dto.PostResponse{
		ID:        post.ID,
		AuthorID:  post.AuthorID,
		Text:      post.Text,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}

	return &response, nil
}

// GetFeed возвращает ленту постов друзей
func (s *PostService) GetFeed(ctx context.Context, userID, offset, limit int) ([]dto.PostResponse, error) {
	// Валидация параметров пагинации
	if offset < 0 {
		return nil, fmt.Errorf("%w: offset cannot be negative", ErrInvalidPaginationParams)
	}
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("%w: limit must be between 1 and 100", ErrInvalidPaginationParams)
	}

	posts, err := s.postUC.GetFeed(ctx, userID, offset, limit)
	if err != nil {
		if errors.Is(err, postUC.ErrDatabaseOperation) {
			return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
		}
		return nil, fmt.Errorf("failed to get feed: %w", err)
	}

	var response []dto.PostResponse
	for _, post := range posts {
		response = append(response, dto.PostResponse{
			ID:        post.ID,
			AuthorID:  post.AuthorID,
			Text:      post.Text,
			CreatedAt: post.CreatedAt,
			UpdatedAt: post.UpdatedAt,
		})
	}

	return response, nil
}
