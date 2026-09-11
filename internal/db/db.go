package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(databaseDSN string) (*sql.DB, error) {
	return sql.Open("pgx", databaseDSN)
}

//go:embed migrations/*.sql
var migrationFiles embed.FS

func RunMigrations(database *sql.DB, databaseDSN string) error {
	srcDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("error create source driver %v", err)
	}

	_, err = postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("error create db driver %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", srcDriver, databaseDSN)
	if err != nil {
		return fmt.Errorf("error create source %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error migrate %v", err)
	}
	return nil
}
