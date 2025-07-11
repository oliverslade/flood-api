package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/oliverslade/flood-api/internal/constants"
	"github.com/oliverslade/flood-api/internal/domain"
	"github.com/oliverslade/flood-api/internal/repository"
	"github.com/oliverslade/flood-api/internal/repository/postgres/gen"
)

type RainfallRepo struct {
	queries *gen.Queries
}

func NewRainfallRepo(db *sql.DB) repository.RainfallRepository {
	return &RainfallRepo{
		queries: gen.New(db),
	}
}

// GetReadingsByStation returns rainfall readings for a specific station
func (r *RainfallRepo) GetReadingsByStation(ctx context.Context, params domain.GetRainfallParams) ([]domain.RainfallReading, error) {
	station, err := r.getStationByName(ctx, params.StationName)
	if err != nil {
		return nil, err
	}

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

	// Calculate offset for pagination
	offset := (page - 1) * pageSize

	if params.StartDate != nil {
		queryParams := gen.GetRainfallReadingsByStationWithStartDateParams{
			Stationid: station.ID,
			Timestamp: *params.StartDate,
			Limit:     int32(pageSize),
			Offset:    int32(offset),
		}
		dbReadings, err := r.queries.GetRainfallReadingsByStationWithStartDate(ctx, queryParams)
		if err != nil {
			return nil, err
		}

		readings := make([]domain.RainfallReading, len(dbReadings))
		for i, dbReading := range dbReadings {
			readings[i] = r.toDomainRainfallReading(dbReading.Timestamp, dbReading.Level, params.StationName)
		}
		return readings, nil
	}

	queryParams := gen.GetRainfallReadingsByStationParams{
		Stationid: station.ID,
		Limit:     int32(pageSize),
		Offset:    int32(offset),
	}
	dbReadings, err := r.queries.GetRainfallReadingsByStation(ctx, queryParams)
	if err != nil {
		return nil, err
	}

	readings := make([]domain.RainfallReading, len(dbReadings))
	for i, dbReading := range dbReadings {
		readings[i] = r.toDomainRainfallReading(dbReading.Timestamp, dbReading.Level, params.StationName)
	}
	return readings, nil
}

// getStationByName returns station information by name (internal helper for validation)
func (r *RainfallRepo) getStationByName(ctx context.Context, stationName string) (*domain.Station, error) {
	dbStation, err := r.queries.GetStationByName(ctx, stationName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &domain.Station{
		ID:   dbStation.ID,
		Name: dbStation.Name,
	}, nil
}

// converts a database row to domain rainfall reading
func (r *RainfallRepo) toDomainRainfallReading(timestamp time.Time, level float64, stationName string) domain.RainfallReading {
	return domain.RainfallReading{
		Timestamp:   timestamp,
		Level:       level,
		StationName: stationName,
	}
}
