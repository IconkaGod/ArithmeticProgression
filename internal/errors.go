package internal

import "errors"

var (
	ErrEmptyQueue = errors.New("queue is empty")
	ErrNotFound   = errors.New("not found")
)
