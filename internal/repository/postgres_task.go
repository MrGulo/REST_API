package repository

import (
	"NTEC_task_RESTAPI/internal/model"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *PostgresTaskRepository {
	return &PostgresTaskRepository{db: db}
}

func (r *PostgresTaskRepository) Create(ctx context.Context, task *model.Task) (int64, error) {
	query := `
		INSERT INTO tasks (title, description, deadline, status, creator_id, responsible_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	var taskID int64

	err := r.db.QueryRow(ctx, query, task.Title, task.Description, task.Deadline, task.Status, task.CreatorID, task.ResponsibleID).Scan(&taskID)
	if err != nil {
		return 0, fmt.Errorf("error creating task: %w", err)
	}

	return taskID, nil
}

func (r *PostgresTaskRepository) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	query := `
			  SELECT id, title, description, deadline, status, creator_id, responsible_id, created_at, updated_at 
			  FROM tasks 
			  WHERE id = $1`

	var task model.Task

	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Deadline,
		&task.Status,
		&task.CreatorID,
		&task.ResponsibleID,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting task: %w", err)
	}
	return &task, nil
}

func (r *PostgresTaskRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM tasks WHERE id = $1"

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting task: %w", err)
	}

	return nil
}

func (r *PostgresTaskRepository) Update(ctx context.Context, task *model.Task) error {
	query := `UPDATE tasks 
			  SET title = $1, description = $2, deadline = $3, status = $4, responsible_id = $5, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = $6`

	_, err := r.db.Exec(ctx, query, task.Title, task.Description, task.Deadline, task.Status, task.ResponsibleID, task.ID)
	if err != nil {
		return fmt.Errorf("error updating task: %w", err)
	}
	return nil
}

func (r *PostgresTaskRepository) GetAll(ctx context.Context, filter model.TaskFilter) ([]model.Task, error) {
	baseQuery := `
		SELECT id, title, description, deadline, status, creator_id, responsible_id, created_at, updated_at 
		FROM tasks 
		WHERE 1=1
	`

	var conditions []string
	var args []interface{}
	argID := 1

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argID))
		args = append(args, *filter.Status)
		argID++
	}

	if filter.ResponsibleID != nil {
		conditions = append(conditions, fmt.Sprintf("responsible_id = $%d", argID))
		args = append(args, *filter.ResponsibleID)
		argID++
	}

	if filter.DeadlineFrom != nil {
		conditions = append(conditions, fmt.Sprintf("deadline >= $%d", argID))
		args = append(args, *filter.DeadlineFrom)
		argID++
	}

	if filter.DeadlineTo != nil {
		conditions = append(conditions, fmt.Sprintf("deadline <= $%d", argID))
		args = append(args, *filter.DeadlineTo)
		argID++
	}

	finalQuery := baseQuery
	if len(conditions) > 0 {
		finalQuery += " AND " + strings.Join(conditions, " AND ")
	}

	finalQuery += " ORDER BY created_at DESC"

	finalQuery += fmt.Sprintf(" LIMIT $%d", argID)
	args = append(args, filter.Limit)
	argID++

	finalQuery += fmt.Sprintf(" OFFSET $%d", argID)
	args = append(args, filter.Offset)

	rows, err := r.db.Query(ctx, finalQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Description,
			&t.Deadline,
			&t.Status,
			&t.CreatorID,
			&t.ResponsibleID,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}
		tasks = append(tasks, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return tasks, nil
}

func (r *PostgresTaskRepository) MarkOverdueTasks(ctx context.Context) error {
	query := `
		UPDATE tasks 
		SET status = 'expired' 
		WHERE deadline < $1 AND status NOT IN ('completed', 'expired')
	`

	tag, err := r.db.Exec(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to mark overdue tasks: %w", err)
	}

	if tag.RowsAffected() > 0 {
		_ = tag.RowsAffected()
	}

	return nil
}
