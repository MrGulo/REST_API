package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(userHandler *UserHandler, taskHandler *TaskHandler, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/register", userHandler.Register)
	r.Post("/login", userHandler.Login)

	r.Route("/tasks", func(r chi.Router) {
		r.Use(AuthMiddleware(jwtSecret))

		r.Post("/", taskHandler.Create)
		r.Get("/", taskHandler.GetAll)
		r.Get("/{id}", taskHandler.GetByID)
		r.Put("/{id}", taskHandler.Update)
		r.Delete("/{id}", taskHandler.Delete)
	})

	return r
}
