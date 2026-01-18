package infra

import (
	"database/sql"
	"errors"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// OpenDatabase returns a ready-to-use Postgres connection.
func OpenDatabase() (*sql.DB, error) {
	dsn := os.Getenv("WILLIAM_DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/william?sslmode=disable"
	}

	database, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := database.Ping(); err != nil {
		_ = database.Close()
		return nil, err
	}

	return database, nil
}

// RunMigrations applies schema migrations to the database.
func RunMigrations(database *sql.DB) error {
	sourceURL := os.Getenv("WILLIAM_MIGRATIONS")
	if sourceURL == "" {
		sourceURL = "file://db/migrations"
	}

	driver, err := postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return err
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
