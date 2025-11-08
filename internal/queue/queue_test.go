package queue

import (
	"log/slog"
	"testing"

	"github.com/IconkaGod/ArithmeticProgression/internal"
	"github.com/stretchr/testify/assert"
)

func Test_queue_Pop(t *testing.T) {
	queue := NewQueue(slog.Default())
	queue.Push(1)
	id, err := queue.Pop()

	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, int64(1), id)
}

func Test_queue_PopError(t *testing.T) {
	queue := NewQueue(slog.Default())

	id, err := queue.Pop()

	assert.ErrorIs(t, err, internal.ErrEmptyQueue)
	assert.Equal(t, int64(0), id)
}

func Test_queue_GetPos(t *testing.T) {
	queue := NewQueue(slog.Default())
	queue.Push(1)
	queue.Push(2)
	queue.Push(3)
	queue.Push(4)

	tests := []struct {
		id   int64
		want int
	}{
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4},
	}

	for _, tt := range tests {
		pos := queue.GetPos(tt.id)
		assert.Equal(t, tt.want, pos, "unexpected pos")
	}

	pos := queue.GetPos(5)

	assert.Equal(t, -1, pos, "unexpected pos")
}

func Test_queue_FIFO_Order(t *testing.T) {
	queue := NewQueue(slog.Default())

	items := []int64{1, 2, 3, 4, 5}
	for _, item := range items {
		queue.Push(item)
	}

	for _, want := range items {
		got, err := queue.Pop()

		assert.NoError(t, err, "unexpected error")
		assert.Equal(t, want, got)
	}

	_, err := queue.Pop()
	assert.ErrorIs(t, err, internal.ErrEmptyQueue, "unexpected error")
}

func Test_queue_Push(t *testing.T) {
	queue := NewQueue(slog.Default())

	queue.Push(1)

	id, err := queue.Pop()

	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, int64(1), id)
}
