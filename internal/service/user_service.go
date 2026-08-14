package service

import (
	"NTEC_task_RESTAPI/internal/model"
	"NTEC_task_RESTAPI/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo      repository.UserRepository
	jwtSecret []byte
}

func NewUserService(repo repository.UserRepository, jwtSecret string) UserService {
	return &userService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

var (
	ErrEmptyCredentials   = errors.New("username and password cannot be empty")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

func (s *userService) Register(ctx context.Context, username, password string) (int64, error) {
	if username == "" || password == "" {
		return 0, ErrEmptyCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("error hashing password: %w", err)
	}

	user := &model.User{
		Username:     username,
		PasswordHash: string(hash),
	}

	id, err := s.repo.Create(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return id, nil
}

func (s *userService) Login(ctx context.Context, username, password string) (*model.User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *userService) GenerateToken(user *model.User) (string, error) {
	timeLive := time.Now().Add(time.Hour * 24).Unix()

	claims := jwt.MapClaims{"user_id": user.ID, "exp": timeLive}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return tokenString, nil
}
