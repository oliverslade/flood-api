package inmemory

import (
	"context"
	"time"

	"github.com/oliverslade/flood-api/internal/domain"
	"github.com/oliverslade/flood-api/internal/repository"
)

// This is an in-memory implementation fake for use with the service layer unit tests
type RainfallRepo struct {
	readings []domain.RainfallReading
	stations map[string]domain.Station
}

func NewRainfallRepo() repository.RainfallRepository {
	// mirrors actual database structure
	readings := []domain.RainfallReading{
		{Timestamp: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), Level: 2.1, StationName: "catcleugh"},
		{Timestamp: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), Level: 2.2, StationName: "catcleugh"},
		{Timestamp: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC), Level: 2.3, StationName: "catcleugh"},
		{Timestamp: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), Level: 1.5, StationName: "haltwhistle"},
		{Timestamp: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), Level: 1.6, StationName: "haltwhistle"},
	}

	// Mirror actual database station mapping (ID -> Name)
	stations := map[string]domain.Station{
		"catcleugh":                 {ID: "010660", Name: "catcleugh"},
		"haltwhistle":               {ID: "014555", Name: "haltwhistle"},
		"hexham-firtrees":           {ID: "016140", Name: "hexham-firtrees"},
		"kielder-ridge-end":         {ID: "008850", Name: "kielder-ridge-end"},
		"chirdon":                   {ID: "010312", Name: "chirdon"},
		"garrigill-noonstones-hill": {ID: "013045", Name: "garrigill-noonstones-hill"},
		"hartside":                  {ID: "013336", Name: "hartside"},
		"alston":                    {ID: "013553", Name: "alston"},
		"knarsdale":                 {ID: "013878", Name: "knarsdale"},
		"acomb-codlaw-hill":         {ID: "015313", Name: "acomb-codlaw-hill"},
		"allenheads-allen-lodge":    {ID: "015347", Name: "allenheads-allen-lodge"},
	}

	return &RainfallRepo{readings: readings, stations: stations}
}

func (r *RainfallRepo) GetReadingsByStation(ctx context.Context, params domain.GetRainfallParams) ([]domain.RainfallReading, error) {
	if _, exists := r.stations[params.StationName]; !exists {
		return nil, domain.ErrNotFound
	}

	var filtered []domain.RainfallReading
	for _, reading := range r.readings {
		if reading.StationName == params.StationName {
			filtered = append(filtered, reading)
		}
	}

	if params.StartDate != nil {
		var filteredByDate []domain.RainfallReading
		for _, reading := range filtered {
			if reading.Timestamp.After(*params.StartDate) || reading.Timestamp.Equal(*params.StartDate) {
				filteredByDate = append(filteredByDate, reading)
			}
		}
		filtered = filteredByDate
	}

	offset := (params.Pagination.Page - 1) * params.Pagination.PageSize
	end := offset + params.Pagination.PageSize

	if offset >= len(filtered) {
		return []domain.RainfallReading{}, nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], nil
}
