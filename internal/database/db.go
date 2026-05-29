package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func Connect(databasePath string) (*DB, error) {
	dir := filepath.Dir(databasePath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", databasePath)

	if err != nil {
		return nil, fmt.Errorf("falied to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

		pragmas := []string {
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA busy_timeout=5000;",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to set pragma %s: %w", pragma, err)
		}
	}

	log.Printf("Connected to database: %s", databasePath)
	return &DB{db}, nil;
}

func (db *DB) RunMigrations(migrationsPath string) error {
	migration, err := os.ReadFile(migrationsPath)

	if err != nil {
		return fmt.Errorf("failed to read migrations file: %w", err)
	}

	// execute all SQL statements in the file
	if _, err := db.Exec(string(migration)); err != nil {
		return fmt.Errorf("failed to run migration: %w", err)
	}

	log.Println("migration completed successfully")
	return nil
}
