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

	r.Route("/task", func(r chi.Router) {
		r.Post("/set", hand.SetTask)
		r.Get("/list", hand.ListTasks)
	})

	return r
}
