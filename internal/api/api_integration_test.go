//go:build integration

package api

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	postgresrepo "github.com/oliverslade/flood-api/internal/repository/postgres"
	"github.com/oliverslade/flood-api/internal/service"
)

//go:embed testdata/migrations/*.sql
var migrationFiles embed.FS

// Pre-compiled regex for timestamp format validation
var timestampPattern = regexp.MustCompile(`^2[0-9]{3}-(0[0-9]|1[0-2])-([0-2][0-9]|3[01])T([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$`)

// Common response types to reduce boilerplate
type apiResponse struct {
	Readings []reading `json:"readings"`
}

type reading struct {
	Timestamp string  `json:"timestamp"`
	Level     float64 `json:"level"`
	Station   string  `json:"station,omitempty"` // omitempty for river readings
}

// Test constants
const (
	timeLayout      = "2006-01-02T15:04:05"
	testStationID   = "TEST001"
	testStationName = "teststation"
)

var (
	baseTime   = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	httpClient = &http.Client{Timeout: 2 * time.Second}
)

// Helper function to make GET requests and parse API responses
func mustGET(t *testing.T, url string) apiResponse {
	t.Helper()

	resp, err := httpClient.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result apiResponse
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	return result
}

// Helper function to make GET requests and expect specific status codes
func makeGETWithStatus(t *testing.T, url string, expectedStatus int) *http.Response {
	t.Helper()

	resp, err := httpClient.Get(url)
	require.NoError(t, err)

	assert.Equal(t, expectedStatus, resp.StatusCode)
	return resp
}

// Helper function to validate timestamp format and parse to time.Time
func validateAndParseTimestamp(t *testing.T, timestamp string) time.Time {
	t.Helper()

	assert.True(t, timestampPattern.MatchString(timestamp), "Timestamp should match OpenAPI pattern: %s", timestamp)

	parsedTime, err := time.Parse(timeLayout, timestamp)
	require.NoError(t, err, "Should be able to parse timestamp: %s", timestamp)

	return parsedTime
}

// Helper function to validate reading data types and constraints
func validateReading(t *testing.T, r reading, expectedStation string) {
	t.Helper()

	validateAndParseTimestamp(t, r.Timestamp)
	assert.GreaterOrEqual(t, r.Level, 0.0, "Level should be >= 0 per OpenAPI spec")

	if expectedStation != "" {
		assert.Equal(t, expectedStation, r.Station, "Station should match expected value")
	}
}

func TestFloodAPIIntegration(t *testing.T) {
	t.Parallel()

	// Spin up test container and seed data
	db := startTestPostgres(t)
	seedTestData(t, db)
	t.Logf("Test database ready with seed data")

	silentLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Wire up application stack
	t.Logf("Wiring up application stack...")
	riverRepo := postgresrepo.NewRiverRepo(db)
	riverService := service.NewRiverService(riverRepo)
	riverHandler := NewRiverHandler(riverService, silentLogger)

	rainfallRepo := postgresrepo.NewRainfallRepo(db)
	rainfallService := service.NewRainfallService(rainfallRepo)
	rainfallHandler := NewRainfallHandler(rainfallService, silentLogger)

	// Create HTTP server with all endpoints
	t.Logf("Creating HTTP server...")
	router := chi.NewRouter()
	router.Get("/river", riverHandler.GetReadings)
	router.Get("/rainfall/{station}", rainfallHandler.GetReadingsByStation)
	server := httptest.NewServer(router)
	defer server.Close()
	t.Logf("HTTP server created at: %s", server.URL)

	t.Run("River Endpoints", func(t *testing.T) {
		testRiverEndpoints(t, server.URL)
	})

	t.Run("Rainfall Endpoints", func(t *testing.T) {
		testRainfallEndpoints(t, server.URL)
	})
}

func testRiverEndpoints(t *testing.T, baseURL string) {
	t.Run("GET /river - basic functionality", func(t *testing.T) {
		result := mustGET(t, fmt.Sprintf("%s/river", baseURL))

		require.Len(t, result.Readings, 3)

		expectedReadings := []struct {
			timeOffset time.Duration
			level      float64
		}{
			{0, 1.5},
			{1 * time.Hour, 2.0},
			{2 * time.Hour, 2.5},
		}

		for i, expected := range expectedReadings {
			expectedTime := baseTime.Add(expected.timeOffset).Format(timeLayout)
			assert.Equal(t, expectedTime, result.Readings[i].Timestamp)
			assert.Equal(t, expected.level, result.Readings[i].Level)
			validateReading(t, result.Readings[i], "")
		}
	})

	t.Run("GET /river - pagination", func(t *testing.T) {
		result := mustGET(t, fmt.Sprintf("%s/river?page=1&pagesize=2", baseURL))

		require.Len(t, result.Readings, 2)

		expectedTime1 := baseTime.Format(timeLayout)
		expectedTime2 := baseTime.Add(1 * time.Hour).Format(timeLayout)

		assert.Equal(t, expectedTime1, result.Readings[0].Timestamp)
		assert.Equal(t, 1.5, result.Readings[0].Level)
		assert.Equal(t, expectedTime2, result.Readings[1].Timestamp)
		assert.Equal(t, 2.0, result.Readings[1].Level)

		for _, reading := range result.Readings {
			validateReading(t, reading, "")
		}
	})

	t.Run("GET /river - date filtering", func(t *testing.T) {
		result := mustGET(t, fmt.Sprintf("%s/river?start=2024-01-01", baseURL))

		require.Len(t, result.Readings, 3)

		for _, reading := range result.Readings {
			validateReading(t, reading, "")
		}
	})

	t.Run("GET /river - error cases", func(t *testing.T) {
		errorCases := []struct {
			name   string
			params string
		}{
			{"Invalid page parameter", "?page=invalid"},
			{"Invalid pagesize parameter", "?pagesize=invalid"},
			{"Invalid start date", "?start=invalid"},
			{"Zero page", "?page=0"},
			{"Negative page", "?page=-1"},
			{"Zero pagesize", "?pagesize=0"},
			{"Negative pagesize", "?pagesize=-1"},
		}

		for _, tc := range errorCases {
			t.Run(tc.name, func(t *testing.T) {
				resp := makeGETWithStatus(t, fmt.Sprintf("%s/river%s", baseURL, tc.params), http.StatusBadRequest)
				resp.Body.Close()
			})
		}
	})
}

func testRainfallEndpoints(t *testing.T, baseURL string) {
	t.Run("GET /rainfall/{station} - basic functionality", func(t *testing.T) {
		result := mustGET(t, fmt.Sprintf("%s/rainfall/%s", baseURL, testStationName))

		require.Len(t, result.Readings, 3)

		expectedReadings := []struct {
			timeOffset time.Duration
			level      float64
		}{
			{0, 0.5},
			{1 * time.Hour, 0.8},
			{2 * time.Hour, 1.2},
		}

		for i, expected := range expectedReadings {
			expectedTime := baseTime.Add(expected.timeOffset).Format(timeLayout)
			assert.Equal(t, expectedTime, result.Readings[i].Timestamp)
			assert.Equal(t, expected.level, result.Readings[i].Level)
			validateReading(t, result.Readings[i], testStationName)
		}
	})

	t.Run("GET /rainfall/{station} - pagination", func(t *testing.T) {
		result := mustGET(t, fmt.Sprintf("%s/rainfall/%s?page=1&pagesize=2", baseURL, testStationName))

		require.Len(t, result.Readings, 2)

		for _, reading := range result.Readings {
			validateReading(t, reading, testStationName)
		}
	})

	t.Run("GET /rainfall/{station} - date filtering", func(t *testing.T) {
		result := mustGET(t, fmt.Sprintf("%s/rainfall/%s?start=2024-01-01", baseURL, testStationName))

		require.Len(t, result.Readings, 3)

		for _, reading := range result.Readings {
			validateReading(t, reading, testStationName)
		}
	})

	t.Run("GET /rainfall/{station} - error cases", func(t *testing.T) {
		errorCases := []struct {
			name   string
			params string
		}{
			{"Invalid page parameter", "?page=invalid"},
			{"Invalid pagesize parameter", "?pagesize=invalid"},
			{"Invalid start date", "?start=invalid"},
			{"Zero page", "?page=0"},
			{"Negative page", "?page=-1"},
			{"Zero pagesize", "?pagesize=0"},
			{"Negative pagesize", "?pagesize=-1"},
		}

		for _, tc := range errorCases {
			t.Run(tc.name, func(t *testing.T) {
				resp := makeGETWithStatus(t, fmt.Sprintf("%s/rainfall/%s%s", baseURL, testStationName, tc.params), http.StatusBadRequest)
				resp.Body.Close()
			})
		}
	})
}

// spins up a disposable Postgres container and applies migrations.
func startTestPostgres(t testing.TB) *sql.DB {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:17-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
				return fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
			}).WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	source, err := iofs.New(migrationFiles, "testdata/migrations")
	require.NoError(t, err)

	m, err := migrate.NewWithSourceInstance("iofs", source, connStr)
	require.NoError(t, err)
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err, "Failed to apply migrations")
	}
	t.Logf("All migrations applied successfully")

	//Cleanup runs even if test panics. No leaked containers.
	t.Cleanup(func() {
		db.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	return db
}

func seedTestData(t testing.TB, db *sql.DB) {
	t.Logf("Starting to seed test data...")

	_, err := db.Exec(
		"INSERT INTO stationnames (id, name) VALUES ($1, $2)",
		testStationID, testStationName,
	)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO riverlevels (timestamp, level) VALUES 
		($1, $2), ($3, $4), ($5, $6)`,
		baseTime, 1.5,
		baseTime.Add(1*time.Hour), 2.0,
		baseTime.Add(2*time.Hour), 2.5,
	)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO rainfalls (stationid, timestamp, level) VALUES 
		($1, $2, $3), ($4, $5, $6), ($7, $8, $9)`,
		testStationID, baseTime, 0.5,
		testStationID, baseTime.Add(1*time.Hour), 0.8,
		testStationID, baseTime.Add(2*time.Hour), 1.2,
	)
	require.NoError(t, err)

	t.Logf("Successfully seeded test data in batches")
}
