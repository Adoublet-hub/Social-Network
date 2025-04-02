package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type Store interface {
	OpenDatabase() (*sql.DB, error)
	CloseDatabase(db *sql.DB) error
}

type DBStore struct{}

func (s *DBStore) OpenDatabase() (*sql.DB, error) {

	db, err := sql.Open("sqlite3", "pkg/db/data.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping to database: %v", err)
	}
	log.Println("Ping to database")

	// Applique les migrations à l'ouverture de la base de données
	if err := s.ApplyMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %v", &err)
	}

	return db, nil
}

func (s *DBStore) CloseDatabase(db *sql.DB) error {
	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}

// Fonction pour appliquer les migrations à la base de données
func (s *DBStore) ApplyMigrations(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://pkg/db/migrations/sqlite",
		"sqlite3", driver)
	if err != nil {
		return err
	}

	// applique les migrations
	log.Println("Applying migrations...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("Error applying migration: %v", err)
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	log.Println("Migrations applied successfully")

	return nil
}
