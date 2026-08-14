package handler

import (
	"NTEC_task_RESTAPI/internal/model"
	"NTEC_task_RESTAPI/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type TaskHandler struct {
	taskService service.TaskService
}

type CreateTaskRequest struct {
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Deadline      time.Time `json:"deadline"`
	ResponsibleID *int64    `json:"responsible_id"`
}
type CreateTaskResponse struct {
	ID int64 `json:"id"`
}

type UpdateTaskRequest struct {
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	Deadline      time.Time `json:"deadline"`
	ResponsibleID *int64    `json:"responsible_id"`
}

type TaskResponse struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	Deadline      time.Time `json:"deadline"`
	ResponsibleID *int64    `json:"responsible_id"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	creatorID := r.Context().Value(UserIDKey).(int64)

	task := model.Task{
		Title:         req.Title,
		Description:   req.Description,
		Deadline:      req.Deadline,
		ResponsibleID: req.ResponsibleID,
		CreatorID:     creatorID,
	}

	id, err := h.taskService.Create(r.Context(), &task)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateTaskResponse{ID: id})
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	filter := model.TaskFilter{}

	statusStr := r.URL.Query().Get("status")
	if statusStr != "" {
		filter.Status = &statusStr
	}

	respStr := r.URL.Query().Get("responsible_id")
	if respStr != "" {
		respID, err := strconv.ParseInt(respStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid responsible_id parameter", http.StatusBadRequest)
			return
		}
		filter.ResponsibleID = &respID
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid limit parameter", http.StatusBadRequest)
			return
		}
		filter.Limit = limit
	}

	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		offset, err := strconv.ParseInt(offsetStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid offset parameter", http.StatusBadRequest)
			return
		}
		filter.Offset = offset
	}

	tasks, err := h.taskService.GetAll(r.Context(), filter)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	task, err := h.taskService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(int64)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	task := model.Task{
		ID:            id,
		Title:         req.Title,
		Description:   req.Description,
		Status:        req.Status,
		Deadline:      req.Deadline,
		ResponsibleID: req.ResponsibleID,
	}

	if err := h.taskService.Update(r.Context(), userID, &task); err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			http.Error(w, "access denied: only creator can modify this task", http.StatusForbidden)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(TaskResponse{
		ID:            task.ID,
		Title:         task.Title,
		Description:   task.Description,
		Status:        string(task.Status),
		Deadline:      task.Deadline,
		ResponsibleID: task.ResponsibleID,
	})
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(int64)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	if err := h.taskService.Delete(r.Context(), userID, id); err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			http.Error(w, "access denied: only creator can delete this task", http.StatusForbidden)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
