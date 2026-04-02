package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"task-rest-api/internal/handlers"
	"task-rest-api/internal/storage"
)

func main() {
	s := storage.New()
	taskHandler := handlers.NewTaskHandler(s)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Task REST API - use /tasks"))
	})

	r.Route("/tasks", func(r chi.Router) {
		r.Get("/", taskHandler.List)
		r.Post("/", taskHandler.Create)
		r.Delete("/", taskHandler.DeleteAll)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", taskHandler.Get)
			r.Put("/", taskHandler.Update)
			r.Delete("/", taskHandler.Delete)
		})
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
