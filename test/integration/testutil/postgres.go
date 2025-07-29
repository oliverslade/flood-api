package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	sharedContainer testcontainers.Container
	sharedDB        *sql.DB
	sharedConnStr   string // Store connection string for migrations
	containerMu     sync.Mutex
	initOnce        sync.Once
)

// GetTestDB returns a database connection for integration tests.
// First call starts the container; subsequent calls reuse it.
func GetTestDB(t testing.TB) *sql.DB {
	t.Helper()

	var initErr error
	initOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		initErr = initialiseContainer(ctx)
	})

	if initErr != nil {
		t.Fatalf("Failed to initialize test container: %v", initErr)
	}

	return sharedDB
}

func GetTestDBConnString() string {
	return sharedConnStr
}

func initialiseContainer(ctx context.Context) error {
	containerMu.Lock()
	defer containerMu.Unlock()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:17-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
				return fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
			}),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	sharedContainer = pgContainer
	sharedDB = db
	sharedConnStr = connStr

	return nil
}

func Cleanup() {
	if sharedDB != nil {
		sharedDB.Close()
	}

	if sharedContainer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = sharedContainer.Terminate(ctx)
	}
}
