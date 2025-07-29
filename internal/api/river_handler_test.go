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
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(repo, logger)

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
		assert.Equal(t, "2024-01-01T09:00:00Z", response.Readings[0].Timestamp)
		assert.Equal(t, 1.2, response.Readings[0].Level)
		assert.Equal(t, "2024-01-01T10:00:00Z", response.Readings[1].Timestamp)
		assert.Equal(t, 1.3, response.Readings[1].Level)
		assert.Equal(t, "2024-01-01T11:00:00Z", response.Readings[2].Timestamp)
		assert.Equal(t, 1.4, response.Readings[2].Level)
		assert.Equal(t, "2024-01-01T12:00:00Z", response.Readings[3].Timestamp)
		assert.Equal(t, 1.5, response.Readings[3].Level)
		assert.Equal(t, "2024-01-02T09:00:00Z", response.Readings[4].Timestamp)
		assert.Equal(t, 1.1, response.Readings[4].Level)
	})

	t.Run("returns paginated readings", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(repo, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river?page=2&pagesize=2", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response struct {
			Readings []struct {
				Timestamp string  `json:"timestamp"`
				Level     float64 `json:"level"`
			} `json:"readings"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Len(t, response.Readings, 2)
		assert.Equal(t, "2024-01-01T11:00:00Z", response.Readings[0].Timestamp)
		assert.Equal(t, 1.4, response.Readings[0].Level)
	})

	t.Run("validates invalid page parameter", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(repo, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river?page=0", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("validates invalid start date format", func(t *testing.T) {
		repo := inmemory.NewRiverRepo()
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(repo, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river?start=invalid-date", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("handles repository errors gracefully", func(t *testing.T) {
		repo := &mockErrorRepo{}
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRiverHandler(repo, logger)

		router := chi.NewRouter()
		router.Get("/river", handler.GetReadings)

		req, err := http.NewRequest("GET", "/river", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error")
	})
}
