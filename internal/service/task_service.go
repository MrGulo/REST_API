package service

import (
	"NTEC_task_RESTAPI/internal/model"
	"NTEC_task_RESTAPI/internal/repository"
	"context"
	"errors"
	"fmt"
)

type taskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrForbidden    = errors.New("user is not allowed to modify this task")
)

func (s *taskService) Create(ctx context.Context, task *model.Task) (int64, error) {
	if task.ResponsibleID == nil {
		task.ResponsibleID = &task.CreatorID
	}

	id, err := s.repo.Create(ctx, task)
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %w", err)
	}

	return id, nil
}

func (s *taskService) Update(ctx context.Context, userID int64, task *model.Task) error {
	t, err := s.repo.GetByID(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("failed to get task for update: %w", err)
	}

	if t == nil {
		return ErrTaskNotFound
	}

	if t.CreatorID != userID {
		return ErrForbidden
	}

	err = s.repo.Update(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (s *taskService) Delete(ctx context.Context, userID, taskID int64) error {
	t, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task for deletion: %w", err)
	}

	if t == nil {
		return ErrTaskNotFound
	}

	if t.CreatorID != userID {
		return ErrForbidden
	}

	err = s.repo.Delete(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

func (s *taskService) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *taskService) GetAll(ctx context.Context, filter model.TaskFilter) ([]model.Task, error) {
	if filter.DeadlineFrom != nil && filter.DeadlineTo != nil {
		if filter.DeadlineFrom.After(*filter.DeadlineTo) {
			return nil, errors.New("invalid deadline range: 'from' must be before 'to'")
		}
	}

	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return s.repo.GetAll(ctx, filter)
}
