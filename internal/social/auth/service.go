package auth

import (
	"context"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/usecase/repository"
)

type AuthService struct {
	userRepo   repository.UserRepository
	hasher     PasswordHasher
	jwtService *JWTGenerator
}

func NewAuthService(repo repository.UserRepository, hasher PasswordHasher, jwt *JWTGenerator) *AuthService {
	return &AuthService{
		userRepo:   repo,
		hasher:     hasher,
		jwtService: jwt,
	}
}

func (s *AuthService) Register(ctx context.Context, user *entity.User) error {
	// ... (как в предыдущем примере)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if !s.hasher.Check(password, user.PasswordHash) {
		return "", ErrInvalidCredentials
	}

	token, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}
