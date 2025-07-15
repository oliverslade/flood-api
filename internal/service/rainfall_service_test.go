package service

import (
	"context"
	"testing"
	"time"

	"github.com/oliverslade/flood-api/internal/repository/inmemory"
	"github.com/stretchr/testify/require"
)

func TestRainfallService_GetReadingsByStation(t *testing.T) {
	ctx := context.Background()
	memRepo := inmemory.NewRainfallRepo()
	service := NewRainfallService(memRepo)

	t.Run("returns all readings for a station, no start date, large page size", func(t *testing.T) {
		out, err := service.GetReadingsByStation(ctx, "catcleugh", 1, 10, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 3)

		require.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 2.1, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 2.2, out[1].Level)
		require.Equal(t, time.Date(2024, 1, 3, 11, 0, 0, 0, time.UTC), out[2].Timestamp)
		require.Equal(t, 2.3, out[2].Level)
	})
	t.Run("returns first two readings for a station when page size is 2", func(t *testing.T) {
		out, err := service.GetReadingsByStation(ctx, "catcleugh", 1, 2, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 2)

		require.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 2.1, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 2.2, out[1].Level)
	})

	t.Run("returns remaining readings for a station when page size is 2 and page is 2", func(t *testing.T) {
		out, err := service.GetReadingsByStation(ctx, "catcleugh", 2, 2, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 1)

		require.Equal(t, time.Date(2024, 1, 3, 11, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 2.3, out[0].Level)
	})

	t.Run("returns no readings for a station for a page beyond the data", func(t *testing.T) {
		out, err := service.GetReadingsByStation(ctx, "catcleugh", 3, 2, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 0)
	})

	t.Run("returns all readings for a station for a large page size", func(t *testing.T) {
		out, err := service.GetReadingsByStation(ctx, "catcleugh", 1, 10, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 3)

		require.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 2.1, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 2.2, out[1].Level)
		require.Equal(t, time.Date(2024, 1, 3, 11, 0, 0, 0, time.UTC), out[2].Timestamp)
		require.Equal(t, 2.3, out[2].Level)
	})

	t.Run("uses default page size when page size is 0", func(t *testing.T) {
		out, err := service.GetReadingsByStation(ctx, "catcleugh", 1, 0, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 3)

		require.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 2.1, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 2.2, out[1].Level)
		require.Equal(t, time.Date(2024, 1, 3, 11, 0, 0, 0, time.UTC), out[2].Timestamp)
		require.Equal(t, 2.3, out[2].Level)
	})

	t.Run("returns only the readings for a station after a start date", func(t *testing.T) {
		out, err := service.GetReadingsByStation(ctx, "catcleugh", 1, 10, time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Len(t, out, 2)

		require.Equal(t, time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 2.2, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 3, 11, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 2.3, out[1].Level)
	})

	t.Run("sets page to 1 when page is 0", func(t *testing.T) {
		out, err := service.GetReadingsByStation(ctx, "catcleugh", 0, 10, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 3)

		require.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 2.1, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 2.2, out[1].Level)
		require.Equal(t, time.Date(2024, 1, 3, 11, 0, 0, 0, time.UTC), out[2].Timestamp)
		require.Equal(t, 2.3, out[2].Level)
	})

	t.Run("sets page size to default when page size is 0", func(t *testing.T) {
		out, err := service.GetReadingsByStation(ctx, "catcleugh", 1, 0, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 3)

		require.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 2.1, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 2.2, out[1].Level)
		require.Equal(t, time.Date(2024, 1, 3, 11, 0, 0, 0, time.UTC), out[2].Timestamp)
		require.Equal(t, 2.3, out[2].Level)
	})

	t.Run("sets page size to max when page size is greater than max", func(t *testing.T) {
		out, err := service.GetReadingsByStation(ctx, "catcleugh", 1, 10000, time.Time{})
		require.NoError(t, err)
		require.Len(t, out, 3)

		require.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 2.1, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 2.2, out[1].Level)
		require.Equal(t, time.Date(2024, 1, 3, 11, 0, 0, 0, time.UTC), out[2].Timestamp)
		require.Equal(t, 2.3, out[2].Level)
	})
}
