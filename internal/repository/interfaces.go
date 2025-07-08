package repository

import (
	"context"
	"time"

	"github.com/oliverslade/flood-api/internal/domain"
)

type RiverRepository interface {
	// returns river level readings with pagination and optional date filtering
	GetReadings(ctx context.Context, params domain.GetReadingsParams) ([]domain.RiverReading, error)

	// returns the total count of river readings for pagination
	GetReadingsCount(ctx context.Context, startDate *time.Time) (int64, error)
}

type RainfallRepository interface {
	// returns rainfall readings for a station name
	GetReadingsByStation(ctx context.Context, params domain.GetRainfallParams) ([]domain.RainfallReading, error)

	// returns the total count of rainfall readings for a station name
	GetReadingsCountByStation(ctx context.Context, stationName string, startDate *time.Time) (int64, error)

	// returns station information by name
	GetStationByName(ctx context.Context, stationName string) (*domain.Station, error)
}
