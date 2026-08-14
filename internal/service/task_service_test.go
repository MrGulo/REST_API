package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"NTEC_task_RESTAPI/internal/model"
)

type mockTaskRepository struct {
	mu     sync.RWMutex
	tasks  map[int64]*model.Task
	lastID int64
}

func newMockTaskRepository() *mockTaskRepository {
	return &mockTaskRepository{
		tasks: make(map[int64]*model.Task),
	}
}

func (m *mockTaskRepository) Create(ctx context.Context, task *model.Task) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastID++
	task.ID = m.lastID

	m.tasks[task.ID] = task

	return task.ID, nil
}

func (m *mockTaskRepository) GetAll(ctx context.Context, filter model.TaskFilter) ([]model.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []model.Task
	for _, task := range m.tasks {
		result = append(result, *task)
	}

	return result, nil
}

func (m *mockTaskRepository) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

func (m *mockTaskRepository) Update(ctx context.Context, task *model.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[task.ID]; !exists {
		return ErrTaskNotFound
	}

	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) Delete(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[id]; !exists {
		return ErrTaskNotFound
	}

	delete(m.tasks, id)
	return nil
}

func (m *mockTaskRepository) MarkOverdueTasks(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

func TestTaskService_GetByID(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	mockRepo.tasks[1] = &model.Task{
		ID:    1,
		Title: "Существующая задача",
	}

	tests := []struct {
		name    string
		taskID  int64
		wantErr error
	}{
		{
			name:    "Успешное получение задачи",
			taskID:  1,
			wantErr: nil,
		},
		{
			name:    "Ошибка: задача не найдена",
			taskID:  999,
			wantErr: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := taskService.GetByID(context.Background(), tt.taskID)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("GetByID() expected error %v, but got nil", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetByID() expected error %v, but got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("GetByID() unexpected error: %v", err)
				return
			}

			if task == nil {
				t.Errorf("GetByID() expected task, got nil")
			} else if task.ID != tt.taskID {
				t.Errorf("GetByID() expected task ID %d, got %d", tt.taskID, task.ID)
			}
		})
	}
}

func TestTaskService_GetAll(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	mockRepo.tasks[1] = &model.Task{ID: 1, Title: "Задача 1"}
	mockRepo.tasks[2] = &model.Task{ID: 2, Title: "Задача 2"}

	tests := []struct {
		name      string
		filter    model.TaskFilter
		wantErr   error
		wantCount int
	}{
		{
			name:      "Успешное получение всех задач без фильтра",
			filter:    model.TaskFilter{},
			wantErr:   nil,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := taskService.GetAll(context.Background(), tt.filter)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("GetAll() expected error %v, but got nil", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetAll() expected error %v, but got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("GetAll() unexpected error: %v", err)
				return
			}

			if len(tasks) != tt.wantCount {
				t.Errorf("GetAll() expected %d tasks, got %d", tt.wantCount, len(tasks))
			}
		})
	}
}

func TestTaskService_Create(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	creatorID := int64(1)
	respID := int64(2)

	tests := []struct {
		name              string
		task              *model.Task
		wantErr           error
		wantResponsibleID int64
	}{
		{
			name: "Успешное создание с указанным ResponsibleID",
			task: &model.Task{
				Title:         "Настроить CI/CD",
				CreatorID:     creatorID,
				ResponsibleID: &respID,
			},
			wantErr:           nil,
			wantResponsibleID: respID,
		},
		{
			name: "Успешное создание без ResponsibleID (подстановка CreatorID)",
			task: &model.Task{
				Title:     "Написать документацию",
				CreatorID: creatorID,
			},
			wantErr:           nil,
			wantResponsibleID: creatorID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := taskService.Create(context.Background(), tt.task)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Create() expected error %v, but got nil", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Create() expected error %v, but got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Create() unexpected error: %v", err)
				return
			}

			if tt.task.ResponsibleID == nil {
				t.Errorf("Expected ResponsibleID to be set, but got nil")
			} else if *tt.task.ResponsibleID != tt.wantResponsibleID {
				t.Errorf("Create() got ResponsibleID = %v, want %v", *tt.task.ResponsibleID, tt.wantResponsibleID)
			}
		})
	}
}

func TestTaskService_Update(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	mockRepo.tasks[1] = &model.Task{
		ID:        1,
		Title:     "Старая задача",
		CreatorID: 10,
	}

	tests := []struct {
		name    string
		userID  int64
		task    *model.Task
		wantErr error
	}{
		{
			name:   "Успешное обновление (userID совпадает с CreatorID)",
			userID: 10,
			task: &model.Task{
				ID:    1,
				Title: "Новое название задачи",
			},
			wantErr: nil,
		},
		{
			name:   "Ошибка доступа: обновляет не автор",
			userID: 99,
			task: &model.Task{
				ID:    1,
				Title: "Попытка взлома",
			},
			wantErr: ErrForbidden,
		},
		{
			name:   "Ошибка: задача не найдена",
			userID: 10,
			task: &model.Task{
				ID:    999,
				Title: "Задача-призрак",
			},
			wantErr: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := taskService.Update(context.Background(), tt.userID, tt.task)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Update() expected error %v, but got nil", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Update() expected error %v, but got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Update() unexpected error: %v", err)
			}
		})
	}
}

func TestTaskService_Delete(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	mockRepo.tasks[1] = &model.Task{
		ID:        1,
		Title:     "Устаревшая задача",
		CreatorID: 10,
	}

	tests := []struct {
		name    string
		userID  int64
		taskID  int64
		wantErr error
	}{
		{
			name:    "Ошибка доступа: удаляет не автор",
			userID:  99,
			taskID:  1,
			wantErr: ErrForbidden,
		},
		{
			name:    "Успешное удаление автором",
			userID:  10,
			taskID:  1,
			wantErr: nil,
		},
		{
			name:    "Ошибка: попытка удалить несуществующую задачу",
			userID:  10,
			taskID:  999,
			wantErr: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := taskService.Delete(context.Background(), tt.userID, tt.taskID)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Delete() expected error %v, but got nil", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Delete() expected error %v, but got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Delete() unexpected error: %v", err)
			}
		})
	}
}
