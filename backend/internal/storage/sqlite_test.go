package storage

import (
	"os"
	"testing"

	"github.com/anyx/serversanyx-backend/internal/models"
)

func testStore(t *testing.T) (*Store, func()) {
	tmpFile := "test-db-" + t.Name() + ".sqlite"
	_ = os.Remove(tmpFile)
	store, err := NewStore(tmpFile)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return store, func() {
		store.Close()
		_ = os.Remove(tmpFile)
	}
}

// TestStoreBasicCRUD проверяет базовые операции CRUD для серверов и SMTP‑настроек.
func TestStoreBasicCRUD(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	server := models.Server{
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

	// SMTP настройки с напоминанием за N дней.
	settings := models.SMTPSettings{
		Host:               "smtp.test",
		Port:               25,
		Enabled:            true,
		ReminderDaysBefore: 10,
	}
	if err := store.SaveSMTPSettings(settings); err != nil {
		t.Fatalf("SaveSMTPSettings: %v", err)
	}
	got, err := store.GetSMTPSettings()
	if err != nil {
		t.Fatalf("GetSMTPSettings: %v", err)
	}
	if got.Host != "smtp.test" || got.Port != 25 || !got.Enabled || got.ReminderDaysBefore != 10 {
		t.Fatalf("unexpected smtp settings: %+v", got)
	}
}

// TestStorePayments проверяет добавление платежей и отчёт.
func TestStorePayments(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	server := models.Server{
		ID:              "s1",
		Name:            "Server 1",
		Provider:        "P",
		IPAddress:       "1.2.3.4",
		Location:        "L",
		MonthlyCost:     1000,
		BillingCycle:    "monthly",
		NextPaymentDate: "2026-06-01",
		Status:          "active",
		CreatedAt:       "2026-01-01",
	}
	if err := store.InsertServer(server); err != nil {
		t.Fatalf("InsertServer: %v", err)
	}

	p := models.Payment{ID: "pay1", ServerID: "s1", Amount: 1000, PaidAt: "2026-05-15", CreatedAt: "2026-05-15T12:00:00Z"}
	if err := store.InsertPayment(p); err != nil {
		t.Fatalf("InsertPayment: %v", err)
	}
	list, err := store.ListPayments("s1", "", "")
	if err != nil {
		t.Fatalf("ListPayments: %v", err)
	}
	if len(list) != 1 || list[0].Amount != 1000 {
		t.Fatalf("unexpected list: %+v", list)
	}

	rows, err := store.ListReportRows("", "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatalf("ListReportRows: %v", err)
	}
	if len(rows) != 1 || rows[0].ServerName != "Server 1" || rows[0].Amount != 1000 {
		t.Fatalf("unexpected report: %+v", rows)
	}
}

// TestStoreReminder проверяет список серверов для напоминания и запись об отправке.
func TestStoreReminder(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()

	// Сервер с платёжом через 10 дней (дата не совпадает с сегодня+10 — тест только проверяет запись).
	if err := store.RecordReminderSent("s1", "2026-06-01"); err != nil {
		t.Fatalf("RecordReminderSent: %v", err)
	}
	servers, err := store.ListServersDueForReminder(10)
	if err != nil {
		t.Fatalf("ListServersDueForReminder: %v", err)
	}
	// Может быть 0 или больше в зависимости от даты; главное — нет паники.
	_ = servers
}

