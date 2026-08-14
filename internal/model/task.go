package model

import "time"

type Task struct {
	ID            int64
	Title         string
	Description   string
	Deadline      time.Time
	Status        string
	CreatorID     int64
	ResponsibleID *int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TaskFilter struct {
	Status        *string
	ResponsibleID *int64
	Limit         int64
	Offset        int64
	DeadlineFrom  *time.Time
	DeadlineTo    *time.Time
}
