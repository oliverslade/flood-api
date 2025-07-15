package service

import (
	"context"
	"time"

	"github.com/oliverslade/flood-api/internal/constants"
	"github.com/oliverslade/flood-api/internal/domain"
	"github.com/oliverslade/flood-api/internal/repository"
)

type RiverService struct {
	repo repository.RiverRepository
}

func NewRiverService(r repository.RiverRepository) *RiverService {
	return &RiverService{repo: r}
}

func (s *RiverService) GetReadings(ctx context.Context, page, pageSize int, startDate time.Time) ([]domain.RiverReading, error) {
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}

	if startDate.IsZero() {
		params := domain.GetReadingsParams{
			Pagination: domain.PaginationParams{Page: page, PageSize: pageSize},
			StartDate:  nil,
		}
		return s.repo.GetReadings(ctx, params)
	} else {
		params := domain.GetReadingsParams{
			Pagination: domain.PaginationParams{Page: page, PageSize: pageSize},
			StartDate:  &startDate,
		}
		return s.repo.GetReadings(ctx, params)
	}
}
