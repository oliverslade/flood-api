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

type mockRainfallErrorRepo struct{}

func (m *mockRainfallErrorRepo) GetReadingsByStation(ctx context.Context, params domain.GetRainfallParams) ([]domain.RainfallReading, error) {
	return nil, fmt.Errorf("repository error")
}

func TestRainfallHandler_GetReadingsByStation(t *testing.T) {
	t.Run("returns readings successfully for valid station", func(t *testing.T) {
		repo := inmemory.NewRainfallRepo()
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(repo, logger)

		router := chi.NewRouter()
		router.Get("/rainfall/{station}", handler.GetReadingsByStation)

		req, err := http.NewRequest("GET", "/rainfall/catcleugh", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response struct {
			Readings []struct {
				Timestamp string  `json:"timestamp"`
				Station   string  `json:"station"`
				Level     float64 `json:"level"`
			} `json:"readings"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Len(t, response.Readings, 3)
		assert.Equal(t, "2024-01-01T09:00:00Z", response.Readings[0].Timestamp)
		assert.Equal(t, "catcleugh", response.Readings[0].Station)
		assert.Equal(t, 2.1, response.Readings[0].Level)
	})

	t.Run("returns 404 for non-existent station", func(t *testing.T) {
		repo := inmemory.NewRainfallRepo()
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(repo, logger)

		router := chi.NewRouter()
		router.Get("/rainfall/{station}", handler.GetReadingsByStation)

		req, err := http.NewRequest("GET", "/rainfall/non-existent", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Station not found")
	})

	t.Run("returns paginated readings", func(t *testing.T) {
		repo := inmemory.NewRainfallRepo()
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(repo, logger)

		router := chi.NewRouter()
		router.Get("/rainfall/{station}", handler.GetReadingsByStation)

		req, err := http.NewRequest("GET", "/rainfall/catcleugh?page=2&pagesize=1", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response struct {
			Readings []struct {
				Timestamp string  `json:"timestamp"`
				Station   string  `json:"station"`
				Level     float64 `json:"level"`
			} `json:"readings"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Len(t, response.Readings, 1)
		assert.Equal(t, "2024-01-02T10:00:00Z", response.Readings[0].Timestamp)
		assert.Equal(t, 2.2, response.Readings[0].Level)
	})

	t.Run("validates invalid page parameter", func(t *testing.T) {
		repo := inmemory.NewRainfallRepo()
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(repo, logger)

		router := chi.NewRouter()
		router.Get("/rainfall/{station}", handler.GetReadingsByStation)

		req, err := http.NewRequest("GET", "/rainfall/catcleugh?page=-1", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("validates invalid start date format", func(t *testing.T) {
		repo := inmemory.NewRainfallRepo()
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(repo, logger)

		router := chi.NewRouter()
		router.Get("/rainfall/{station}", handler.GetReadingsByStation)

		req, err := http.NewRequest("GET", "/rainfall/catcleugh?start=invalid", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("handles repository errors gracefully", func(t *testing.T) {
		repo := &mockRainfallErrorRepo{}
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(repo, logger)

		router := chi.NewRouter()
		router.Get("/rainfall/{station}", handler.GetReadingsByStation)

		req, err := http.NewRequest("GET", "/rainfall/catcleugh", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error")
	})
}
