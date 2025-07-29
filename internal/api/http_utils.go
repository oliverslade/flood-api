package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/oliverslade/flood-api/internal/constants"
	"github.com/oliverslade/flood-api/internal/domain"
)

// adds a 5s timeout to all requests
func TimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ParsePaginationParams(r *http.Request) (domain.PaginationParams, string) {
	q := r.URL.Query()

	page := 1
	if pageParam := q.Get("page"); pageParam != "" {
		p, err := strconv.Atoi(pageParam)
		if err != nil || p <= 0 {
			return domain.PaginationParams{}, "Page must be a positive integer"
		}
		page = p
	}

	pageSize := constants.DefaultPageSize
	if pageSizeParam := q.Get("pagesize"); pageSizeParam != "" {
		ps, err := strconv.Atoi(pageSizeParam)
		if err != nil || ps <= 0 {
			return domain.PaginationParams{}, "Page size must be a positive integer"
		}
		if ps > constants.MaxPageSize {
			ps = constants.MaxPageSize
		}
		pageSize = ps
	}

	return domain.PaginationParams{Page: page, PageSize: pageSize}, ""
}

func ParseStartDate(r *http.Request) (*time.Time, string) {
	startDateParam := r.URL.Query().Get("start")
	if startDateParam == "" {
		return nil, ""
	}

	startDate, err := time.Parse("2006-01-02", startDateParam)
	if err != nil {
		return nil, "Start date must be in format YYYY-MM-DD"
	}

	return &startDate, ""
}
