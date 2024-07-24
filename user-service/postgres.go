package main

import (
	"database/sql"
	"embed"
	"fmt"
	_ "github.com/lib/pq"
	"io/fs"
	"sort"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func NewPostgresDB(databaseConfig DatabaseConfig) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		databaseConfig.Host, databaseConfig.Port,
		databaseConfig.User, databaseConfig.Password, databaseConfig.Name)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Successfully connected to the database!")

	// Run migrations
	fmt.Println("Migrating the database schema")
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}
	fmt.Println("Successfully applied migrations")
	return db, nil
}

func migrate(db *sql.DB) error {
	// Ensure the 'migrations' table exists so we don't duplicate migrations.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS migrations (name TEXT PRIMARY KEY);`); err != nil {
		return fmt.Errorf("cannot create migrations table: %w", err)
	}

	// Read migration files from our embedded file system.
	// This uses Go 1.16's 'embed' package.
	names, err := fs.Glob(migrationFS, "migrations/*.sql")
	if err != nil {
		return err
	}

	if len(names) == 0 {
		fmt.Println("No migration files found.")
		return nil
	}
	fmt.Println("Found migration files:", names)
	sort.Strings(names)

	// Loop over all migration files and execute them in order.
	for _, name := range names {
		if err := migrateFile(db, name); err != nil {
			return fmt.Errorf("migration error: name=%q err=%w", name, err)
		}
	}
	return nil
}

// migrate runs a single migration file within a transaction. On success, the
// migration file name is saved to the "migrations" table to prevent re-running.
func migrateFile(db *sql.DB, name string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Ensure migration has not already been run.
	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM migrations WHERE name = $1`, name).Scan(&n); err != nil {
		return err
	} else if n != 0 {
		return nil // already run migration, skip
	}

	// Read and execute migration file.
	if buf, err := fs.ReadFile(migrationFS, name); err != nil {
		return err
	} else if _, err := tx.Exec(string(buf)); err != nil {
		return err
	}

	// Insert record into migrations to prevent re-running migration.
	if _, err := tx.Exec(`INSERT INTO migrations (name) VALUES ($1)`, name); err != nil {
		return err
	}

	return tx.Commit()
}
