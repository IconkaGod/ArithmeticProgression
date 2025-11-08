package cache

import (
	"context"
	"log/slog"
	"testing"
	"testing/synctest"
	"time"

	"github.com/IconkaGod/ArithmeticProgression/internal/models"
	"github.com/stretchr/testify/assert"
)

func Test_cache_ttlCleanup(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := NewCache(t.Context(), slog.Default(), 12*time.Second)

		c.Set(t.Context(), models.Task{
			TTL:         0.001,
			Status:      "COMPLETE",
			CompletedAt: time.Now().Add(-1 * time.Second),
		})

		c.Set(t.Context(), models.Task{
			TTL:         1000,
			Status:      "COMPLETE",
			CompletedAt: time.Now(),
		})

		c.Set(t.Context(), models.Task{
			TTL:       0.001,
			Status:    "IN_PROGRESS",
			StartedAt: time.Now(),
		})

		c.(*cache).CleanUp()
		synctest.Wait()

		tasks, err := c.List(t.Context())
		assert.NoError(t, err, "List shouldn't fail after cleanup")

		assert.Len(t, tasks, 2, "Expected 2 tasks to remain after cleanup")
	})
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
		id      int64
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

			if tt.wantErr {
				assert.Error(t, gotErr, "Expected error for not found task")
				return
			}

			assert.NoError(t, gotErr, "Should not get error for existing task")
			assert.EqualValues(t, tt.want, got, "Task mismatch")
		})
	}
}

func Test_cache_setStatus(t *testing.T) {
	cache := NewCache(context.Background(), slog.Default(), 1*time.Second)
	taskID, _ := cache.Set(context.Background(), models.Task{})

	inQueueTime := time.Now()
	inProgressTime := time.Now()
	completeTime := time.Now()
	tests := []struct {
		name       string
		id         int64
		time       time.Time
		status     string
		wantErr    bool
		checkField string
	}{
		{
			name:       "set IN_QUEUE status",
			id:         taskID,
			time:       inQueueTime,
			status:     "IN_QUEUE",
			wantErr:    false,
			checkField: "SettingAt",
		},

		{
			name:       "set IN_PROGRESS status",
			id:         taskID,
			time:       inProgressTime,
			status:     "IN_PROGRESS",
			wantErr:    false,
			checkField: "StartedAt",
		},

		{
			name:       "set COMPLETE status",
			id:         taskID,
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

			if tt.wantErr {
				assert.Error(t, gotErr, "Expected error for 'not found' case")
				return
			}

			assert.NoError(t, gotErr, "Error setting status")

			model, err := cache.GetById(tt.id)

			assert.NoError(t, err, "GetById shouldn't fail after successful SetStatus")
			assert.Equal(t, tt.status, model.Status, "Status mismatch")

			switch tt.checkField {
			case "SettingAt":
				assert.True(t, model.SettingAt.Equal(tt.time), "SettingAt mismatch")
			case "StartedAt":
				assert.True(t, model.StartedAt.Equal(tt.time), "StartedAt mismatch")
			case "CompletedAt":
				assert.True(t, model.CompletedAt.Equal(tt.time), "CompletedAt mismatch")
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
		id        int64
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

			if tt.wantErr {
				assert.Error(t, err, "Expected error for 'not found' case")
				return
			}

			assert.NoError(t, err, "UpdateProgress failed unexpectedly")

			gotTask, err := cache.GetById(tt.id)

			assert.NoError(t, err, "GetById failed after UpdateProgress")
			assert.Equal(t, tt.wantTask.Iteration, gotTask.Iteration, "Iteration mismatch")
			assert.InDelta(t, tt.wantTask.Result, gotTask.Result, 0.0001, "Result mismatch")
		})
	}
}

func Test_cache_Delete(t *testing.T) {
	cache := NewCache(context.Background(), slog.Default(), 1*time.Second)
	cache.Set(context.Background(), models.Task{})

	tests := []struct {
		name    string
		id      int64
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

			if tt.wantErr {
				assert.Error(t, err, "Expected error for 'not found' case")
				return
			}

			assert.NoError(t, err, "Delete failed unexpectedly")

			_, err = cache.GetById(tt.id)
			assert.Error(t, err, "GetById should fail for deleted task")
		})
	}
}

func Test_cache_Set(t *testing.T) {
	cache := NewCache(context.Background(), slog.Default(), 1*time.Second)

	id1, err := cache.Set(context.Background(), models.Task{})
	assert.NoError(t, err, "Set failed for ID 1")
	assert.Equal(t, int64(1), id1, "ID 1 mismatch")

	id2, err := cache.Set(context.Background(), models.Task{})
	assert.NoError(t, err, "Set failed for ID 2")
	assert.Equal(t, int64(2), id2, "ID 2 mismatch")

	task, _ := cache.GetById(id1)
	assert.Equal(t, id1, task.Id, "Task ID mismatch in repo")
	assert.Equal(t, models.Task{}.ElementsCount, task.ElementsCount, "ElementsCount should be 0")
}

func Test_cache_List(t *testing.T) {
	ctx := context.Background()
	cache := NewCache(ctx, slog.Default(), 5*time.Second)

	cache.Set(ctx, models.Task{ElementsCount: 3, Status: "IN_PROGRESS", TTL: 100})
	cache.Set(ctx, models.Task{ElementsCount: 1, Status: "IN_QUEUE", TTL: 100})
	cache.Set(ctx, models.Task{ElementsCount: 3, Status: "COMPLETE", TTL: 100, CompletedAt: time.Now()})
	cache.Set(ctx, models.Task{ElementsCount: 1, Status: "COMPLETE", TTL: 100, CompletedAt: time.Now()})
	cache.Set(ctx, models.Task{ElementsCount: 2, Status: "COMPLETE", TTL: 100, CompletedAt: time.Now()})
	cache.Set(ctx, models.Task{ElementsCount: 2, Status: "COMPLETE", TTL: 100})

	tasks, err := cache.List(ctx)

	assert.NoError(t, err, "List failed unexpectedly")
	assert.Len(t, tasks, 5, "Expected 5 tasks in the primary list (ID 6 filtered)")

	expectedIDs := []int64{1, 2, 3, 4, 5}
	for i, expectedID := range expectedIDs {
		assert.Equal(t, expectedID, tasks[i].Id, "Tasks not sorted by ID at index %d", i)
	}

	cache.Set(ctx, models.Task{
		Status:      "COMPLETE",
		TTL:         0.001,
		CompletedAt: time.Now().Add(-1 * time.Second),
	})

	tasksFiltered, err := cache.List(ctx)
	assert.NoError(t, err, "List failed on second call")
	assert.Len(t, tasksFiltered, 5, "Expected 5 tasks after filtration, the new expired task should be excluded")
}
