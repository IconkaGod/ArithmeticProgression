package router

import (
	"net/http"

	"github.com/IconkaGod/ArithmeticProgression/internal/api/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(hand handlers.Handlers) http.Handler {
	r := chi.NewRouter()

	r.Use(
		middleware.RequestID,
	)

	r.Get("/", DefaultRootHandler)

	r.Route("/task", func(r chi.Router) {
		r.Post("/set", hand.SetTask)
		r.Get("/list", hand.ListTasks)
	})

	return r
}

func DefaultRootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("POST /task/set\nGET /task/list"))
}
