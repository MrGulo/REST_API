package service

import (
	"context"

	"NTEC_task_RESTAPI/internal/model"
)

type UserService interface {
	Register(ctx context.Context, username, password string) (int64, error)
	Login(ctx context.Context, username, password string) (*model.User, error)
	GenerateToken(user *model.User) (string, error)
}

type TaskService interface {
	Create(ctx context.Context, task *model.Task) (int64, error)
	Update(ctx context.Context, userID int64, task *model.Task) error
	Delete(ctx context.Context, userID, taskID int64) error
	GetByID(ctx context.Context, id int64) (*model.Task, error)
	GetAll(ctx context.Context, filter model.TaskFilter) ([]model.Task, error)
}
