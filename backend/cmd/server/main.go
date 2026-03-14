package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	appServer "github.com/anyx/serversanyx-backend/internal/server"
	"github.com/anyx/serversanyx-backend/internal/storage"
)

// main — точка входа для бинарника сервера.
// Настраивает хранилище, HTTP‑маршруты и отдачу статических файлов.
func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/app.db"
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatalf("create db dir: %v", err)
	}

	store, err := storage.NewStore(dbPath)
	if err != nil {
		log.Fatalf("init store: %v", err)
	}
	defer store.Close()

	srv := appServer.New(store)

	mux := http.NewServeMux()
	// API.
	mux.Handle("/api/", srv.Handler())
	// Статика фронтенда (Vite build).
	fs := http.FileServer(http.Dir("frontend"))
	mux.Handle("/", fs)

	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}

	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

