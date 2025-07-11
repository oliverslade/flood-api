package inmemory

import (
	"context"
	"time"

	"github.com/oliverslade/flood-api/internal/constants"
	"github.com/oliverslade/flood-api/internal/domain"
	"github.com/oliverslade/flood-api/internal/repository"
)

// This is an in-memory implementation fake for use with the service layer unit tests
type RiverRepo struct {
	readings []domain.RiverReading
}

func NewRiverRepo() repository.RiverRepository {
	readings := []domain.RiverReading{
		{Timestamp: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), Level: 1.2},
		{Timestamp: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), Level: 1.3},
		{Timestamp: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC), Level: 1.4},
		{Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), Level: 1.5},
		{Timestamp: time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC), Level: 1.1},
	}

	return &RiverRepo{readings: readings}
}

func (r *RiverRepo) GetReadings(ctx context.Context, params domain.GetReadingsParams) ([]domain.RiverReading, error) {
	// Clamp pagination parameters defensively
	page := params.Pagination.Page
	if page < 1 {
		page = 1
	}

	pageSize := params.Pagination.PageSize
	if pageSize <= 0 {
		pageSize = constants.DefaultPageSize
	}
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	var filtered []domain.RiverReading

	if params.StartDate != nil {
		for _, reading := range r.readings {
			if reading.Timestamp.After(*params.StartDate) || reading.Timestamp.Equal(*params.StartDate) {
				filtered = append(filtered, reading)
			}
		}
	} else {
		filtered = r.readings
	}

	offset := (page - 1) * pageSize
	end := offset + pageSize

	if offset >= len(filtered) {
		return []domain.RiverReading{}, nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], nil
}
