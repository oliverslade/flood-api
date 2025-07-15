package service

import (
	"context"
	"testing"
	"time"

	"github.com/oliverslade/flood-api/internal/domain"
	"github.com/oliverslade/flood-api/internal/repository/inmemory"
	"github.com/stretchr/testify/require"
)

func TestRiverService_GetReadings(t *testing.T) {
	ctx := context.Background()
	memRepo := inmemory.NewRiverRepo()
	service := NewRiverService(memRepo)

	// Expected readings for river (all readings in chronological order)
	expectedAllReadings := []domain.RiverReading{
		{Timestamp: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), Level: 1.2},
		{Timestamp: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), Level: 1.3},
		{Timestamp: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC), Level: 1.4},
		{Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), Level: 1.5},
		{Timestamp: time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC), Level: 1.1},
	}

	t.Run("returns two readings for the first page", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 1, 2, time.Time{})
		require.NoError(t, err)
		require.Equal(t, expectedAllReadings[:2], out)
	})

	t.Run("returns the next two readings for the second page", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 2, 2, time.Time{})
		require.NoError(t, err)
		require.Equal(t, expectedAllReadings[2:4], out)
	})

	t.Run("returns the last reading for the third page", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 3, 2, time.Time{})
		require.NoError(t, err)
		require.Equal(t, expectedAllReadings[4:], out)
	})

	t.Run("returns no readings for a page beyond the data", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 10, 2, time.Time{})
		require.NoError(t, err)
		require.Equal(t, []domain.RiverReading{}, out)
	})

	t.Run("returns all readings for a larger page size", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 1, 10, time.Time{})
		require.NoError(t, err)
		require.Equal(t, expectedAllReadings, out)
	})

	t.Run("returns only the readings after a start date", func(t *testing.T) {
		startDate := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
		out, err := service.GetReadings(ctx, 1, 2, startDate)
		require.NoError(t, err)
		require.Equal(t, expectedAllReadings[1:3], out) // Includes start date onwards
	})

	t.Run("clamps page size to max when page size is greater than max", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 1, 10000, time.Time{})
		require.NoError(t, err)
		require.Equal(t, expectedAllReadings, out)
	})
}
