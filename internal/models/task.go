package models

import "time"

type Task struct {
	Status        string    `json:"status"`
	Delta         float64   `json:"delta"`
	StartNumber   float64   `json:"start_number"`
	Interval      float64   `json:"interval"`
	Result        float64   `json:"result"`
	ElementsCount int       `json:"elements_count"`
	Iteration     int       `json:"iteration"`
	Id            int       `json:"id"`
	NumInQueue    int       `json:"number_in_queue"`
	TTL           float64   `json:"ttl"`
	StartedAt     time.Time `json:"started_at,omitempty"`
	SettingAt     time.Time `json:"setting_at"`
	CompletedAt   time.Time `json:"completed_at,omitempty"`
}

type CreateTaskModel struct {
	ElementsCount int     `json:"elements_count"`
	Delta         float64 `json:"delta"`
	StartNumber   float64 `json:"start_number"`
	Interval      float64 `json:"interval"`
	TTL           float64 `json:"ttl"`
}
