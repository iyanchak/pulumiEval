package main

import (
	"encoding/json"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

// NewMux creates an http.ServeMux with the /hits endpoint wired to the provided DB.
func NewMux(db *DB) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/hits", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			count, err := db.GetCount()
			if err != nil {
				http.Error(w, "Failed to get count", http.StatusInternalServerError)
				log.Printf("Error getting count: %v", err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]int{"hits": count})
		case http.MethodPost:
			if err := db.Increment(); err != nil {
				http.Error(w, "Failed to increment count", http.StatusInternalServerError)
				log.Printf("Error incrementing count: %v", err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return mux
}

// GetCount removed; using implementation from db.go

func main() {
	db, err := NewDB("/data/counter.db")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	mux := NewMux(db)

	log.Printf("Server listening on :8080")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
