package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"NTEC_task_RESTAPI/internal/model"

	"golang.org/x/crypto/bcrypt"
)

var errUserAlreadyExists = errors.New("user already exists")

type mockUserRepository struct {
	mu     sync.RWMutex
	users  map[string]*model.User
	lastID int64
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*model.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *model.User) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.Username]; exists {
		return 0, errUserAlreadyExists
	}

	m.lastID++
	user.ID = m.lastID

	m.users[user.Username] = user

	return user.ID, nil
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[username]
	if !exists {
		return nil, errors.New("user not found in mock db")
	}

	return user, nil
}

func TestUserService_Register(t *testing.T) {
	mockRepo := newMockUserRepository()
	jwtSecret := "test_secret"
	userService := NewUserService(mockRepo, jwtSecret)

	tests := []struct {
		name     string
		username string
		password string
		wantErr  error
	}{
		{
			name:     "Успешная регистрация",
			username: "ivan_dev",
			password: "password123",
			wantErr:  nil,
		},
		{
			name:     "Регистрация с существующим username",
			username: "ivan_dev",
			password: "another_password",
			wantErr:  errUserAlreadyExists,
		},
		{
			name:     "Пустой пароль",
			username: "petr_dev",
			password: "",
			wantErr:  ErrEmptyCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := userService.Register(context.Background(), tt.username, tt.password)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Register() expected error %v, but got nil", tt.wantErr)
					return
				}

				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Register() expected error %v, but got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Register() unexpected error: %v", err)
				return
			}

			if id <= 0 {
				t.Errorf("Register() got invalid id = %d, expected > 0", id)
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	mockRepo := newMockUserRepository()
	jwtSecret := "test_secret"
	userService := NewUserService(mockRepo, jwtSecret)

	testPassword := "correct_password"
	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password for test setup: %v", err)
	}

	mockRepo.users["existing_user"] = &model.User{
		ID:           1,
		Username:     "existing_user",
		PasswordHash: string(hash),
	}

	tests := []struct {
		name     string
		username string
		password string
		wantErr  error
	}{
		{
			name:     "Успешный логин",
			username: "existing_user",
			password: "correct_password",
			wantErr:  nil,
		},
		{
			name:     "Неверный пароль",
			username: "existing_user",
			password: "wrong_password",
			wantErr:  ErrInvalidCredentials,
		},
		{
			name:     "Несуществующий пользователь",
			username: "ghost_user",
			password: "any_password",
			wantErr:  ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := userService.Login(context.Background(), tt.username, tt.password)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Login() expected error %v, but got nil", tt.wantErr)
					return
				}

				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Login() expected error %v, but got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Login() unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Errorf("Login() expected a valid user object, but got nil")
			} else if user.Username != tt.username {
				t.Errorf("Login() expected user %q, but got %q", tt.username, user.Username)
			}
		})
	}
}
