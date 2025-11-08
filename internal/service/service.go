package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/IconkaGod/ArithmeticProgression/internal"
	"github.com/IconkaGod/ArithmeticProgression/internal/cache"
	"github.com/IconkaGod/ArithmeticProgression/internal/models"
	"github.com/IconkaGod/ArithmeticProgression/internal/queue"
	"github.com/go-chi/chi/v5/middleware"
)

type Service interface {
	SetTask(ctx context.Context, task models.CreateTaskModel) error
	GetTaskById(id int64) (models.Task, error)
	SetInProgress(id int64, startTime time.Time) error
	SetComplete(id int64, completeAt time.Time) error
	UpdateProgress(id int64, iteration int, result float64) error
	Delete(d int64) error
	ListTasks(ctx context.Context) ([]models.Task, error)
}

type service struct {
	queue queue.Queue
	cache cache.Cache
	log   *slog.Logger
}

func NewService(q queue.Queue, c cache.Cache, l *slog.Logger) Service {
	return &service{
		queue: q,
		cache: c,
		log:   l,
	}
}

func (s *service) SetTask(ctx context.Context, task models.CreateTaskModel) error {
	log := s.log.With(
		slog.String("operation", "service.SetTask"),
		slog.Any("request_id", middleware.GetReqID(ctx)),
	)

	log.Info(
		"task payload",
		slog.Int("elements_count", task.ElementsCount),
		slog.Float64("delta", task.Delta),
		slog.Float64("start_number", task.StartNumber),
		slog.Float64("interval", task.Interval),
		slog.Float64("ttl", task.TTL),
	)

	newTask := models.Task{
		ElementsCount: task.ElementsCount,
		Delta:         task.Delta,
		StartNumber:   task.StartNumber,
		Interval:      task.Interval,
		TTL:           task.TTL,
	}

	id, err := s.cache.Set(ctx, newTask)
	if err != nil {
		log.Error(
			"failed to create task in cache",
			slog.Any("error", err),
		)
		return err
	}

	log = log.With(
		slog.Int64("task_id", id),
	)

	log.Info(
		"task created",
	)

	settingTime := time.Now()

	err = s.cache.SetInQueue(id, settingTime)
	if err != nil {
		log.Error(
			"failed to change status to IN_QUEUE",
			slog.Any("error", err),
		)
		return err
	}

	s.queue.Push(id)

	log.Info(
		"task status changed to IN_QUEUE",
	)

	return nil
}

func (s *service) GetTaskById(id int64) (models.Task, error) {
	log := s.log.With(
		slog.String("operation", "service.GetTaskById"),
		slog.Int64("task_id", id),
	)

	task, err := s.cache.GetById(id)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			log.Warn(
				"task not found in cache",
				slog.Any("error", err),
			)
		} else {
			log.Error(
				"failed to get task",
				slog.Any("error", err),
			)
		}
		return models.Task{}, err
	}

	log.Info(
		"task fetched by id",
		slog.String("status", task.Status),
		slog.Int("number_in_queue", task.NumInQueue),
		slog.Int("iteration", task.Iteration),
		slog.Float64("result", task.Result),
	)

	return task, nil
}

func (s *service) SetInProgress(id int64, startTime time.Time) error {
	log := s.log.With(
		slog.String("operation", "service.SetInProgress"),
		slog.Int64("task_id", id),
	)

	err := s.cache.SetInProgress(id, startTime)
	if err != nil {
		log.Error(
			"failed to change status to IN_PROGRESS",
			slog.Any("error", err),
		)
		return err
	}

	log.Info(
		"task status changed to IN_PROGRESS",
	)

	return nil
}

func (s *service) SetComplete(id int64, completeAt time.Time) error {
	log := s.log.With(
		slog.String("operation", "service.SetComplete"),
		slog.Int64("task_id", id),
	)

	err := s.cache.SetComplete(id, completeAt)
	if err != nil {
		log.Error(
			"failed to change status to COMPLETE",
			slog.Any("error", err),
		)
		return err
	}

	log.Info(
		"task status changed to COMPLETE",
	)

	return nil
}

func (s *service) UpdateProgress(id int64, iteration int, result float64) error {
	log := s.log.With(
		slog.String("operation", "service.UpdateProgress"),
		slog.Int64("task_id", id),
	)

	err := s.cache.UpdateProgress(id, iteration, result)
	if err != nil {
		log.Error(
			"failed to update progress",
			slog.Any("error", err),
		)
		return err
	}

	log.Info(
		"task progress changed",
		slog.Int("iteration", iteration),
		slog.Float64("result", result),
	)

	return nil
}

func (s *service) Delete(id int64) error {
	log := s.log.With(
		slog.String("operation", "service.Delete"),
		slog.Int64("task_id", id),
	)

	err := s.cache.Delete(id)
	if err != nil {
		log.Error(
			"failed to delete task",
			slog.Any("error", err),
		)
		return err
	}

	log.Info(
		"task deleted",
	)

	return nil
}

func (s *service) ListTasks(ctx context.Context) ([]models.Task, error) {
	log := s.log.With(
		slog.String("operation", "service.ListTasks"),
		slog.Any("request_id", middleware.GetReqID(ctx)),
	)

	models, err := s.cache.List(ctx)
	if err != nil {
		log.Error(
			"failed to listing task",
			slog.Any("error", err),
		)
		return nil, err
	}

	for i, v := range models {
		if v.Status == "IN_QUEUE" {
			models[i].NumInQueue = s.queue.GetPos(v.Id)
		}
	}

	log.Info(
		"task listed",
	)

	return models, nil
}
