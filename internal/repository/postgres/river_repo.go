//go:generate sqlc generate

package postgres

import (
	"context"
	"database/sql"

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
	// Calculate how many records to skip for pagination
	offset := (params.Pagination.Page - 1) * params.Pagination.PageSize

	if params.StartDate != nil {
		queryParams := gen.GetRiverReadingsWithStartDateParams{
			Timestamp: *params.StartDate,
			Limit:     int32(params.Pagination.PageSize),
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
		Limit:  int32(params.Pagination.PageSize),
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
