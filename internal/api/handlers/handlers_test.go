package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/IconkaGod/ArithmeticProgression/internal/models"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) SetTask(ctx context.Context, task models.CreateTaskModel) error {
	args := m.Called(task)
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

func Test_handlers_SetTask(t *testing.T) {
	mockService := new(MockService)
	h := NewHandlers(mockService, slog.Default())

	expectedModel := models.CreateTaskModel{
		ElementsCount: 10,
		Delta:         2.0,
		StartNumber:   1.0,
		Interval:      1.0,
	}

	mockService.On("SetTask", expectedModel).Return(nil).Once()

	body, _ := json.Marshal(expectedModel)

	req := httptest.NewRequest(http.MethodPost, "/task/set", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-req-id")
	req = req.WithContext(ctx)

	h.SetTask(rec, req)

	var response models.Response
	err := json.Unmarshal(rec.Body.Bytes(), &response)

	assert.NoError(t, err, "Failed to unmarshal error body")
	assert.Equal(t, "ok", response.Status, "Expected status 'ok'")
	assert.Empty(t, response.Error, "Expected informative error message")

	assert.Equal(t, http.StatusCreated, rec.Code, "Expected HTTP 201 Created status")

	mockService.AssertExpectations(t)
}

func Test_handlers_SetTaskDecodeError(t *testing.T) {
	mockService := new(MockService)
	h := NewHandlers(mockService, slog.Default())

	body := bytes.NewBufferString(`{"elements_count": "not_an_int"}`)

	req := httptest.NewRequest(http.MethodPost, "/task/set", body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-req-id")
	req = req.WithContext(ctx)

	h.SetTask(rec, req)

	var response models.Response
	err := json.Unmarshal(rec.Body.Bytes(), &response)

	assert.NoError(t, err, "Failed to unmarshal error body")
	assert.Equal(t, "error", response.Status, "Expected status 'error'")
	assert.Equal(t, response.Error, "Invalid JSON format", "Expected informative error message")

	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected HTTP 400 Bad Request status")

	mockService.AssertNotCalled(t, "SetTask", mock.Anything)
}

func Test_handlers_SetTaskValidationError(t *testing.T) {
	mockService := new(MockService)
	h := NewHandlers(mockService, slog.Default())

	expectedModel := models.CreateTaskModel{
		ElementsCount: 10,
		Delta:         2.0,
		StartNumber:   1.0,
		Interval:      -1.0,
	}

	body, _ := json.Marshal(expectedModel)

	req := httptest.NewRequest(http.MethodPost, "/task/set", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-req-id")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.SetTask(rec, req)

	var response models.Response
	err := json.Unmarshal(rec.Body.Bytes(), &response)

	assert.NoError(t, err, "Failed to unmarshal error body")
	assert.Equal(t, "error", response.Status, "Expected status 'error'")
	assert.Equal(t, response.Error, "Validation error", "Expected informative error message")

	assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected HTTP 400 Bad Request status")

	mockService.AssertExpectations(t)
}

func Test_handlers_SetTaskInternalError(t *testing.T) {
	mockService := new(MockService)
	h := NewHandlers(mockService, slog.Default())

	expectedModel := models.CreateTaskModel{
		ElementsCount: 10,
		Delta:         2.0,
		StartNumber:   1.0,
		Interval:      1.0,
	}

	mockService.On("SetTask", expectedModel).Return(errors.New("some error")).Once()

	body, _ := json.Marshal(expectedModel)

	req := httptest.NewRequest(http.MethodPost, "/task/set", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-req-id")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	h.SetTask(rec, req)

	var response models.Response
	err := json.Unmarshal(rec.Body.Bytes(), &response)

	assert.NoError(t, err, "Failed to unmarshal error body")
	assert.Equal(t, "error", response.Status, "Expected status 'error'")
	assert.Equal(t, response.Error, "Failed to set task", "Expected informative error message")

	assert.Equal(t, http.StatusInternalServerError, rec.Code, "Expected HTTP 500 Internal Server Error")

	mockService.AssertExpectations(t)
}

func Test_handlers_ListTasks(t *testing.T) {
	mockService := new(MockService)

	h := NewHandlers(mockService, slog.Default())

	expectedTasks := []models.Task{{Id: 1, Status: "COMPLETE"}}

	mockService.On("ListTasks").Return(expectedTasks, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/task/list", nil)
	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-req-id")
	req = req.WithContext(ctx)

	h.ListTasks(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected HTTP 200 Status OK")

	var actualTasks []models.Task
	err := json.Unmarshal(rec.Body.Bytes(), &actualTasks)

	assert.NoError(t, err, "Failed to unmarshal success body")
	assert.Equal(t, expectedTasks, actualTasks, "Returned task list mismatch")

	mockService.AssertExpectations(t)
}

func Test_handlers_ListTasksError(t *testing.T) {
	mockService := new(MockService)

	h := NewHandlers(mockService, slog.Default())

	mockService.On("ListTasks").Return([]models.Task{}, errors.New("some error")).Once()

	req := httptest.NewRequest(http.MethodGet, "/task/list", nil)
	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-req-id")
	req = req.WithContext(ctx)

	h.ListTasks(rec, req)

	var response models.Response
	err := json.Unmarshal(rec.Body.Bytes(), &response)

	assert.NoError(t, err, "Failed to unmarshal error body")
	assert.Equal(t, "error", response.Status, "Expected status 'error'")
	assert.Equal(t, response.Error, "Failed to get tasks", "Expected informative error message")

	assert.Equal(t, http.StatusInternalServerError, rec.Code, "Expected HTTP 500 Internal Server Error")

	mockService.AssertExpectations(t)
}
