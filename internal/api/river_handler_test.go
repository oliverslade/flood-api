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
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response struct {
			Readings []struct {
				Timestamp string  `json:"timestamp"`
				Level     float64 `json:"level"`
			} `json:"readings"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Len(t, response.Readings, 5)
		assert.Equal(t, "2024-01-01T09:00:00", response.Readings[0].Timestamp)
		assert.Equal(t, 1.2, response.Readings[0].Level)
		assert.Equal(t, "2024-01-01T10:00:00", response.Readings[1].Timestamp)
		assert.Equal(t, 1.3, response.Readings[1].Level)
		assert.Equal(t, "2024-01-01T11:00:00", response.Readings[2].Timestamp)
		assert.Equal(t, 1.4, response.Readings[2].Level)
		assert.Equal(t, "2024-01-01T12:00:00", response.Readings[3].Timestamp)
		assert.Equal(t, 1.5, response.Readings[3].Level)
		assert.Equal(t, "2024-01-02T09:00:00", response.Readings[4].Timestamp)
		assert.Equal(t, 1.1, response.Readings[4].Level)
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

		var response struct {
			Readings []struct {
				Timestamp string  `json:"timestamp"`
				Level     float64 `json:"level"`
			} `json:"readings"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Len(t, response.Readings, 5)
		assert.Equal(t, "2024-01-01T09:00:00", response.Readings[0].Timestamp)
		assert.Equal(t, 1.2, response.Readings[0].Level)
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

		var response struct {
			Readings []struct {
				Timestamp string  `json:"timestamp"`
				Level     float64 `json:"level"`
			} `json:"readings"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Len(t, response.Readings, 1) // Only the last reading from 2024-01-02
		assert.Equal(t, "2024-01-02T09:00:00", response.Readings[0].Timestamp)
		assert.Equal(t, 1.1, response.Readings[0].Level)
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
		assert.Contains(t, rr.Body.String(), "Page must be a positive integer")
	})

	t.Run("returns bad request when page parameter is below one", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river?page=0", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Page must be a positive integer")
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
		assert.Contains(t, rr.Body.String(), "Page size must be a positive integer")
	})

	t.Run("returns bad request when pagesize parameter is below one", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		service := service.NewRiverService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river?pagesize=0", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Page size must be a positive integer")
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
		assert.Equal(t, []domain.RiverReading{}, readings)
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
