package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/repository"
)

type PostUseCase struct {
	postRepo repository.PostRepository
}

func NewPostUseCase(postRepo repository.PostRepository) *PostUseCase {
	return &PostUseCase{
		postRepo: postRepo,
	}
}

func (uc *PostUseCase) Create(ctx context.Context, authorID int, text string) (string, error) {
	post := &entity.Post{
		AuthorID:  authorID,
		Text:      text,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := uc.postRepo.Create(ctx, post)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	return id, nil
}

func (uc *PostUseCase) Update(ctx context.Context, postID string, authorID int, text string) error {
	post, err := uc.postRepo.Get(ctx, postID)
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			return ErrPostNotFound
		}
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	if post.AuthorID != authorID {
		return ErrNotPostOwner
	}

	post.Text = text
	post.UpdatedAt = time.Now()

	err = uc.postRepo.Update(ctx, post)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	return nil
}

func (uc *PostUseCase) Delete(ctx context.Context, postID string, authorID int) error {
	post, err := uc.postRepo.Get(ctx, postID)
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			return ErrPostNotFound
		}
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	if post.AuthorID != authorID {
		return ErrNotPostOwner
	}

	err = uc.postRepo.Delete(ctx, postID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	return nil
}

func (uc *PostUseCase) Get(ctx context.Context, postID string) (*entity.Post, error) {
	post, err := uc.postRepo.Get(ctx, postID)
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	return post, nil
}

func (uc *PostUseCase) GetFeed(ctx context.Context, userID, offset, limit int) ([]*entity.Post, error) {
	posts, err := uc.postRepo.GetFeed(ctx, userID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	return posts, nil
}
