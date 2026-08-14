package background

import (
	"context"
	"log"
	"time"

	"NTEC_task_RESTAPI/internal/repository"
)

type TaskWorker struct {
	repo     repository.TaskRepository
	interval time.Duration
}

func NewTaskWorker(repo repository.TaskRepository, interval time.Duration) *TaskWorker {
	return &TaskWorker{
		repo:     repo,
		interval: interval,
	}
}

func (w *TaskWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("Background worker started. Interval: %v", w.interval)

	for {
		select {
		case <-ticker.C:
			taskCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			err := w.repo.MarkOverdueTasks(taskCtx)

			cancel()

			if err != nil {
				log.Printf("Worker error marking overdue tasks: %v", err)
			} else {
				log.Println("Worker successfully checked for overdue tasks")
			}

		case <-ctx.Done():
			log.Println("Background worker is stopping...")
			return
		}
	}
}
