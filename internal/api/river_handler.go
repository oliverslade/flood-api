package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/oliverslade/flood-api/internal/domain"
	"github.com/oliverslade/flood-api/internal/repository"
)

type RiverHandler struct {
	repo   repository.RiverRepository
	logger *slog.Logger
}

func NewRiverHandler(repo repository.RiverRepository, logger *slog.Logger) *RiverHandler {
	return &RiverHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *RiverHandler) GetReadings(w http.ResponseWriter, r *http.Request) {
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

	params := domain.GetReadingsParams{
		Pagination: pagination,
		StartDate:  startDate,
	}

	readings, err := h.repo.GetReadings(r.Context(), params)
	if err != nil {
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

func (h *RiverHandler) returnBadRequest(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
