//go:build integration

package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/oliverslade/flood-api/internal/api"
	postgresrepo "github.com/oliverslade/flood-api/internal/repository/postgres"
	"github.com/oliverslade/flood-api/test/integration/testutil"
)

// BenchmarkIntegrationAPI measures real HTTP+DB performance through public API
// Reports: ns/op (nanoseconds per operation), B/op (bytes allocated), allocs/op (allocations per operation)
func BenchmarkIntegrationAPI(b *testing.B) {
	if testDB == nil {
		b.Skip("TestMain not run")
	}
	
	ctx := context.Background()
	
	// Seed data once outside measurement
	seedBenchmarkData(b, testDB)
	
	// Create test server (same as production setup)
	server := createBenchmarkServer(b)
	defer server.Close()
	
	benchmarks := []struct {
		name string
		path string
	}{
		{"River_DefaultPage", "/river"},
		{"River_LargePage", "/river?pagesize=100"},
		{"River_WithDateFilter", "/river?start=2024-01-01&pagesize=50"},
		{"River_LastPage", "/river?page=1000&pagesize=10"},
		{"Rainfall_DefaultPage", "/rainfall/benchmark-station"},
		{"Rainfall_LargePage", "/rainfall/benchmark-station?pagesize=100"},
		{"Rainfall_WithDateFilter", "/rainfall/benchmark-station?start=2024-01-01&pagesize=50"},
	}
	
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			url := server.URL + bm.path
			b.ReportAllocs() // Report memory allocations per HTTP request
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				resp := mustGETBench(b, ctx, url)
				// Basic validation - don't fail on LastPage having 0 results
				if len(resp.Readings) == 0 && bm.name != "River_LastPage" {
					b.Fatal("Expected readings")
				}
			}
		})
	}
}

// createBenchmarkServer sets up HTTP server for benchmarks
func createBenchmarkServer(b *testing.B) *httptest.Server {
	b.Helper()
	
	riverRepo := postgresrepo.NewRiverRepo(testDB)
	riverHandler := api.NewRiverHandler(riverRepo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	
	rainfallRepo := postgresrepo.NewRainfallRepo(testDB)
	rainfallHandler := api.NewRainfallHandler(rainfallRepo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	
	router := chi.NewRouter()
	router.Use(api.TimeoutMiddleware)
	router.Get("/river", riverHandler.GetReadings)
	router.Get("/rainfall/{station}", rainfallHandler.GetReadingsByStation)
	
	return httptest.NewServer(router)
}

// mustGETBench is a lightweight GET helper for benchmarks
func mustGETBench(b *testing.B, ctx context.Context, url string) testutil.APIResponse {
	b.Helper()
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		b.Fatal(err)
	}
	
	resp, err := testutil.HTTPClient.Do(req)
	if err != nil {
		b.Fatal(err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		b.Fatalf("Expected 200, got %d for %s", resp.StatusCode, url)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		b.Fatal(err)
	}
	
	var result testutil.APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		b.Fatal(err)
	}
	
	return result
}

// seedBenchmarkData creates enough data for meaningful benchmarks
func seedBenchmarkData(b *testing.B, db *sql.DB) {
	b.Helper()
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Check if already seeded
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM riverlevels").Scan(&count)
	if err == nil && count > 10000 {
		return // Already seeded
	}
	
	// Clean first to avoid conflicts
	cleanDB(b, db)
	
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		b.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()
	
	// Insert benchmark station (simple INSERT, no ON CONFLICT for test schema)
	_, err = tx.ExecContext(ctx,
		"INSERT INTO stationnames (id, name) VALUES ($1, $2)",
		"BENCH001", "benchmark-station",
	)
	if err != nil {
		b.Fatalf("Failed to insert station: %v", err)
	}
	
	// Batch insert 10k river readings
	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	for batch := 0; batch < 100; batch++ {
		query := "INSERT INTO riverlevels (timestamp, level) VALUES "
		args := make([]interface{}, 0, 200)
		
		for i := 0; i < 100; i++ {
			idx := batch*100 + i
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
			args = append(args,
				baseTime.Add(time.Duration(idx)*time.Hour).Format("2006-01-02T15:04:05"),
				float64(idx%100)/10.0,
			)
		}
		
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			b.Fatalf("Failed to insert river batch %d: %v", batch, err)
		}
	}
	
	// Batch insert 10k rainfall readings
	for batch := 0; batch < 100; batch++ {
		query := "INSERT INTO rainfalls (stationid, timestamp, level) VALUES "
		args := make([]interface{}, 0, 300)
		
		for i := 0; i < 100; i++ {
			idx := batch*100 + i
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
			args = append(args,
				"BENCH001",
				baseTime.Add(time.Duration(idx)*time.Hour).Format("2006-01-02T15:04:05"),
				float64(idx%50)/10.0,
			)
		}
		
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			b.Fatalf("Failed to insert rainfall batch %d: %v", batch, err)
		}
	}
	
	if err := tx.Commit(); err != nil {
		b.Fatalf("Failed to commit benchmark data: %v", err)
	}
} 