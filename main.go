package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"example.com/pz4-todo/internal/task"
	myMW "example.com/pz4-todo/pkg/middleware"
)

func main() {
	repo := task.NewRepo()
	repo.Load()
	h := task.NewHandler(repo)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.Recoverer)
	r.Use(myMW.Logger)
	r.Use(myMW.SimpleCORS)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Route("/v1", func(v1 chi.Router) {
		v1.Mount("/tasks", h.Routes())
	})

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
