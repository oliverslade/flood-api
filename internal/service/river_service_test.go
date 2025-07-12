package service

import (
	"context"
	"testing"
	"time"

	"github.com/oliverslade/flood-api/internal/repository/inmemory"
	"github.com/stretchr/testify/require"
)

func TestRiverService_GetReadings(t *testing.T) {
	ctx := context.Background()
	memRepo := inmemory.NewRiverRepo()
	service := NewRiverService(memRepo)

	t.Run("returns two readings for the first page", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 1, 2, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 2)

		require.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 1.2, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 1.3, out[1].Level)
	})

	t.Run("returns the next two readings for the second page", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 2, 2, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 2)

		require.Equal(t, time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 1.4, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 1.5, out[1].Level)
	})

	t.Run("returns the last reading for the third page", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 3, 2, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 1)

		require.Equal(t, time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 1.1, out[0].Level)
	})

	t.Run("returns no readings for a page beyond the data", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 10, 2, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 0)
	})

	t.Run("returns all readings for a larger page size", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 1, 10, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 5)

		expectedLevels := []float64{1.2, 1.3, 1.4, 1.5, 1.1}
		for i, reading := range out {
			require.Equal(t, expectedLevels[i], reading.Level)
		}
	})

	t.Run("uses default page size when page size is 0", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 0, 2, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 2)

		require.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 1.2, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 1.3, out[1].Level)
	})

	t.Run("returns only the readings after a start date", func(t *testing.T) {
		startDate := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
		out, err := service.GetReadings(ctx, 1, 2, startDate)
		require.NoError(t, err)
		require.Len(t, out, 2)

		require.Equal(t, time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 1.3, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 1.4, out[1].Level)
	})

	t.Run("sets page to 1 when page is 0", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 0, 10, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 5)

		require.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 1.2, out[0].Level)
	})

	t.Run("sets page size to default when page size is 0", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 1, 0, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 5)

		require.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 1.2, out[0].Level)
	})

	t.Run("sets page size to max when page size is greater than max", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 1, 10000, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 5)

		require.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 1.2, out[0].Level)
	})
}
