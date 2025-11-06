package service

import (
	"time"

	"github.com/IconkaGod/ArithmeticProgression/internal/models"
	"github.com/IconkaGod/ArithmeticProgression/internal/queue"
	"github.com/IconkaGod/ArithmeticProgression/internal/repository"
)

type Service interface {
	SetTask(task models.CreateTaskModel) error
	GetTaskById(id int) (models.Task, error)
	SetInProgress(id int, startTime time.Time) error
	SetComplete(id int, completeAt time.Time) error
	UpdateProgress(id int, iteration int, result float64) error
	Delete(id int) error
	ListTasks() ([]models.Task, error)
}

type service struct {
	queue queue.Queue
	repo  repository.Repository
}

func NewService(q queue.Queue, r repository.Repository) Service {
	return &service{
		queue: q,
		repo:  r,
	}
}

func (s *service) SetTask(task models.CreateTaskModel) error {
	newTask := models.Task{
		ElementsCount: task.ElementsCount,
		Delta:         task.Delta,
		StartNumber:   task.StartNumber,
		Interval:      task.Interval,
		TTL:           task.TTL,
	}

	id := s.repo.Set(newTask)

	settingTime := time.Now()
	err := s.repo.SetInQueue(id, settingTime)
	if err != nil {
		return err
	}

	s.queue.Push(id)

	return nil
}

func (s *service) GetTaskById(id int) (models.Task, error) {
	task, err := s.repo.GetById(id)
	if err != nil {
		return models.Task{}, err
	}

	return task, nil
}

func (s *service) SetInProgress(id int, startTime time.Time) error {
	err := s.repo.SetInProgress(id, startTime)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) SetComplete(id int, completeAt time.Time) error {
	err := s.repo.SetComplete(id, completeAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateProgress(id int, iteration int, result float64) error {
	err := s.repo.UpdateProgress(id, iteration, result)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) Delete(id int) error {
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) ListTasks() ([]models.Task, error) {
	models := s.repo.List()

	for _, v := range models {
		if v.Status == "in queue" {
			v.NumInQueue = s.queue.GetPos(v.Id)
		}
	}

	return models, nil
}
