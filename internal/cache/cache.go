package cache

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/IconkaGod/ArithmeticProgression/internal"
	"github.com/IconkaGod/ArithmeticProgression/internal/models"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	StatusInQueue    = "IN_QUEUE"
	StatusInProgress = "IN_PROGRESS"
	StatusComplete   = "COMPLETE"
)

type Cache interface {
	Set(ctx context.Context, task models.Task) (int64, error)
	GetById(id int64) (models.Task, error)
	SetComplete(id int64, completeAt time.Time) error
	SetInQueue(id int64, settingAt time.Time) error
	SetInProgress(id int64, startTime time.Time) error
	UpdateProgress(id int64, iteration int, result float64) error
	Delete(id int64) error
	List(ctx context.Context) ([]models.Task, error)
}

type cache struct {
	id   int64
	mu   sync.RWMutex
	repo map[int64]models.Task
	log  *slog.Logger
}

func NewCache(ctx context.Context, l *slog.Logger, cleanupTick time.Duration) Cache {
	c := &cache{
		id:   1,
		mu:   sync.RWMutex{},
		repo: make(map[int64]models.Task),
		log:  l,
	}

	go c.ttlCleanup(ctx, cleanupTick)
	return c
}

func (c *cache) CleanUp() {
	toDelete := make([]int64, 0)

	c.mu.RLock()
	for k, v := range c.repo {
		if v.Status == StatusComplete {
			ttlDuration := time.Duration(v.TTL * float64(time.Second))
			if time.Now().After(v.CompletedAt.Add(ttlDuration)) {
				toDelete = append(toDelete, k)
			}
		}
	}
	c.mu.RUnlock()

	if len(toDelete) > 0 {
		c.log.Info(
			"cleaning up expired tasks",
			slog.String("operation", "cache.cleanUp()"),
			slog.Int("tasks_deleted", len(toDelete)),
		)
		c.mu.Lock()
		for _, id := range toDelete {
			delete(c.repo, id)
		}
		c.mu.Unlock()
	}
}

func (c *cache) ttlCleanup(ctx context.Context, cleanupTick time.Duration) {
	ticker := time.NewTicker(cleanupTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.CleanUp()
		case <-ctx.Done():
			return
		}
	}
}

func (c *cache) Set(ctx context.Context, task models.Task) (int64, error) {
	log := c.log.With(
		slog.String("operation", "cache.Set"),
		slog.Any("request_id", middleware.GetReqID(ctx)),
	)

	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.id
	task.Id = id
	c.repo[id] = task
	c.id++

	log = log.With(
		slog.Int64("task_id", task.Id),
	)

	log.Debug(
		"task saved in cache",
	)

	return id, nil
}

func (c *cache) GetById(id int64) (models.Task, error) {
	log := c.log.With(
		slog.String("operation", "cache.GetById"),
		slog.Int64("task_id", id),
	)

	c.mu.RLock()
	defer c.mu.RUnlock()

	task, ok := c.repo[id]
	if !ok {
		log.Warn(
			"task not found",
		)
		return models.Task{}, internal.ErrNotFound
	}

	log.Debug("task fetched in cache")
	return task, nil
}

func (c *cache) setStatus(id int64, time time.Time, status string) error {
	log := c.log.With(
		slog.String("operation", "cache.setStatus"),
		slog.Int64("task_id", id),
	)

	c.mu.Lock()
	defer c.mu.Unlock()

	task, ok := c.repo[id]
	if !ok {
		log.Warn(
			"task not found",
		)
		return internal.ErrNotFound
	}

	task.Status = status

	switch status {
	case StatusComplete:
		task.CompletedAt = time
	case StatusInProgress:
		task.StartedAt = time
	case StatusInQueue:
		task.SettingAt = time
	}

	c.repo[id] = task

	log.Debug(
		"task status changed",
		slog.String("status", status),
	)

	return nil
}

func (c *cache) SetComplete(id int64, completeAt time.Time) error {
	return c.setStatus(id, completeAt, StatusComplete)
}

func (c *cache) SetInProgress(id int64, startAt time.Time) error {
	return c.setStatus(id, startAt, StatusInProgress)
}

func (c *cache) SetInQueue(id int64, settingAt time.Time) error {
	return c.setStatus(id, settingAt, StatusInQueue)
}

func (c *cache) UpdateProgress(id int64, iteration int, result float64) error {
	log := c.log.With(
		slog.String("operation", "cache.UpdateProgress"),
		slog.Int64("task_id", id),
		slog.Int("iteration", iteration),
		slog.Float64("result", result),
	)

	c.mu.Lock()
	defer c.mu.Unlock()

	task, ok := c.repo[id]
	if !ok {
		log.Warn(
			"task not found",
		)
		return internal.ErrNotFound
	}

	task.Iteration = iteration
	task.Result = result

	c.repo[id] = task

	log.Debug(
		"task progress updated",
	)

	return nil
}

func (c *cache) Delete(id int64) error {
	log := c.log.With(
		slog.String("operation", "cache.Delete"),
		slog.Int64("task_id", id),
	)
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.repo[id]
	if !ok {
		log.Warn(
			"task not found",
		)
		return internal.ErrNotFound
	}

	delete(c.repo, id)

	log.Debug(
		"task deleted",
	)

	return nil
}

func (c *cache) List(ctx context.Context) ([]models.Task, error) {
	log := c.log.With(
		slog.String("operation", "cache.List"),
		slog.Any("request_id", middleware.GetReqID(ctx)),
	)

	c.mu.RLock()
	defer c.mu.RUnlock()

	tasks := make([]models.Task, 0, len(c.repo))

	currentTime := time.Now()

	for _, v := range c.repo {
		if v.Status == StatusComplete {
			ttlDuration := time.Duration(v.TTL * float64(time.Second))

			if currentTime.After(v.CompletedAt.Add(ttlDuration)) {
				continue
			}
		}
		tasks = append(tasks, v)
	}

	filteredCount := len(c.repo) - len(tasks)

	log.Debug(
		"finished listing tasks",
		slog.Int("repo_count", len(c.repo)),
		slog.Int("returned_count", len(tasks)),
		slog.Int("expired_filtered", filteredCount),
	)

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Id < tasks[j].Id
	})

	return tasks, nil
}
