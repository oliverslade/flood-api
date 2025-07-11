//go:generate sqlc generate

package postgres

import (
	"context"
	"database/sql"

	"github.com/oliverslade/flood-api/internal/constants"
	"github.com/oliverslade/flood-api/internal/domain"
	"github.com/oliverslade/flood-api/internal/repository"
	"github.com/oliverslade/flood-api/internal/repository/postgres/gen"
)

type RiverRepo struct {
	queries *gen.Queries
}

func NewRiverRepo(db *sql.DB) repository.RiverRepository {
	return &RiverRepo{
		queries: gen.New(db),
	}
}

// GetReadings returns slice of river level readings with pagination and optional date filtering
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

	// Calculate how many records to skip for pagination
	offset := (page - 1) * pageSize

	if params.StartDate != nil {
		queryParams := gen.GetRiverReadingsWithStartDateParams{
			Timestamp: *params.StartDate,
			Limit:     int32(pageSize),
			Offset:    int32(offset),
		}
		dbReadings, err := r.queries.GetRiverReadingsWithStartDate(ctx, queryParams)
		if err != nil {
			return nil, err
		}

		readings := make([]domain.RiverReading, len(dbReadings))
		for i, dbReading := range dbReadings {
			readings[i] = domain.RiverReading{
				Timestamp: dbReading.Timestamp,
				Level:     dbReading.Level,
			}
		}
		return readings, nil
	}

	queryParams := gen.GetRiverReadingsParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	}
	dbReadings, err := r.queries.GetRiverReadings(ctx, queryParams)
	if err != nil {
		return nil, err
	}

	readings := make([]domain.RiverReading, len(dbReadings))
	for i, dbReading := range dbReadings {
		readings[i] = domain.RiverReading{
			Timestamp: dbReading.Timestamp,
			Level:     dbReading.Level,
		}
	}
	return readings, nil
}
