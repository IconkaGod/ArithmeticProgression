package cache

import (
	"context"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/IconkaGod/ArithmeticProgression/internal/models"
)

func Test_cache_ttlCleanup(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setupCache func() Cache
		wantCount  int
		waitTime   time.Duration
	}{
		{
			name: "should delete expired complete task",
			setupCache: func() Cache {
				cache := NewCache(context.Background(), slog.Default(), 1*time.Millisecond)
				cache.Set(ctx, models.Task{
					Id:          1,
					TTL:         0.001,
					Status:      "COMPLETE",
					CompletedAt: time.Now(),
				})
				return cache
			},
			wantCount: 0,
			waitTime:  50 * time.Millisecond,
		},
		{
			name: "should not delete non-expired task",
			setupCache: func() Cache {
				cache := NewCache(context.Background(), slog.Default(), 1*time.Millisecond)
				cache.Set(ctx, models.Task{
					Id:          2,
					TTL:         1000,
					Status:      "COMPLETE",
					CompletedAt: time.Now(),
				})
				return cache
			},
			wantCount: 1,
			waitTime:  50 * time.Millisecond,
		},
		{
			name: "should not delete in-progress task even if TTL expired",
			setupCache: func() Cache {
				cache := NewCache(context.Background(), slog.Default(), 1*time.Millisecond)
				cache.Set(ctx, models.Task{
					Id:          3,
					TTL:         0.001,
					Status:      "IN_PROGRESS",
					CompletedAt: time.Now(),
				})
				return cache
			},
			wantCount: 1,
			waitTime:  50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := tt.setupCache()

			time.Sleep(tt.waitTime)

			tasks, err := cache.List(ctx)
			if err != nil {
				t.Fatalf("List failed: %v", err)
			}

			if len(tasks) != tt.wantCount {
				t.Errorf("Expected %d tasks, got %d", tt.wantCount, len(tasks))
			}
		})
	}
}

func Test_cache_GetById(t *testing.T) {

	cache := NewCache(context.Background(), slog.Default(), 1*time.Second)

	now := time.Now()
	testTask := models.Task{
		Id:            1,
		ElementsCount: 10,
		Delta:         2.5,
		StartNumber:   1.0,
		Interval:      1.0,
		TTL:           60.0,
		Status:        "IN_QUEUE",
		Iteration:     0,
		Result:        0,
		SettingAt:     now,
		StartedAt:     time.Time{},
		CompletedAt:   time.Time{},
	}

	taskID, err := cache.Set(context.Background(), testTask)
	if err != nil {
		t.Fatalf("Failed to set test task: %v", err)
	}

	tests := []struct {
		name    string
		id      int
		want    models.Task
		wantErr bool
	}{
		{
			name:    "existing task",
			id:      taskID,
			want:    testTask,
			wantErr: false,
		},
		{
			name:    "not found error",
			id:      999,
			want:    models.Task{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := cache.GetById(tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetById() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetById() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("GetById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cache_setStatus(t *testing.T) {
	cache := NewCache(context.Background(), slog.Default(), 1*time.Second)
	cache.Set(context.Background(), models.Task{})

	inQueueTime := time.Now()
	inProgressTime := time.Now()
	completeTime := time.Now()
	tests := []struct {
		name       string
		id         int
		time       time.Time
		status     string
		wantErr    bool
		checkField string
	}{
		{
			name:       "set IN_QUEUE status",
			id:         1,
			time:       inQueueTime,
			status:     "IN_QUEUE",
			wantErr:    false,
			checkField: "SettingAt",
		},

		{
			name:       "set IN_PROGRESS status",
			id:         1,
			time:       inProgressTime,
			status:     "IN_PROGRESS",
			wantErr:    false,
			checkField: "StartedAt",
		},

		{
			name:       "set COMPLETE status",
			id:         1,
			time:       completeTime,
			status:     "COMPLETE",
			wantErr:    false,
			checkField: "CompletedAt",
		},

		{
			name:    "not found error",
			id:      112312,
			time:    time.Now(),
			status:  "COMPLETE",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotErr error
			switch tt.status {
			case "IN_QUEUE":
				gotErr = cache.SetInQueue(tt.id, tt.time)
			case "IN_PROGRESS":
				gotErr = cache.SetInProgress(tt.id, tt.time)
			case "COMPLETE":
				gotErr = cache.SetComplete(tt.id, tt.time)
			}

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("setStatus() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			model, err := cache.GetById(tt.id)
			if err != nil {
				t.Fatalf("GetById error: %v", err)
			}

			if model.Status != tt.status {
				t.Errorf("GetById().Status = %v, want %v", model.Status, tt.status)
			}

			switch tt.checkField {
			case "SettingAt":
				if !model.SettingAt.Equal(tt.time) {
					t.Errorf("GetById().SettingAt = %v, want %v", model.SettingAt, tt.time)
				}
			case "StartedAt":
				if !model.StartedAt.Equal(tt.time) {
					t.Errorf("GetById().StartedAt = %v, want %v", model.StartedAt, tt.time)
				}
			case "CompletedAt":
				if !model.CompletedAt.Equal(tt.time) {
					t.Errorf("GetById().CompletedAt = %v, want %v", model.CompletedAt, tt.time)
				}
			}
		})
	}
}

func Test_cache_UpdateProgress(t *testing.T) {
	cache := NewCache(context.Background(), slog.Default(), 1*time.Second)

	cache.Set(context.Background(), models.Task{
		ElementsCount: 10,
		Delta:         2.5,
		StartNumber:   1.0,
		Interval:      1.0,
		TTL:           60.0,
		Status:        "IN_PROGRESS",
		Iteration:     0,
		Result:        0,
	})

	tests := []struct {
		name      string
		id        int
		iteration int
		result    float64
		wantErr   bool
		wantTask  models.Task
	}{
		{
			name:      "successful progress update",
			id:        1,
			iteration: 5,
			result:    12.5,
			wantErr:   false,
			wantTask: models.Task{
				ElementsCount: 10,
				Delta:         2.5,
				StartNumber:   1.0,
				Interval:      1.0,
				TTL:           60.0,
				Status:        "IN_PROGRESS",
				Iteration:     5,
				Result:        12.5,
			},
		},
		{
			name:      "task not found",
			id:        999,
			iteration: 1,
			result:    1.0,
			wantErr:   true,
			wantTask:  models.Task{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.UpdateProgress(tt.id, tt.iteration, tt.result)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("UpdateProgress() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}

			if tt.wantErr {
				return
			}

			gotTask, err := cache.GetById(tt.id)
			if err != nil {
				t.Fatalf("GetById failed after UpdateProgress: %v", err)
			}

			if reflect.DeepEqual(gotTask, tt.wantTask) {
				t.Errorf("Expected %v, got %v", tt.wantTask, gotTask)
			}
		})
	}
}

func Test_cache_Delete(t *testing.T) {
	cache := NewCache(context.Background(), slog.Default(), 1*time.Second)
	cache.Set(context.Background(), models.Task{})

	tests := []struct {
		name    string
		id      int
		wantErr bool
	}{
		{
			name:    "success delete",
			id:      1,
			wantErr: false,
		},
		{
			name:    "not found",
			id:      1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Delete(tt.id)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Delete() failed: %v", err)
				}
			}

			if tt.wantErr {
				return
			}

			_, err = cache.GetById(tt.id)
			if err == nil {
				t.Errorf("GetById() succed: %v", err)
			}
		})
	}
}
