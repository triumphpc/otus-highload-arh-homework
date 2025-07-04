package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

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

type cacheWarmer interface {
	WarmForNewPost(ctx context.Context, authorID int) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string, dest any) error
}

type postFeeder interface {
	Publish(ctx context.Context, key string, value any) error
}

type PostService struct {
	postUC      postUseCase
	friendUC    friendUseCase
	cacheWarmer cacheWarmer
	postFeeder  postFeeder
}

func NewPostService(
	postUC postUseCase,
	friendUC friendUseCase,
	warmer cacheWarmer,
	feeder postFeeder,
) *PostService {
	return &PostService{
		postUC:      postUC,
		friendUC:    friendUC,
		cacheWarmer: warmer,
		postFeeder:  feeder,
	}
}

// CreatePost создает новый пост
func (s *PostService) CreatePost(ctx context.Context, authorID int, text string) (string, error) {
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

	// Обновление ленты ws
	go func() {
		event := dto.FeedUpdateEvent{
			PostID:    postID,
			AuthorID:  authorID,
			Action:    "create",
			Text:      text,
			Timestamp: time.Now().Unix(),
		}

		// Используем authorID как ключ для партиционирования
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := s.postFeeder.Publish(ctx, strconv.Itoa(authorID), event); err != nil {
			log.Printf("failed to publish feed update event: %v", err)
		}
	}()

	// Прогрев кеша
	go s.warmCache(authorID)

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

	// Отправка в ленту ws
	go func() {
		event := dto.FeedUpdateEvent{
			PostID:    postID,
			AuthorID:  authorID,
			Action:    "update",
			Text:      text,
			Timestamp: time.Now().Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := s.postFeeder.Publish(ctx, strconv.Itoa(authorID), event); err != nil {
			log.Printf("failed to publish feed update event: %v", err)
		}
	}()

	// Прогрев кеша
	go s.warmCache(authorID)

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

	// Отправка в ленту по ws
	go func() {
		event := dto.FeedUpdateEvent{
			PostID:    postID,
			AuthorID:  authorID,
			Action:    "delete",
			Timestamp: time.Now().Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := s.postFeeder.Publish(ctx, strconv.Itoa(authorID), event); err != nil {
			log.Printf("failed to publish feed update event: %v", err)
		}
	}()

	// Прогрев кеша
	go s.warmCache(authorID)

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

	// Пробуем получить из кэша
	cachedPosts, err := s.getFromCache(ctx, userID, offset, limit)
	if err == nil {
		return cachedPosts, nil
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

	go s.warmCache(userID)

	return response, nil
}

func (s *PostService) warmCache(authorID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Warm cache start from ", authorID)

	if err := s.cacheWarmer.WarmForNewPost(ctx, authorID); err != nil {
		log.Printf("cache warm failed: %v", err)
	}
}

// getFromCache получает посты из кэша
func (s *PostService) getFromCache(ctx context.Context, userID, offset, limit int) ([]dto.PostResponse, error) {
	key := s.feedCacheKey(userID)

	var cached []dto.PostResponse
	if err := s.cacheWarmer.Get(ctx, key, &cached); err != nil {
		return nil, ErrCacheMiss
	}

	// Применяем пагинацию
	if offset >= len(cached) {
		return []dto.PostResponse{}, nil
	}

	end := offset + limit
	if end > len(cached) {
		end = len(cached)
	}

	return cached[offset:end], nil
}

// cacheFeed сохраняет фид в кэш
func (s *PostService) cacheFeed(userID int, posts []dto.PostResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := s.feedCacheKey(userID)
	return s.cacheWarmer.Set(ctx, key, posts, time.Hour*24)
}

// feedCacheKey генерирует ключ для кэша фида
func (s *PostService) feedCacheKey(userID int) string {
	return fmt.Sprintf("user:%d:feed", userID)
}

// PreloadUserFriendsFeeds предзагружает кэш фидов друзей пользователя
func (s *PostService) PreloadUserFriendsFeeds(ctx context.Context, userID int) error {
	friendIDs, err := s.friendUC.GetFriendsIDs(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get friends ids: %w", err)
	}

	for _, friendID := range friendIDs {
		log.Printf("Start warm for %d\n", friendID)
		// Получаем 1000 последних постов
		posts, err := s.postUC.GetFeed(ctx, friendID, 0, 1000)
		if err != nil {
			return fmt.Errorf("failed to get feed for preload: %w", err)
		}

		// Конвертируем в DTO
		response := make([]dto.PostResponse, 0, len(posts))
		for _, post := range posts {
			response = append(response, dto.PostResponse{
				ID:        post.ID,
				AuthorID:  post.AuthorID,
				Text:      post.Text,
				CreatedAt: post.CreatedAt,
				UpdatedAt: post.UpdatedAt,
			})
		}

		// Сохраняем в кэш
		if err := s.cacheFeed(friendID, response); err != nil {
			return fmt.Errorf("failed to cache feed: %w", err)
		}

	}

	return nil
}
