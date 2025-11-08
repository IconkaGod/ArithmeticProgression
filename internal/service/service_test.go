package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/IconkaGod/ArithmeticProgression/internal"
	"github.com/IconkaGod/ArithmeticProgression/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(ctx context.Context, task models.Task) (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCache) GetById(id int64) (models.Task, error) {
	args := m.Called()
	return args.Get(0).(models.Task), args.Error(1)
}

func (m *MockCache) SetComplete(id int64, completeAt time.Time) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCache) SetInProgress(id int64, startAt time.Time) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCache) SetInQueue(id int64, settingAt time.Time) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCache) UpdateProgress(id int64, iteration int, result float64) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCache) Delete(id int64) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCache) List(ctx context.Context) ([]models.Task, error) {
	args := m.Called()
	return args.Get(0).([]models.Task), args.Error(1)
}

type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) Push(id int64) {
	m.Called(id)
}

func (m *MockQueue) Pop() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQueue) GetPos(id int64) int {
	args := m.Called(id)
	return int(args.Int(0))
}

func Test_service_SetTask(t *testing.T) {
	mockCache := new(MockCache)
	mockCache.On("Set").Return(int64(1), nil)
	mockCache.On("SetInQueue").Return(nil)

	mockQueue := new(MockQueue)
	mockQueue.On("Push", int64(1)).Return()

	srv := NewService(mockQueue, mockCache, slog.Default())

	err := srv.SetTask(t.Context(), models.CreateTaskModel{})
	assert.NoError(t, err)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_SetTaskCacheError(t *testing.T) {
	mockCache := new(MockCache)
	mockCache.On("Set").Return(int64(0), errors.New("some error"))

	mockQueue := new(MockQueue)

	srv := NewService(mockQueue, mockCache, slog.Default())

	err := srv.SetTask(t.Context(), models.CreateTaskModel{})
	assert.Error(t, err)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_SetTaskSetInQueueError(t *testing.T) {
	mockCache := new(MockCache)
	mockCache.On("Set").Return(int64(1), nil)
	mockCache.On("SetInQueue").Return(errors.New("some error"))

	mockQueue := new(MockQueue)
	srv := NewService(mockQueue, mockCache, slog.Default())

	err := srv.SetTask(t.Context(), models.CreateTaskModel{})
	assert.Error(t, err)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_GetTaskById(t *testing.T) {
	expectedTask := models.Task{Id: 1, Status: "COMPLETE"}

	mockCache := new(MockCache)

	mockCache.On("GetById").Return(expectedTask, nil)

	mockQueue := new(MockQueue)

	srv := NewService(mockQueue, mockCache, slog.Default())

	task, err := srv.GetTaskById(int64(1))
	assert.NoError(t, err)

	assert.EqualValues(t, expectedTask, task)
	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_GetTaskByIdErrorNotFound(t *testing.T) {

	mockCache := new(MockCache)
	mockCache.On("GetById").Return(models.Task{}, internal.ErrNotFound)

	mockQueue := new(MockQueue)
	srv := NewService(mockQueue, mockCache, slog.Default())

	_, err := srv.GetTaskById(int64(1))
	assert.Error(t, err)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_GetTaskByIdSomeError(t *testing.T) {

	mockCache := new(MockCache)
	mockCache.On("GetById").Return(models.Task{}, errors.New("some error"))

	mockQueue := new(MockQueue)
	srv := NewService(mockQueue, mockCache, slog.Default())

	_, err := srv.GetTaskById(int64(1))
	assert.Error(t, err)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_SetInProgress(t *testing.T) {
	mockCache := new(MockCache)

	mockCache.On("SetInProgress").Return(nil)

	mockQueue := new(MockQueue)

	srv := NewService(mockQueue, mockCache, slog.Default())

	err := srv.SetInProgress(int64(1), time.Now())
	assert.NoError(t, err)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_SetInProgressError(t *testing.T) {
	mockCache := new(MockCache)

	mockCache.On("SetInProgress").Return(errors.New("some error"))

	mockQueue := new(MockQueue)

	srv := NewService(mockQueue, mockCache, slog.Default())

	err := srv.SetInProgress(int64(1), time.Now())
	assert.Error(t, err)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_SetComplete(t *testing.T) {
	mockCache := new(MockCache)

	mockCache.On("SetComplete").Return(nil)

	mockQueue := new(MockQueue)

	srv := NewService(mockQueue, mockCache, slog.Default())

	err := srv.SetComplete(int64(1), time.Now())
	assert.NoError(t, err)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_SetCompleteError(t *testing.T) {
	mockCache := new(MockCache)

	mockCache.On("SetComplete").Return(errors.New("some error"))

	mockQueue := new(MockQueue)

	srv := NewService(mockQueue, mockCache, slog.Default())

	err := srv.SetComplete(int64(1), time.Now())
	assert.Error(t, err)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_UpdateProgress(t *testing.T) {
	mockCache := new(MockCache)

	mockCache.On("UpdateProgress").Return(nil)

	mockQueue := new(MockQueue)

	srv := NewService(mockQueue, mockCache, slog.Default())

	err := srv.UpdateProgress(int64(1), 1, 10.5)
	assert.NoError(t, err)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_UpdateProgressError(t *testing.T) {
	mockCache := new(MockCache)

	mockCache.On("UpdateProgress").Return(errors.New("some error"))

	mockQueue := new(MockQueue)

	srv := NewService(mockQueue, mockCache, slog.Default())

	err := srv.UpdateProgress(int64(1), 1, 10.5)
	assert.Error(t, err)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_Delete(t *testing.T) {
	mockCache := new(MockCache)

	mockCache.On("Delete").Return(nil)

	mockQueue := new(MockQueue)

	srv := NewService(mockQueue, mockCache, slog.Default())

	err := srv.Delete(int64(1))
	assert.NoError(t, err)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_DeleteError(t *testing.T) {
	mockCache := new(MockCache)

	mockCache.On("Delete").Return(errors.New("some error"))

	mockQueue := new(MockQueue)

	srv := NewService(mockQueue, mockCache, slog.Default())

	err := srv.Delete(int64(1))
	assert.Error(t, err)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_ListTasks(t *testing.T) {
	returningTasks := []models.Task{
		{
			Id:            1,
			Status:        "IN_PROGRESS",
			ElementsCount: 3,
			TTL:           100.0,
		},
		{
			Id:            2,
			Status:        "IN_QUEUE",
			ElementsCount: 1,
			TTL:           100.0,
		},
		{
			Id:            3,
			Status:        "COMPLETE",
			ElementsCount: 3,
			TTL:           100.0,
		},
	}

	expectedTasks := []models.Task{
		{
			Id:            1,
			Status:        "IN_PROGRESS",
			ElementsCount: 3,
			TTL:           100.0,
		},
		{
			Id:            2,
			Status:        "IN_QUEUE",
			ElementsCount: 1,
			NumInQueue:    1,
			TTL:           100.0,
		},
		{
			Id:            3,
			Status:        "COMPLETE",
			ElementsCount: 3,
			TTL:           100.0,
		},
	}

	mockCache := new(MockCache)

	mockCache.On("List").Return(returningTasks, nil)

	mockQueue := new(MockQueue)
	mockQueue.On("GetPos", int64(2)).Return(1).Once()

	srv := NewService(mockQueue, mockCache, slog.Default())

	tasks, err := srv.ListTasks(t.Context())

	assert.NoError(t, err)
	assert.EqualValues(t, expectedTasks, tasks)

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func Test_service_ListTasksError(t *testing.T) {
	mockCache := new(MockCache)

	mockCache.On("List").Return([]models.Task{}, errors.New("some error"))

	mockQueue := new(MockQueue)

	srv := NewService(mockQueue, mockCache, slog.Default())

	tasks, err := srv.ListTasks(t.Context())

	assert.Error(t, err)
	assert.Nil(t, tasks, "Tasks slice should be nil on error")

	mockCache.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}
