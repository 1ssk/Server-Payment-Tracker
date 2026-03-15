package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/anyx/serversanyx-backend/internal/mailer"
	appServer "github.com/anyx/serversanyx-backend/internal/server"
	"github.com/anyx/serversanyx-backend/internal/storage"
)

// main — точка входа для бинарника сервера.
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

	go runReminderLoop(store)

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

// runReminderLoop раз в час проверяет сервера с предстоящим платежом и отправляет напоминания на почту.
func runReminderLoop(store *storage.Store) {
	tick := time.NewTicker(1 * time.Hour)
	defer tick.Stop()
	for range tick.C {
		settings, err := store.GetSMTPSettings()
		if err != nil || !settings.Enabled || settings.Host == "" || settings.To == "" {
			continue
		}
		days := settings.ReminderDaysBefore
		if days <= 0 {
			days = 10
		}
		servers, err := store.ListServersDueForReminder(days)
		if err != nil {
			log.Printf("reminder: list due: %v", err)
			continue
		}
		for _, sv := range servers {
			if err := mailer.SendReminder(settings, settings.To, sv.Name, sv.NextPaymentDate, sv.MonthlyCost, days); err != nil {
				log.Printf("reminder: send %s: %v", sv.Name, err)
				continue
			}
			if err := store.RecordReminderSent(sv.ID, sv.NextPaymentDate); err != nil {
				log.Printf("reminder: record sent: %v", err)
			}
		}
	}
}

