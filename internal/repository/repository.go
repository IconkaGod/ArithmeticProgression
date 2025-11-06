package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/IconkaGod/ArithmeticProgression/internal/lib"
	"github.com/IconkaGod/ArithmeticProgression/internal/models"
)

const (
	StatusInQueue    = "in queue"
	StatusInProgress = "in progress"
	StatusComplete   = "complete"
)

type Repository interface {
	Set(task models.Task) int
	GetById(id int) (models.Task, error)
	SetComplete(id int, completeAt time.Time) error
	SetInQueue(id int, settingAt time.Time) error
	SetInProgress(id int, startTime time.Time) error
	UpdateProgress(id int, iteration int, result float64) error
	Delete(id int) error
	List() []models.Task
}

type repository struct {
	id   int
	mu   sync.RWMutex
	repo map[int]models.Task
}

func NewRepository(ctx context.Context) Repository {
	r := &repository{
		id:   1,
		mu:   sync.RWMutex{},
		repo: make(map[int]models.Task),
	}

	go r.checkTTL(ctx)
	return r
}

func (r *repository) checkTTL(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			toDelete := []int{}

			r.mu.RLock()
			for k, v := range r.repo {
				if v.Status == StatusComplete {
					ttlDuration := time.Duration(v.TTL * float64(time.Second))
					if time.Since(v.CompletedAt) > ttlDuration {
						toDelete = append(toDelete, k)
					}
				}
			}
			r.mu.RUnlock()

			if len(toDelete) > 0 {
				r.mu.Lock()
				for _, id := range toDelete {
					delete(r.repo, id)
				}
				r.mu.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (r *repository) Set(task models.Task) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	task.Id = r.id
	r.repo[r.id] = task

	id := r.id
	r.id++

	return id
}

func (r *repository) GetById(id int) (models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.repo[id]
	if !ok {
		return models.Task{}, lib.ErrNotFound
	}

	return task, nil
}

func (r *repository) SetComplete(id int, completeAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.repo[id]
	if !ok {
		return lib.ErrNotFound
	}

	task.Status = StatusComplete
	task.CompletedAt = completeAt

	r.repo[id] = task

	return nil
}

func (r *repository) SetInProgress(id int, startAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.repo[id]
	if !ok {
		return lib.ErrNotFound
	}

	task.Status = StatusInProgress
	task.StartedAt = startAt

	r.repo[id] = task

	return nil
}

func (r *repository) SetInQueue(id int, settingAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.repo[id]
	if !ok {
		return lib.ErrNotFound
	}

	task.Status = StatusInQueue
	task.SettingAt = settingAt

	r.repo[id] = task

	return nil
}

func (r *repository) UpdateProgress(id int, iteration int, result float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.repo[id]
	if !ok {
		return lib.ErrNotFound
	}

	task.Iteration = iteration
	task.Result = result

	r.repo[id] = task

	return nil
}

func (r *repository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.repo[id]
	if !ok {
		return lib.ErrNotFound
	}

	delete(r.repo, id)

	return nil
}

func (r *repository) List() []models.Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]models.Task, 0, len(r.repo))

	currentTime := time.Now()

	for _, v := range r.repo {
		if v.Status == StatusComplete {
			ttlDuration := time.Duration(v.TTL * float64(time.Second))

			if currentTime.Sub(v.CompletedAt) > ttlDuration {
				continue
			}
		}
		tasks = append(tasks, v)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Id < tasks[j].Id
	})

	return tasks
}
