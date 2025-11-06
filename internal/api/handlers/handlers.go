package handlers

import (
	"net/http"

	"github.com/IconkaGod/ArithmeticProgression/internal/models"
	"github.com/IconkaGod/ArithmeticProgression/internal/service"
	"github.com/go-chi/render"
)

type Handlers interface {
	SetTask(w http.ResponseWriter, r *http.Request)
	ListTasks(w http.ResponseWriter, r *http.Request)
}

type handlers struct {
	srv service.Service
}

func NewHandlers(s service.Service) Handlers {
	return &handlers{
		srv: s,
	}
}

func (h *handlers) SetTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskModel

	err := render.DecodeJSON(r.Body, &req)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, models.Response{
			Status: "error",
			Error:  "Invalid JSON format",
		})
		return
	}

	h.srv.SetTask(req)

}

func (h *handlers) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.srv.ListTasks()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, models.Response{
			Status: "error",
			Error:  "Failed to get tasks",
		})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, tasks)
}
