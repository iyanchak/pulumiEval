package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHitsEndpoint(t *testing.T) {
	// Setup temporary SQLite DB
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Initialize(); err != nil {
		t.Fatalf("Initialize error: %v", err)
	}

	// Create mux as in main.go
	mux := NewMux(db)

	// Test GET (should be 0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/hits", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET expected status 200, got %d", rec.Code)
	}
	var resp struct {
		Hits int `json:"hits"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode GET response: %v", err)
	}
	if resp.Hits != 0 {
		t.Fatalf("Expected hits 0, got %d", resp.Hits)
	}

	// Test POST (increment)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/hits", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST expected status 204, got %d", rec.Code)
	}

	// Test GET after increment (should be 1)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/hits", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET after POST expected status 200, got %d", rec.Code)
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode GET after POST response: %v", err)
	}
	if resp.Hits != 1 {
		t.Fatalf("Expected hits 1 after increment, got %d", resp.Hits)
	}

	// Cleanup
	db.db.Close()
	os.Remove(dbPath)
}
