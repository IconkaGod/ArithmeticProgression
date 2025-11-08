package models

import (
	"encoding/json"
	"time"
)

type Task struct {
	Id            int64     `json:"id"`
	Status        string    `json:"status"`
	NumInQueue    int       `json:"number_in_queue,omitempty"`
	ElementsCount int       `json:"elements_count"`
	Delta         float64   `json:"delta"`
	StartNumber   float64   `json:"start_number"`
	Interval      float64   `json:"interval"`
	TTL           float64   `json:"ttl"`
	Iteration     int       `json:"iteration"`
	Result        float64   `json:"result"`
	SettingAt     time.Time `json:"setting_at,omitempty"`
	StartedAt     time.Time `json:"started_at,omitempty"`
	CompletedAt   time.Time `json:"completed_at,omitempty"`
}

type CreateTaskModel struct {
	ElementsCount int     `json:"elements_count"`
	Delta         float64 `json:"delta"`
	StartNumber   float64 `json:"start_number"`
	Interval      float64 `json:"interval"`
	TTL           float64 `json:"ttl"`
}

func (t Task) MarshalJSON() ([]byte, error) {
	type Alias Task
	aux := &struct {
		StartedAt   interface{} `json:"started_at,omitempty"`
		SettingAt   interface{} `json:"setting_at,omitempty"`
		CompletedAt interface{} `json:"completed_at,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(&t),
	}

	if !t.StartedAt.IsZero() {
		aux.StartedAt = t.StartedAt
	}
	if !t.SettingAt.IsZero() {
		aux.SettingAt = t.SettingAt
	}
	if !t.CompletedAt.IsZero() {
		aux.CompletedAt = t.CompletedAt
	}

	return json.Marshal(aux)
}
