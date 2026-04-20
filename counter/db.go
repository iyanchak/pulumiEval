package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// DB represents a connection to the SQLite database.
type DB struct {
	db *sql.DB
}

// NewDB returns a new database connection.
func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

// Initialize creates the hits table if it does not exist and inserts a row with count=0 if the table is empty.
func (d *DB) Initialize() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS hits (
			count INTEGER
		);
	`)
	if err != nil {
		return err
	}
	var count int
	row := d.db.QueryRow(`SELECT count FROM hits`)
	err = row.Scan(&count)
	if err == sql.ErrNoRows {
		_, err = d.db.Exec(`INSERT INTO hits (count) VALUES (0)`)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// Increment increments the count in the hits table.
func (d *DB) Increment() error {
	_, err := d.db.Exec(`UPDATE hits SET count = count + 1`)
	return err
}

// GetCount retrieves the current count from the hits table.
func (d *DB) GetCount() (int, error) {
	var count int
	row := d.db.QueryRow(`SELECT count FROM hits`)
	err := row.Scan(&count)
	return count, err
}
