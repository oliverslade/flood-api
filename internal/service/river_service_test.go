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

	t.Run("basic pagination - first page", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 1, 2)
		require.NoError(t, err)
		require.Len(t, out, 2)

		require.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 1.2, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 1.3, out[1].Level)
	})

	t.Run("pagination - second page", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 2, 2)
		require.NoError(t, err)
		require.Len(t, out, 2)

		require.Equal(t, time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 1.4, out[0].Level)
		require.Equal(t, time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), out[1].Timestamp)
		require.Equal(t, 1.5, out[1].Level)
	})

	t.Run("pagination - partial last page", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 3, 2)
		require.NoError(t, err)
		require.Len(t, out, 1)

		require.Equal(t, time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC), out[0].Timestamp)
		require.Equal(t, 1.1, out[0].Level)
	})

	t.Run("pagination - page beyond data", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 10, 2)
		require.NoError(t, err)
		require.Len(t, out, 0)
	})

	t.Run("larger page size", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 1, 10)
		require.NoError(t, err)
		require.Len(t, out, 5)

		expectedLevels := []float64{1.2, 1.3, 1.4, 1.5, 1.1}
		for i, reading := range out {
			require.Equal(t, expectedLevels[i], reading.Level)
		}
	})

	t.Run("invalid pagination parameters", func(t *testing.T) {
		out, err := service.GetReadings(ctx, 0, 2)
		require.NoError(t, err)
		require.Len(t, out, 2)

		out, err = service.GetReadings(ctx, 1, 0)
		require.NoError(t, err)
		require.Len(t, out, 5)
	})
}
