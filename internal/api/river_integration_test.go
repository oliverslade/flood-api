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
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/oliverslade/flood-api/internal/domain"
	postgresrepo "github.com/oliverslade/flood-api/internal/repository/postgres"
	"github.com/oliverslade/flood-api/internal/service"
)

//go:embed testdata/migrations/*.sql
var migrationFiles embed.FS

// the goal was to make the integration tests thin, explicit, and fast enough for CI.
// This test spins up a disposable container, applies real migrations, and hits the
// actual HTTP handler stack
func TestGetRiverReadings(t *testing.T) {
	t.Parallel()

	//Spin up testcontainer
	db := startTestPostgres(t)

	seedTestData(t, db)
	t.Logf("Test data seeded successfully")

	t.Logf("Wiring up application stack...")
	riverRepo := postgresrepo.NewRiverRepo(db)
	riverService := service.NewRiverService(riverRepo)
	riverHandler := NewRiverHandler(riverService, slog.Default())

	// Created an in memory Chi router (needs reviewing)
	t.Logf("Creating HTTP server...")
	router := chi.NewRouter()
	router.Get("/river", riverHandler.GetReadings)
	server := httptest.NewServer(router)
	defer server.Close()
	t.Logf("HTTP server created at: %s", server.URL)

	testURL := fmt.Sprintf("%s/river?page=1&pagesize=2", server.URL)
	t.Logf("Making HTTP request to: %s", testURL)
	resp, err := http.Get(testURL)
	require.NoError(t, err)
	defer resp.Body.Close()
	t.Logf("HTTP response status: %d", resp.StatusCode)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	t.Logf("Response body: %s", string(body))

	var result struct {
		Readings []domain.RiverReading `json:"readings"`
	}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	require.Len(t, result.Readings, 2)

	require.True(t, result.Readings[0].Timestamp.Before(result.Readings[1].Timestamp))

	require.Equal(t, 1.5, result.Readings[0].Level)
	require.Equal(t, 2.0, result.Readings[1].Level)
}

// spins up a disposable Postgres container and applies migrations.
func startTestPostgres(t testing.TB) *sql.DB {
	ctx := context.Background()

	pgContainer, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:17-alpine"),
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

	// Apply all migrations
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

// inserts exactly 3 river level readings for deterministic tests.
func seedTestData(t testing.TB, db *sql.DB) {
	t.Logf("Starting to seed test data...")
	// Using explicit timestamps to ensure ordering
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	testData := []struct {
		timestamp time.Time
		level     float64
	}{
		{baseTime, 1.5},
		{baseTime.Add(1 * time.Hour), 2.0},
		{baseTime.Add(2 * time.Hour), 2.5},
	}

	for i, data := range testData {
		t.Logf("Inserting row %d: timestamp=%s, level=%f", i+1, data.timestamp.Format(time.RFC3339), data.level)
		_, err := db.Exec(
			"INSERT INTO riverlevels (timestamp, level) VALUES ($1, $2)",
			data.timestamp, data.level,
		)
		require.NoError(t, err)
	}
	t.Logf("Successfully seeded %d rows", len(testData))
}
