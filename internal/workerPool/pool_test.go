package pool

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/IconkaGod/ArithmeticProgression/internal"
	"github.com/IconkaGod/ArithmeticProgression/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) SetTask(ctx context.Context, task models.CreateTaskModel) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockService) GetTaskById(id int64) (models.Task, error) {
	args := m.Called(id)
	return args.Get(0).(models.Task), args.Error(1)
}

func (m *MockService) SetInProgress(id int64, startAt time.Time) error {
	args := m.Called(id, startAt)
	return args.Error(0)
}

func (m *MockService) SetComplete(id int64, completeAt time.Time) error {
	args := m.Called(id, completeAt)
	return args.Error(0)
}

func (m *MockService) UpdateProgress(id int64, iteration int, result float64) error {
	args := m.Called(id, iteration, result)
	return args.Error(0)
}

func (m *MockService) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockService) ListTasks(ctx context.Context) ([]models.Task, error) {
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

func Test_pool_processTask(t *testing.T) {

	mockQueue := new(MockQueue)
	mockService := new(MockService)

	var wg sync.WaitGroup
	wg.Add(1)

	taskID := int64(1)
	testTask := models.Task{
		Id:            1,
		ElementsCount: 3,
		StartNumber:   1.0,
		Delta:         2.0,
		Interval:      0.1,
	}

	mockQueue.On("Pop").Return(taskID, nil).Once()
	mockQueue.On("Pop").Return(int64(0), internal.ErrEmptyQueue)

	mockService.On("GetTaskById", taskID).Return(testTask, nil).Once()
	mockService.On("SetInProgress", taskID, mock.Anything).Return(nil).Once()

	mockService.On("UpdateProgress", taskID, 1, float64(3.0)).Return(nil).Once()
	mockService.On("UpdateProgress", taskID, 2, float64(5.0)).Return(nil).Once()

	mockService.On("SetComplete", taskID, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
		wg.Done()
	})

	pool := NewWorkerPool(context.Background(), 1, mockQueue, mockService, slog.Default())
	pool.StartWorkers()

	wg.Wait()

	mockQueue.AssertExpectations(t)
	mockService.AssertExpectations(t)

}

func Test_pool_processTaskPopError(t *testing.T) {

	mockQueue := new(MockQueue)
	mockService := new(MockService)

	var wg sync.WaitGroup
	wg.Add(1)

	mockQueue.On("Pop").Return(int64(0), errors.New("some error")).Once().Run(func(args mock.Arguments) {
		wg.Done()
	})

	pool := NewWorkerPool(context.Background(), 1, mockQueue, mockService, slog.Default())
	pool.StartWorkers()

	wg.Wait()

	mockQueue.AssertExpectations(t)
	mockService.AssertExpectations(t)

}

func Test_pool_processTaskGetTaskNotFoundError(t *testing.T) {
	mockQueue := new(MockQueue)
	mockService := new(MockService)

	var wg sync.WaitGroup
	wg.Add(1)

	taskID := int64(1)

	mockQueue.On("Pop").Return(taskID, nil).Once()

	block := make(chan struct{})

	mockQueue.On("Pop").Return(int64(0), errors.New("some error")).Run(func(args mock.Arguments) {
		<-block
	})

	mockService.On("GetTaskById", taskID).Return(models.Task{}, internal.ErrNotFound).Once().Run(func(args mock.Arguments) {
		wg.Done()
	})

	pool := NewWorkerPool(context.Background(), 1, mockQueue, mockService, slog.Default())
	pool.StartWorkers()

	wg.Wait()

	mockQueue.AssertExpectations(t)
	mockService.AssertExpectations(t)

	close(block)
}

func Test_pool_processTaskGetTaskError(t *testing.T) {
	mockQueue := new(MockQueue)
	mockService := new(MockService)

	var wg sync.WaitGroup
	wg.Add(1)

	taskID := int64(1)

	mockQueue.On("Pop").Return(taskID, nil).Once()

	block := make(chan struct{})

	mockQueue.On("Pop").Return(int64(0), errors.New("some error")).Run(func(args mock.Arguments) {
		<-block
	})

	mockService.On("GetTaskById", taskID).Return(models.Task{}, errors.New("some error")).Once().Run(func(args mock.Arguments) {
		wg.Done()
	})

	pool := NewWorkerPool(context.Background(), 1, mockQueue, mockService, slog.Default())
	pool.StartWorkers()

	wg.Wait()

	mockQueue.AssertExpectations(t)
	mockService.AssertExpectations(t)

	close(block)
}

func Test_pool_processTaskZeroElementsCount(t *testing.T) {

	mockQueue := new(MockQueue)
	mockService := new(MockService)

	var wg sync.WaitGroup
	wg.Add(1)

	taskID := int64(1)
	testTask := models.Task{
		Id:            1,
		ElementsCount: 0,
		StartNumber:   1.0,
		Delta:         2.0,
		Interval:      0.1,
	}

	mockQueue.On("Pop").Return(taskID, nil).Once()
	mockQueue.On("Pop").Return(int64(0), internal.ErrEmptyQueue)

	mockService.On("GetTaskById", taskID).Return(testTask, nil).Once()

	mockService.On("SetComplete", taskID, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
		wg.Done()
	})

	pool := NewWorkerPool(context.Background(), 1, mockQueue, mockService, slog.Default())
	pool.StartWorkers()

	wg.Wait()

	mockQueue.AssertExpectations(t)
	mockService.AssertExpectations(t)

}

func Test_pool_processTaskContextStop(t *testing.T) {

	mockQueue := new(MockQueue)
	mockService := new(MockService)

	pool := NewWorkerPool(t.Context(), 1, mockQueue, mockService, slog.Default())
	pool.StartWorkers()

	mockQueue.AssertExpectations(t)
	mockService.AssertExpectations(t)
}
