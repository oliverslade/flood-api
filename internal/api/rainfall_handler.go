package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oliverslade/flood-api/internal/constants"
	"github.com/oliverslade/flood-api/internal/service"
)

type RainfallHandler struct {
	service *service.RainfallService
	logger  *slog.Logger
}

func NewRainfallHandler(service *service.RainfallService, logger *slog.Logger) *RainfallHandler {
	return &RainfallHandler{
		service: service,
		logger:  logger,
	}
}

func (h *RainfallHandler) GetReadingsByStation(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "station")

	q := r.URL.Query()

	var page int
	pageParam := q.Get("page")
	if pageParam == "" {
		page = 0
	} else {
		var err error
		page, err = strconv.Atoi(pageParam)
		if err != nil || page < 0 {
			slog.Warn("Invalid page parameter", "error", err)
			http.Error(w, "Page must be an integer", http.StatusBadRequest)
			return
		}
	}

	var pageSize int
	pageSizeParam := q.Get("pagesize")
	if pageSizeParam == "" {
		pageSize = constants.DefaultPageSize
	} else {
		var err error
		pageSize, err = strconv.Atoi(pageSizeParam)
		if err != nil || pageSize < 0 {
			slog.Warn("Invalid pageSize parameter", "error", err)
			http.Error(w, "Page size must be an integer", http.StatusBadRequest)
			return
		}
	}

	var startDate time.Time
	startDateParam := q.Get("start")
	if startDateParam == "" {
		startDate = time.Time{} // zero value means no date filtering
	} else {
		var err error
		// 2006-01-02 is the reference time format in Go
		startDate, err = time.Parse("2006-01-02", startDateParam)
		if err != nil {
			slog.Warn("Invalid startDate parameter", "error", err)
			http.Error(w, "Start date must be in format YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	readings, err := h.service.GetReadingsByStation(r.Context(), name, page, pageSize, startDate)
	if err != nil {
		h.logger.Error("Error fetching readings", "error", err)
		http.Error(w, "Internal server error when getting readings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"readings": readings,
	}
	json.NewEncoder(w).Encode(response)
}
