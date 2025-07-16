package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/oliverslade/flood-api/internal/repository/inmemory"
	"github.com/oliverslade/flood-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRainfallHandler_GetReadingsByStation(t *testing.T) {

	t.Run("returns readings successfully with default parameters", func(t *testing.T) {
		repo := inmemory.NewRainfallRepo()
		service := service.NewRainfallService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(service, logger)

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
				Level     float64 `json:"level"`
				Station   string  `json:"station"`
			} `json:"readings"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Len(t, response.Readings, 3)
		assert.Equal(t, "2024-01-01T09:00:00", response.Readings[0].Timestamp)
		assert.Equal(t, 2.1, response.Readings[0].Level)
		assert.Equal(t, "catcleugh", response.Readings[0].Station)
		assert.Equal(t, "2024-01-02T10:00:00", response.Readings[1].Timestamp)
		assert.Equal(t, 2.2, response.Readings[1].Level)
		assert.Equal(t, "catcleugh", response.Readings[1].Station)
		assert.Equal(t, "2024-01-03T11:00:00", response.Readings[2].Timestamp)
		assert.Equal(t, 2.3, response.Readings[2].Level)
		assert.Equal(t, "catcleugh", response.Readings[2].Station)
	})

	t.Run("returns readings successfully with page and pageSize parameters", func(t *testing.T) {
		repo := inmemory.NewRainfallRepo()
		service := service.NewRainfallService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/rainfall/{station}", handler.GetReadingsByStation)

		req, err := http.NewRequest("GET", "/rainfall/catcleugh?page=1&pagesize=20", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response struct {
			Readings []struct {
				Timestamp string  `json:"timestamp"`
				Level     float64 `json:"level"`
				Station   string  `json:"station"`
			} `json:"readings"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Len(t, response.Readings, 3)
		assert.Equal(t, "2024-01-01T09:00:00", response.Readings[0].Timestamp)
		assert.Equal(t, 2.1, response.Readings[0].Level)
		assert.Equal(t, "catcleugh", response.Readings[0].Station)
	})

	t.Run("returns readings successfully with start date filter", func(t *testing.T) {
		repo := inmemory.NewRainfallRepo()
		service := service.NewRainfallService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/rainfall/{station}", handler.GetReadingsByStation)

		req, err := http.NewRequest("GET", "/rainfall/catcleugh?start=2024-01-02", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response struct {
			Readings []struct {
				Timestamp string  `json:"timestamp"`
				Level     float64 `json:"level"`
				Station   string  `json:"station"`
			} `json:"readings"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Len(t, response.Readings, 2) // Should skip first reading
		assert.Equal(t, "2024-01-02T10:00:00", response.Readings[0].Timestamp)
		assert.Equal(t, 2.2, response.Readings[0].Level)
		assert.Equal(t, "catcleugh", response.Readings[0].Station)
		assert.Equal(t, "2024-01-03T11:00:00", response.Readings[1].Timestamp)
		assert.Equal(t, 2.3, response.Readings[1].Level)
		assert.Equal(t, "catcleugh", response.Readings[1].Station)
	})

	t.Run("returns bad request when start date parameter is invalid", func(t *testing.T) {
		repo := inmemory.NewRainfallRepo()
		service := service.NewRainfallService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/rainfall/{station}", handler.GetReadingsByStation)

		req, err := http.NewRequest("GET", "/rainfall/catcleugh?start=invalid", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Start date must be in format YYYY-MM-DD")
	})

	t.Run("returns bad request when page parameter is invalid", func(t *testing.T) {
		repo := inmemory.NewRainfallRepo()
		service := service.NewRainfallService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/rainfall/{station}", handler.GetReadingsByStation)

		req, err := http.NewRequest("GET", "/rainfall/catcleugh?page=invalid", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Page must be a positive integer")
	})

	t.Run("returns bad request when page parameter is below one", func(t *testing.T) {
		repo := inmemory.NewRainfallRepo()
		service := service.NewRainfallService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/rainfall/{station}", handler.GetReadingsByStation)

		req, err := http.NewRequest("GET", "/rainfall/catcleugh?page=0", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Page must be a positive integer")
	})

	t.Run("returns bad request when pagesize parameter is invalid", func(t *testing.T) {
		repo := inmemory.NewRainfallRepo()
		service := service.NewRainfallService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/rainfall/{station}", handler.GetReadingsByStation)

		req, err := http.NewRequest("GET", "/rainfall/catcleugh?pagesize=invalid", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Page size must be a positive integer")
	})

	t.Run("returns bad request when pagesize parameter is below one", func(t *testing.T) {
		repo := inmemory.NewRainfallRepo()
		service := service.NewRainfallService(repo)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewRainfallHandler(service, logger)

		router := chi.NewRouter()
		router.Get("/rainfall/{station}", handler.GetReadingsByStation)

		req, err := http.NewRequest("GET", "/rainfall/catcleugh?pagesize=0", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Page size must be a positive integer")
	})
}
