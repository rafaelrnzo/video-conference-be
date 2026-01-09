package service

import (
	"context"
	"errors"
	"time"

	dUser "video-conference-be/internal/domain/user"
	"video-conference-be/pkg/utility"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, username, password string) (*dUser.User, error)
	Login(ctx context.Context, username, password string) (string, *dUser.User, error)
}

type authService struct {
	userRepo dUser.Repository
}

func NewAuthService(userRepo dUser.Repository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Register(ctx context.Context, username, password string) (*dUser.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &dUser.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         dUser.RoleUser,
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *authService) Login(ctx context.Context, username, password string) (string, *dUser.User, error) {
	u, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}
	token, err := utility.GenerateJWT(u.Username, u.Role, 24*time.Hour)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}
