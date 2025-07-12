package service

import (
	"context"
	"time"

	"github.com/oliverslade/flood-api/internal/constants"
	"github.com/oliverslade/flood-api/internal/domain"
	"github.com/oliverslade/flood-api/internal/repository"
)

type RainfallService struct {
	repo repository.RainfallRepository
}

func NewRainfallService(r repository.RainfallRepository) *RainfallService {
	return &RainfallService{repo: r}
}

func (s *RainfallService) GetReadingsByStation(ctx context.Context, stationName string, page int, pageSize int, start time.Time) ([]domain.RainfallReading, error) {
	// Clamp pagination parameters - this is business logic, not data access logic
	if page < 1 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = constants.DefaultPageSize
	}
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	pagination := domain.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	var readingParams domain.GetReadingsParams
	if start.IsZero() {
		readingParams = domain.GetReadingsParams{
			Pagination: pagination,
			StartDate:  nil,
		}
	} else {
		readingParams = domain.GetReadingsParams{
			Pagination: pagination,
			StartDate:  &start,
		}
	}

	params := domain.GetRainfallParams{
		GetReadingsParams: readingParams,
		StationName:       stationName,
	}

	return s.repo.GetReadingsByStation(ctx, params)
}
