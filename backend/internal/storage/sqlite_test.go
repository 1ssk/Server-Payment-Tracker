package storage

import (
	"os"
	"testing"

	"github.com/anyx/serversanyx-backend/internal/models"
)

// TestStoreBasicCRUD проверяет базовые операции CRUD для серверов и SMTP‑настроек.
func TestStoreBasicCRUD(t *testing.T) {
	tmpFile := "test-db.sqlite"
	_ = os.Remove(tmpFile)
	defer os.Remove(tmpFile)

	store, err := NewStore(tmpFile)
	if err != nil {
		t.Fatalf("NewStore error: %v", err)
	}
	defer store.Close()

	// Создаём сервер.
	server := models.VPNServer{
		ID:              "test-1",
		Name:            "Test Server",
		Provider:        "Test",
		IPAddress:       "127.0.0.1",
		Location:        "Test City",
		MonthlyCost:     1000,
		BillingCycle:    "monthly",
		NextPaymentDate: "2026-01-01",
		Status:          "active",
		CreatedAt:       "2025-01-01",
	}
	if err := store.InsertServer(server); err != nil {
		t.Fatalf("InsertServer error: %v", err)
	}

	servers, err := store.ListServers()
	if err != nil {
		t.Fatalf("ListServers error: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}

	// Обновляем сервер.
	server.Name = "Updated"
	if err := store.UpdateServer(server); err != nil {
		t.Fatalf("UpdateServer error: %v", err)
	}

	servers, err = store.ListServers()
	if err != nil {
		t.Fatalf("ListServers error: %v", err)
	}
	if servers[0].Name != "Updated" {
		t.Fatalf("expected Updated name, got %s", servers[0].Name)
	}

	// Удаляем сервер.
	if err := store.DeleteServer(server.ID); err != nil {
		t.Fatalf("DeleteServer error: %v", err)
	}
	servers, err = store.ListServers()
	if err != nil {
		t.Fatalf("ListServers error: %v", err)
	}
	if len(servers) != 0 {
		t.Fatalf("expected 0 servers, got %d", len(servers))
	}

	// SMTP настройки.
	settings := models.SMTPSettings{
		Host:    "smtp.test",
		Port:    25,
		Enabled: true,
	}
	if err := store.SaveSMTPSettings(settings); err != nil {
		t.Fatalf("SaveSMTPSettings error: %v", err)
	}

	got, err := store.GetSMTPSettings()
	if err != nil {
		t.Fatalf("GetSMTPSettings error: %v", err)
	}
	if got.Host != "smtp.test" || got.Port != 25 || !got.Enabled {
		t.Fatalf("unexpected smtp settings: %+v", got)
	}
}

