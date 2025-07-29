package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oliverslade/flood-api/internal/domain"
	"github.com/oliverslade/flood-api/internal/repository"
)

type RainfallHandler struct {
	repo   repository.RainfallRepository
	logger *slog.Logger
}

func NewRainfallHandler(repo repository.RainfallRepository, logger *slog.Logger) *RainfallHandler {
	return &RainfallHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *RainfallHandler) GetReadingsByStation(w http.ResponseWriter, r *http.Request) {
	stationName := chi.URLParam(r, "station")

	pagination, errMsg := ParsePaginationParams(r)
	if errMsg != "" {
		h.logger.Warn("Invalid pagination params", "error", errMsg)
		h.returnBadRequest(w, errMsg)
		return
	}

	startDate, errMsg := ParseStartDate(r)
	if errMsg != "" {
		h.logger.Warn("Invalid start date", "error", errMsg)
		h.returnBadRequest(w, errMsg)
		return
	}

	params := domain.GetRainfallParams{
		StationName: stationName,
		GetReadingsParams: domain.GetReadingsParams{
			Pagination: pagination,
			StartDate:  startDate,
		},
	}

	readings, err := h.repo.GetReadingsByStation(r.Context(), params)
	if err != nil {
		if err == domain.ErrNotFound {
			h.logger.Warn("Station not found", "station", stationName)
			http.Error(w, "Station not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Error fetching readings", "error", err)
		http.Error(w, "Internal server error when getting readings", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"readings": readings,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Error encoding response", "error", err)
	}
}

func (h *RainfallHandler) returnBadRequest(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
