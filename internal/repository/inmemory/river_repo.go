package inmemory

import (
	"context"
	"time"

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

	offset := (params.Pagination.Page - 1) * params.Pagination.PageSize
	end := offset + params.Pagination.PageSize

	if offset >= len(filtered) {
		return []domain.RiverReading{}, nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], nil
}
