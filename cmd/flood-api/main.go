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
	"github.com/oliverslade/flood-api/internal/service"
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

	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(30 * time.Minute)
	if err := db.Ping(); err != nil {
		slog.Error("db ping", "err", err)
		os.Exit(1)
	}

	riverRepo := postgres.NewRiverRepo(db)
	riverService := service.NewRiverService(riverRepo)
	riverHandler := api.NewRiverHandler(riverService, slog.Default())

	router := chi.NewRouter()
	router.Get("/river", riverHandler.GetReadings)

	slog.Info("Listening", "addr", addr)
	server := &http.Server{Addr: addr, Handler: router}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("listen", "err", err)
	}
}
