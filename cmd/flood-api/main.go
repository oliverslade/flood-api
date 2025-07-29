// cmd/flood-api/main.go
package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"

	"github.com/oliverslade/flood-api/internal/api"
	"github.com/oliverslade/flood-api/internal/repository/postgres"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	port := flag.String("port", "9001", "TCP port to listen on")
	flag.Parse()
	addr := ":" + *port

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("db open", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Connection pooling
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute) 
	db.SetConnMaxIdleTime(30 * time.Second)

	if err := db.Ping(); err != nil {
		slog.Error("db ping", "err", err)
		os.Exit(1)
	}

	riverRepo := postgres.NewRiverRepo(db)
	riverHandler := api.NewRiverHandler(riverRepo, slog.Default())

	rainfallRepo := postgres.NewRainfallRepo(db)
	rainfallHandler := api.NewRainfallHandler(rainfallRepo, slog.Default())

	router := chi.NewRouter()
	router.Use(api.TimeoutMiddleware) // Add 5s timeout to all requests
	router.Get("/river", riverHandler.GetReadings)
	router.Get("/rainfall/{station}", rainfallHandler.GetReadingsByStation)

	slog.Info("Listening", "addr", addr)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("listen", "err", err)
	}
}
