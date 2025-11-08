package handlers

import (
	"log/slog"
	"net/http"

	"github.com/IconkaGod/ArithmeticProgression/internal/models"
	"github.com/IconkaGod/ArithmeticProgression/internal/service"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Handlers interface {
	SetTask(w http.ResponseWriter, r *http.Request)
	ListTasks(w http.ResponseWriter, r *http.Request)
}

type handlers struct {
	srv service.Service
	log *slog.Logger
}

func NewHandlers(s service.Service, l *slog.Logger) Handlers {
	return &handlers{
		srv: s,
		log: l,
	}
}

func (h *handlers) SetTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log := h.log.With(
		slog.String("operation", "handler.SetTask"),
		slog.Any("request_id", middleware.GetReqID(ctx)),
	)

	var req models.CreateTaskModel

	err := render.DecodeJSON(r.Body, &req)
	if err != nil {
		log.Warn(
			"invalid request format",
			slog.Any("error", err),
			slog.Int("status", http.StatusBadRequest),
		)

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, models.Response{
			Status: "error",
			Error:  "Invalid JSON format",
		})
		return
	}
	defer r.Body.Close()

	if req.Interval < 0 || req.ElementsCount < 0 {
		log.Warn(
			"validation error",
			slog.Int("status", http.StatusBadRequest),
		)

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, models.Response{
			Status: "error",
			Error:  "Validation error",
		})
		return
	}

	err = h.srv.SetTask(ctx, req)
	if err != nil {
		log.Error(
			"failed to set task",
			slog.Any("error", err),
			slog.Int("status", http.StatusInternalServerError),
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, models.Response{
			Status: "error",
			Error:  "Failed to set task",
		})

		return
	}

	log.Info(
		"task created",
		slog.Int("status", http.StatusCreated),
	)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, models.Response{
		Status: "ok",
	})
}

func (h *handlers) ListTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log := h.log.With(
		slog.String("operation", "handler.SetTask"),
		slog.Any("request_id", middleware.GetReqID(ctx)),
	)

	tasks, err := h.srv.ListTasks(ctx)
	if err != nil {
		log.Error(
			"failed to list task",
			slog.Any("error", err),
			slog.Int("status", http.StatusInternalServerError),
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, models.Response{
			Status: "error",
			Error:  "Failed to get tasks",
		})
		return
	}

	log.Info(
		"task listed",
		slog.Int("status", http.StatusOK),
	)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, tasks)
}
