package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/oliverslade/flood-api/internal/domain"
	"github.com/oliverslade/flood-api/internal/repository/inmemory"
	"github.com/oliverslade/flood-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Add this after the imports and before the test functions
type mockErrorRepo struct{}

func (m *mockErrorRepo) GetReadings(ctx context.Context, params domain.GetReadingsParams) ([]domain.RiverReading, error) {
	return nil, fmt.Errorf("repository error")
}

func TestRiverHandler_GetReadings(t *testing.T) {
	t.Run("successful request with default parameters", func(t *testing.T) {
		// Setup
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		// Create test request
		req, err := http.NewRequest("GET", "/river", nil)
		require.NoError(t, err)

		// Create response recorder
		rr := httptest.NewRecorder()

		// Execute the handler
		handler.GetReadings(rr, req)

		// Assertions
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string][]domain.RiverReading
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		readings := response["readings"]
		assert.Len(t, readings, 5) // The inmemory repo has 5 test readings

		// Verify the readings are sorted chronologically (first reading)
		expectedTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedTime, readings[0].Timestamp)
		assert.Equal(t, 1.2, readings[0].Level)
	})

	t.Run("request with both pagination parameters", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		req, err := http.NewRequest("GET", "/river?page=1&pagesize=20", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		handler.GetReadings(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string][]domain.RiverReading
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		readings := response["readings"]
		assert.Len(t, readings, 5)

		expectedTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedTime, readings[0].Timestamp)
		assert.Equal(t, 1.2, readings[0].Level)
	})

	t.Run("request with start date filter", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		req, err := http.NewRequest("GET", "/river?start=2024-01-02", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		handler.GetReadings(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string][]domain.RiverReading
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		readings := response["readings"]
		assert.Len(t, readings, 1)

		expectedTime := time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedTime, readings[0].Timestamp)
		assert.Equal(t, 1.1, readings[0].Level)
	})

	t.Run("invalid start date format returns 400", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		req, err := http.NewRequest("GET", "/river?start=invalid", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		handler.GetReadings(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Start date must be in format YYYY-MM-DD")
	})

	t.Run("invalid page parameter returns 400", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		req, err := http.NewRequest("GET", "/river?page=invalid", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		handler.GetReadings(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Page must be an integer")
	})

	t.Run("invalid pagesize parameter returns 400", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		req, err := http.NewRequest("GET", "/river?pagesize=invalid", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		handler.GetReadings(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Page size must be an integer")
	})

	t.Run("request with date filter that returns no results", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		req, err := http.NewRequest("GET", "/river?start=2025-01-01", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		handler.GetReadings(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "No river readings found")
	})

	t.Run("return 500 error when something goes wrong", func(t *testing.T) {
		repo := &mockErrorRepo{}
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		req, err := http.NewRequest("GET", "/river", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		handler.GetReadings(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error when getting readings")
	})
}
