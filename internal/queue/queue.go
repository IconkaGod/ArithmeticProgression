package queue

import (
	"log/slog"
	"sync"

	"github.com/IconkaGod/ArithmeticProgression/internal"
)

type Queue interface {
	Push(id int64)
	Pop() (int64, error)
	GetPos(id int64) int
}

type queue struct {
	mu   sync.RWMutex
	buff []int64
	log  *slog.Logger
}

func NewQueue(l *slog.Logger) Queue {
	return &queue{
		buff: make([]int64, 0),
		log:  l,
	}
}

func (q *queue) Push(id int64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	log := q.log.With(
		slog.String("operation", "queue.Push"),
		slog.Int64("task_id", id),
	)

	q.buff = append(q.buff, id)

	log.Debug(
		"task id add to queue",
		slog.Int("queue_length", len(q.buff)),
		slog.Int("queue_cap", cap(q.buff)),
	)
}

func (q *queue) Pop() (int64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	log := q.log.With(
		slog.String("operation", "queue.Pop"),
	)

	if len(q.buff) == 0 {
		return 0, internal.ErrEmptyQueue
	}

	id := q.buff[0]
	q.buff = q.buff[1:]

	log.Debug(
		"task id pop from queue",
		slog.Int64("task_id", id),
		slog.Int("queue_length", len(q.buff)),
		slog.Int("queue_cap", cap(q.buff)),
	)
	return id, nil
}

func (q *queue) GetPos(id int64) int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	log := q.log.With(
		slog.String("operation", "queue.GetPos"),
		slog.Int64("task_id", id),
	)

	for i, v := range q.buff {
		if v == id {
			log.Info(
				"task id fetching in queue",
				slog.Int64("task_id", id),
			)
			return i + 1
		}
	}

	log.Warn(
		"task id not fetching in queue",
	)

	return -1
}
