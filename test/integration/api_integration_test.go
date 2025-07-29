//go:build integration

package integration

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/stretchr/testify/require"

	"github.com/oliverslade/flood-api/internal/api"
	postgresrepo "github.com/oliverslade/flood-api/internal/repository/postgres"
	"github.com/oliverslade/flood-api/test/integration/testutil"
)

//go:embed testdata/migrations/*.sql
var migrationFiles embed.FS

// Test constants
const (
	testStationID   = "TEST001"
	testStationName = "teststation"
)

var (
	baseTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testDB   *sql.DB
)

// TestMain sets up shared container for all tests
func TestMain(m *testing.M) {
	// Get shared test database
	testDB = testutil.GetTestDB(&testing.T{})
	
	// Apply migrations using embedded files
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := applyTestMigrations(ctx); err != nil {
		slog.Error("Failed to apply migrations", "error", err)
		os.Exit(1)
	}
	
	// Run tests
	code := m.Run()
	
	// Cleanup
	testutil.Cleanup()
	os.Exit(code)
}

func TestFloodAPIIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Create HTTP server with complete API setup (black-box testing)
	server := createTestServer(t)
	defer server.Close()
	
	// Run tests serially to avoid data conflicts
	t.Run("River Endpoints", func(t *testing.T) {
		testRiverEndpoints(t, ctx, server.URL)
	})
	
	t.Run("Rainfall Endpoints", func(t *testing.T) {
		testRainfallEndpoints(t, ctx, server.URL)
	})
}

// createTestServer sets up a complete HTTP server for black-box testing
func createTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	
	// Clean and seed data
	cleanDB(t, testDB)
	seedTestData(t, testDB, testStationID, testStationName, baseTime)
	
	// Create handlers - this is the only place we touch internal packages
	riverRepo := postgresrepo.NewRiverRepo(testDB)
	riverHandler := api.NewRiverHandler(riverRepo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	
	rainfallRepo := postgresrepo.NewRainfallRepo(testDB)
	rainfallHandler := api.NewRainfallHandler(rainfallRepo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	
	// Setup router exactly like production
	router := chi.NewRouter()
	router.Use(api.TimeoutMiddleware)
	router.Get("/river", riverHandler.GetReadings)
	router.Get("/rainfall/{station}", rainfallHandler.GetReadingsByStation)
	
	return httptest.NewServer(router)
}

func testRiverEndpoints(t *testing.T, ctx context.Context, baseURL string) {
	t.Run("basic functionality", func(t *testing.T) {
		result := testutil.MustGET(t, ctx, fmt.Sprintf("%s/river", baseURL))
		require.Len(t, result.Readings, 3)
		
		expected := []testutil.Reading{
			{Timestamp: "2024-01-01T00:00:00Z", Level: 1.5},
			{Timestamp: "2024-01-01T01:00:00Z", Level: 2.0},
			{Timestamp: "2024-01-01T02:00:00Z", Level: 2.5},
		}
		testutil.AssertReadingsEqual(t, expected, result.Readings)
	})
	
	t.Run("pagination", func(t *testing.T) {
		result := testutil.MustGET(t, ctx, fmt.Sprintf("%s/river?page=1&pagesize=2", baseURL))
		require.Len(t, result.Readings, 2)
		
		expected := []testutil.Reading{
			{Timestamp: "2024-01-01T00:00:00Z", Level: 1.5},
			{Timestamp: "2024-01-01T01:00:00Z", Level: 2.0},
		}
		testutil.AssertReadingsEqual(t, expected, result.Readings)
	})
	
	t.Run("date filtering", func(t *testing.T) {
		result := testutil.MustGET(t, ctx, fmt.Sprintf("%s/river?start=2024-01-01", baseURL))
		require.Len(t, result.Readings, 3)
		
		// All readings should be after or on start date
		for _, r := range result.Readings {
			testutil.ValidateReading(t, r, "")
		}
	})
	
	t.Run("error cases", func(t *testing.T) {
		testCases := []struct {
			name   string
			params string
		}{
			{"Invalid page", "?page=invalid"},
			{"Invalid pagesize", "?pagesize=invalid"},
			{"Invalid start date", "?start=invalid"},
			{"Zero page", "?page=0"},
			{"Negative page", "?page=-1"},
			{"Zero pagesize", "?pagesize=0"},
			{"Negative pagesize", "?pagesize=-1"},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				url := fmt.Sprintf("%s/river%s", baseURL, tc.params)
				testutil.ExpectHTTPError(t, ctx, url, http.StatusBadRequest)
			})
		}
	})
}

func testRainfallEndpoints(t *testing.T, ctx context.Context, baseURL string) {
	t.Run("basic functionality", func(t *testing.T) {
		result := testutil.MustGET(t, ctx, fmt.Sprintf("%s/rainfall/%s", baseURL, testStationName))
		require.Len(t, result.Readings, 3)
		
		expected := []testutil.Reading{
			{Timestamp: "2024-01-01T00:00:00Z", Level: 0.5, Station: testStationName},
			{Timestamp: "2024-01-01T01:00:00Z", Level: 0.8, Station: testStationName},
			{Timestamp: "2024-01-01T02:00:00Z", Level: 1.2, Station: testStationName},
		}
		testutil.AssertReadingsEqual(t, expected, result.Readings)
	})
	
	t.Run("pagination", func(t *testing.T) {
		result := testutil.MustGET(t, ctx, fmt.Sprintf("%s/rainfall/%s?page=1&pagesize=2", baseURL, testStationName))
		require.Len(t, result.Readings, 2)
		
		for _, r := range result.Readings {
			testutil.ValidateReading(t, r, testStationName)
		}
	})
	
	t.Run("date filtering", func(t *testing.T) {
		result := testutil.MustGET(t, ctx, fmt.Sprintf("%s/rainfall/%s?start=2024-01-01", baseURL, testStationName))
		require.Len(t, result.Readings, 3)
		
		for _, r := range result.Readings {
			testutil.ValidateReading(t, r, testStationName)
		}
	})
	
	t.Run("error cases", func(t *testing.T) {
		testCases := []struct {
			name   string
			params string
		}{
			{"Invalid page", "?page=invalid"},
			{"Invalid pagesize", "?pagesize=invalid"},
			{"Invalid start date", "?start=invalid"},
			{"Zero page", "?page=0"},
			{"Negative page", "?page=-1"},
			{"Zero pagesize", "?pagesize=0"},
			{"Negative pagesize", "?pagesize=-1"},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				url := fmt.Sprintf("%s/rainfall/%s%s", baseURL, testStationName, tc.params)
				testutil.ExpectHTTPError(t, ctx, url, http.StatusBadRequest)
			})
		}
	})
	
	t.Run("non-existent station", func(t *testing.T) {
		url := fmt.Sprintf("%s/rainfall/non-existent", baseURL)
		testutil.ExpectHTTPError(t, ctx, url, http.StatusNotFound)
	})
}

// applyTestMigrations runs migrations on the test database
func applyTestMigrations(ctx context.Context) error {
	// Get connection string from shared test infrastructure
	connStr := testutil.GetTestDBConnString()
	
	// Apply migrations from embedded files
	source, err := iofs.New(migrationFiles, "testdata/migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}
	
	m, err := migrate.NewWithSourceInstance("iofs", source, connStr)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()
	
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	
	return nil
}

// cleanDB removes all test data for consistent test state
func cleanDB(t testing.TB, db *sql.DB) {
	t.Helper()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Clean in reverse dependency order
	_, err := db.ExecContext(ctx, "DELETE FROM rainfalls")
	require.NoError(t, err)
	
	_, err = db.ExecContext(ctx, "DELETE FROM riverlevels")
	require.NoError(t, err)
	
	_, err = db.ExecContext(ctx, "DELETE FROM stationnames")
	require.NoError(t, err)
}

// seedTestData inserts consistent test data
func seedTestData(t testing.TB, db *sql.DB, stationID, stationName string, baseTime time.Time) {
	t.Helper()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := db.ExecContext(ctx,
		"INSERT INTO stationnames (id, name) VALUES ($1, $2)",
		stationID, stationName,
	)
	require.NoError(t, err)
	
	// Batch insert river levels
	_, err = db.ExecContext(ctx, `
		INSERT INTO riverlevels (timestamp, level) VALUES 
		($1, 1.5), ($2, 2.0), ($3, 2.5)`,
		baseTime,
		baseTime.Add(1*time.Hour),
		baseTime.Add(2*time.Hour),
	)
	require.NoError(t, err)
	
	// Batch insert rainfall data
	_, err = db.ExecContext(ctx, `
		INSERT INTO rainfalls (stationid, timestamp, level) VALUES 
		($1, $2, 0.5), ($3, $4, 0.8), ($5, $6, 1.2)`,
		stationID, baseTime,
		stationID, baseTime.Add(1*time.Hour),
		stationID, baseTime.Add(2*time.Hour),
	)
	require.NoError(t, err)
} 