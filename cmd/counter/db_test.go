package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestDBInitializeAndIncrement(t *testing.T) {
	// Use in‑memory SQLite database
	db, err := NewDB(":memory:")
	assert.NoError(t, err)
	defer db.db.Close()

	// Initialize should create table and insert initial row
	err = db.Initialize()
	assert.NoError(t, err)

	// Increment the counter
	err = db.Increment()
	assert.NoError(t, err)

	// Verify count is 1
	var count int
	row := db.db.QueryRow(`SELECT count FROM hits`)
	err = row.Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

// Test GetCount method directly
func TestDBGetCount(t *testing.T) {
	db, err := NewDB(":memory:")
	assert.NoError(t, err)
	defer db.db.Close()

	// Initialize DB
	err = db.Initialize()
	assert.NoError(t, err)

	// GetCount should return 0 initially
	count, err := db.GetCount()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	// Increment and check again
	err = db.Increment()
	assert.NoError(t, err)
	count, err = db.GetCount()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}
