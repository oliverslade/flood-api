package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/oliverslade/flood-api/internal/constants"
	"github.com/oliverslade/flood-api/internal/service"
)

type RiverHandler struct {
	service *service.RiverService
	logger  *slog.Logger
}

func NewRiverHandler(svc *service.RiverService, logger *slog.Logger) *RiverHandler {
	return &RiverHandler{
		service: svc,
		logger:  logger,
	}
}

func (h *RiverHandler) GetReadings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var page int
	pageParam := q.Get("page")
	if pageParam == "" {
		page = 1
	} else {
		var err error
		page, err = strconv.Atoi(pageParam)
		if err != nil || page <= 0 {
			h.logger.Warn("Invalid page parameter", "error", err)
			h.returnBadRequest(w, "Page must be a positive integer")
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
		if err != nil || pageSize <= 0 {
			h.logger.Warn("Invalid pageSize parameter", "error", err)
			h.returnBadRequest(w, "Page size must be a positive integer")
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
			h.logger.Warn("Invalid startDate parameter", "error", err)
			h.returnBadRequest(w, "Start date must be in format YYYY-MM-DD")
			return
		}
	}

	readings, err := h.service.GetReadings(r.Context(), page, pageSize, startDate)
	if err != nil {
		h.logger.Error("Error fetching readings", "error", err)
		http.Error(w, "Internal server error when getting readings", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"readings": readings,
	}
	h.writeResponseToJson(w, http.StatusOK, response)
}

func (h *RiverHandler) returnBadRequest(w http.ResponseWriter, msg string) {
	h.logger.Warn("Invalid parameter", "error", msg)
	http.Error(w, msg, http.StatusBadRequest)
}

func (h *RiverHandler) writeResponseToJson(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
