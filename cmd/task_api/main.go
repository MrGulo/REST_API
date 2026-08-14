package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"NTEC_task_RESTAPI/internal/background"
	"NTEC_task_RESTAPI/internal/handler"
	"NTEC_task_RESTAPI/internal/repository"
	"NTEC_task_RESTAPI/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error reading it, relying on system environment variables")
	}

	dbURL := getEnv("DATABASE_URL", "postgres://admin:adminpassword@localhost:5432/task_db?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "super_secret_key")
	port := getEnv("SERVER_PORT", ":8080")

	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	userRepo := repository.NewUserRepository(dbPool)
	taskRepo := repository.NewTaskRepository(dbPool)

	userService := service.NewUserService(userRepo, jwtSecret)
	taskService := service.NewTaskService(taskRepo)

	userHandler := handler.NewUserHandler(userService)
	taskHandler := handler.NewTaskHandler(taskService)

	router := handler.NewRouter(userHandler, taskHandler, jwtSecret)

	workerIntervalStr := getEnv("WORKER_INTERVAL", "1m")
	workerInterval, err := time.ParseDuration(workerIntervalStr)
	if err != nil {
		log.Printf("Invalid WORKER_INTERVAL format %q: %v. Falling back to 1m", workerIntervalStr, err)
		workerInterval = 1 * time.Minute
	}

	worker := background.NewTaskWorker(taskRepo, workerInterval)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		worker.Start(ctx)
	}()

	srv := &http.Server{
		Addr:    port,
		Handler: router,
	}

	go func() {
		log.Printf("Server is running on port %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully, press Ctrl+C again to force")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Waiting for background worker to finish...")

	wg.Wait()

	log.Println("Server exiting properly")
}
