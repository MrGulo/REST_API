package repository

import (
	"NTEC_task_RESTAPI/internal/model"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *model.User) (int64, error) {
	query := `
			INSERT INTO users (username, password_hash) 
			VALUES ($1, $2) 
			RETURNING id`

	var id int64

	err := r.db.QueryRow(ctx, query, user.Username, user.PasswordHash).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	return id, nil
}

func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
			SELECT id, username, password_hash, created_at, updated_at 
			FROM users 
			WHERE username = $1`

	var u model.User
	err := r.db.QueryRow(ctx, query, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &u, nil
}
