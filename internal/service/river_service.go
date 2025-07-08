package service

import (
	"context"

	"github.com/oliverslade/flood-api/internal/domain"
	"github.com/oliverslade/flood-api/internal/repository"
)

type RiverService struct {
	repo repository.RiverRepository
}

func NewRiverService(r repository.RiverRepository) *RiverService {
	return &RiverService{repo: r}
}

func (s *RiverService) GetReadings(ctx context.Context, page, pageSize int) ([]domain.RiverReading, error) {
	params := domain.GetReadingsParams{
		Pagination: domain.PaginationParams{Page: page, PageSize: pageSize},
		StartDate:  nil,
	}
	return s.repo.GetReadings(ctx, params)
}
