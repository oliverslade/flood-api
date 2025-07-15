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

	"github.com/go-chi/chi/v5"
	"github.com/oliverslade/flood-api/internal/domain"
	"github.com/oliverslade/flood-api/internal/repository/inmemory"
	"github.com/oliverslade/flood-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockErrorRepo struct{}

func (m *mockErrorRepo) GetReadings(ctx context.Context, params domain.GetReadingsParams) ([]domain.RiverReading, error) {
	return nil, fmt.Errorf("repository error")
}

func TestRiverHandler_GetReadings(t *testing.T) {
	t.Run("returns readings successfully with default parameters", func(t *testing.T) {
		// Setup
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		// Create test request
		req, err := http.NewRequest("GET", "/river", nil)
		require.NoError(t, err)

		// Create response recorder
		rr := httptest.NewRecorder()

		// Execute through router
		router.ServeHTTP(rr, req)

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

	t.Run("returns readings successfully with both pagination parameters", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river?page=1&pagesize=20", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

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

	t.Run("returns readings successfully with start date filter", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river?start=2024-01-02", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

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

	t.Run("returns bad request when start date parameter is invalid", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river?start=invalid", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Start date must be in format YYYY-MM-DD")
	})

	t.Run("returns bad request when page parameter is invalid", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river?page=invalid", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Page must be an integer")
	})

	t.Run("returns bad request when page size parameter is invalid", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river?pagesize=invalid", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Page size must be an integer")
	})

	t.Run("returns no readings when start date is in the future", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river?start=2025-01-01", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string][]domain.RiverReading
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		readings := response["readings"]
		assert.Len(t, readings, 0)
	})

	t.Run("returns internal server error when repository returns an error", func(t *testing.T) {
		repo := &mockErrorRepo{}
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error when getting readings")
	})
}
