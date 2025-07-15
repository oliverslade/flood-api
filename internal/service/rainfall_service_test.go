package service

import (
	"context"
	"testing"
	"time"

	"github.com/oliverslade/flood-api/internal/domain"
	"github.com/oliverslade/flood-api/internal/repository/inmemory"
	"github.com/stretchr/testify/require"
)

func TestRainfallService_ListByStation(t *testing.T) {
	ctx := context.Background()
	memRepo := inmemory.NewRainfallRepo()
	service := NewRainfallService(memRepo)

	// Expected readings for catcleugh station (all readings in chronological order)
	expectedAllReadings := []domain.RainfallReading{
		{Timestamp: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), Level: 2.1, StationName: "catcleugh"},
		{Timestamp: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), Level: 2.2, StationName: "catcleugh"},
		{Timestamp: time.Date(2024, 1, 3, 11, 0, 0, 0, time.UTC), Level: 2.3, StationName: "catcleugh"},
	}

	t.Run("returns all readings for a station, no start date, large page size", func(t *testing.T) {
		out, err := service.ListByStation(ctx, "catcleugh", 1, 10, time.Time{})
		require.NoError(t, err)
		require.Equal(t, expectedAllReadings, out)
	})

	t.Run("returns first two readings for a station when page size is 2", func(t *testing.T) {
		out, err := service.ListByStation(ctx, "catcleugh", 1, 2, time.Time{})
		require.NoError(t, err)
		require.Equal(t, expectedAllReadings[:2], out)
	})

	t.Run("returns remaining readings for a station when page size is 2 and page is 2", func(t *testing.T) {
		out, err := service.ListByStation(ctx, "catcleugh", 2, 2, time.Time{})
		require.NoError(t, err)
		require.Equal(t, expectedAllReadings[2:], out)
	})

	t.Run("returns no readings for a station for a page beyond the data", func(t *testing.T) {
		out, err := service.ListByStation(ctx, "catcleugh", 3, 2, time.Time{})
		require.NoError(t, err)
		require.Equal(t, []domain.RainfallReading{}, out)
	})

	t.Run("returns all readings for a station for a large page size", func(t *testing.T) {
		out, err := service.ListByStation(ctx, "catcleugh", 1, 10, time.Time{})
		require.NoError(t, err)
		require.Equal(t, expectedAllReadings, out)
	})

	t.Run("returns only the readings for a station after a start date", func(t *testing.T) {
		out, err := service.ListByStation(ctx, "catcleugh", 1, 10, time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, expectedAllReadings[1:], out)
	})

	t.Run("clamps page size to max when page size is greater than max", func(t *testing.T) {
		out, err := service.ListByStation(ctx, "catcleugh", 1, 10000, time.Time{})
		require.NoError(t, err)
		require.Equal(t, expectedAllReadings, out)
	})
}
