package repository

import (
	"NTEC_task_RESTAPI/internal/model"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (int64, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
}

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) (int64, error)
	GetAll(ctx context.Context, filter model.TaskFilter) ([]model.Task, error)
	GetByID(ctx context.Context, id int64) (*model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id int64) error
	MarkOverdueTasks(ctx context.Context) error
}
