package repository

import (
	"context"

	"github.com/oliverslade/flood-api/internal/domain"
)

type RiverRepository interface {
	// returns river level readings with pagination and optional date filtering
	GetReadings(ctx context.Context, params domain.GetReadingsParams) ([]domain.RiverReading, error)
}

type RainfallRepository interface {
	// returns rainfall readings for a station name
	GetReadingsByStation(ctx context.Context, params domain.GetRainfallParams) ([]domain.RainfallReading, error)
}
