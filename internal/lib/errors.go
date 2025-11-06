package lib

import "errors"

var (
	EmptyQueue  = errors.New("queue is empty")
	ErrNotFound = errors.New("not found")
)
