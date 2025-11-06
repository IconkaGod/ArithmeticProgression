package queue

import (
	"sync"

	"github.com/IconkaGod/ArithmeticProgression/internal/lib"
)

type Queue interface {
	Push(id int)
	Pop() (int, error)
	GetPos(id int) int
}

type queue struct {
	mu   sync.RWMutex
	buff []int
}

func NewQueue() Queue {
	return &queue{
		buff: make([]int, 0),
	}
}

func (q *queue) Push(id int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.buff = append(q.buff, id)
}

func (q *queue) Pop() (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.buff) == 0 {
		return 0, lib.EmptyQueue
	}

	id := q.buff[0]
	q.buff = q.buff[1:]

	return id, nil
}

func (q *queue) GetPos(id int) int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for i, v := range q.buff {
		if v == id {
			return i + 1
		}
	}

	return -1
}
