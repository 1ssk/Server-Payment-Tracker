package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/anyx/serversanyx-backend/internal/models"
)

// Store инкапсулирует работу с SQLite базой данных.
type Store struct {
	DB *sql.DB
}

// NewStore открывает соединение с SQLite по указанному пути и инициализирует схему.
func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	s := &Store{DB: db}
	if err := s.initSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return s, nil
}

// Close закрывает подключение к базе.
func (s *Store) Close() error {
	return s.DB.Close()
}

// initSchema создаёт таблицы, если их ещё нет.
func (s *Store) initSchema() error {
	const schema = `
CREATE TABLE IF NOT EXISTS servers (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	provider TEXT NOT NULL,
	provider_url TEXT,
	ip_address TEXT NOT NULL,
	location TEXT NOT NULL,
	monthly_cost REAL NOT NULL,
	billing_cycle TEXT NOT NULL,
	next_payment_date TEXT NOT NULL,
	status TEXT NOT NULL,
	notes TEXT,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS smtp_settings (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	host TEXT,
	port INTEGER,
	username TEXT,
	password TEXT,
	from_email TEXT,
	to_email TEXT,
	enabled INTEGER NOT NULL DEFAULT 0,
	reminder_days_before INTEGER NOT NULL DEFAULT 10
);

CREATE TABLE IF NOT EXISTS payments (
	id TEXT PRIMARY KEY,
	server_id TEXT NOT NULL,
	amount REAL NOT NULL,
	paid_at TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS reminder_sent (
	server_id TEXT NOT NULL,
	for_payment_date TEXT NOT NULL,
	sent_at TEXT NOT NULL,
	PRIMARY KEY (server_id, for_payment_date)
);
`
	if _, err := s.DB.Exec(schema); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	// Добавить колонку reminder_days_before в старых БД (игнорируем ошибку если уже есть).
	_, _ = s.DB.Exec(`ALTER TABLE smtp_settings ADD COLUMN reminder_days_before INTEGER NOT NULL DEFAULT 10`)
	return nil
}

// ListServers возвращает все сервера.
func (s *Store) ListServers() ([]models.Server, error) {
	rows, err := s.DB.Query(`
SELECT id, name, provider, provider_url, ip_address, location,
       monthly_cost, billing_cycle, next_payment_date, status, notes, created_at
FROM servers
ORDER BY created_at ASC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("query servers: %w", err)
	}
	defer rows.Close()

	var result []models.Server
	for rows.Next() {
		var server models.Server
		var providerURL sql.NullString
		var notes sql.NullString

		if err := rows.Scan(
			&server.ID,
			&server.Name,
			&server.Provider,
			&providerURL,
			&server.IPAddress,
			&server.Location,
			&server.MonthlyCost,
			&server.BillingCycle,
			&server.NextPaymentDate,
			&server.Status,
			&notes,
			&server.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan server: %w", err)
		}

		if providerURL.Valid {
			server.ProviderURL = &providerURL.String
		}
		if notes.Valid {
			server.Notes = &notes.String
		}

		result = append(result, server)
	}

	return result, nil
}

// InsertServer создаёт новый сервер.
func (s *Store) InsertServer(server models.Server) error {
	_, err := s.DB.Exec(`
INSERT INTO servers (
	id, name, provider, provider_url, ip_address, location,
	monthly_cost, billing_cycle, next_payment_date, status, notes, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		server.ID,
		server.Name,
		server.Provider,
		server.ProviderURL,
		server.IPAddress,
		server.Location,
		server.MonthlyCost,
		server.BillingCycle,
		server.NextPaymentDate,
		server.Status,
		server.Notes,
		server.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert server: %w", err)
	}
	return nil
}

// UpdateServer обновляет существующий сервер.
func (s *Store) UpdateServer(server models.Server) error {
	res, err := s.DB.Exec(`
UPDATE servers
SET name = ?, provider = ?, provider_url = ?, ip_address = ?, location = ?,
    monthly_cost = ?, billing_cycle = ?, next_payment_date = ?, status = ?, notes = ?
WHERE id = ?`,
		server.Name,
		server.Provider,
		server.ProviderURL,
		server.IPAddress,
		server.Location,
		server.MonthlyCost,
		server.BillingCycle,
		server.NextPaymentDate,
		server.Status,
		server.Notes,
		server.ID,
	)
	if err != nil {
		return fmt.Errorf("update server: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.New("server not found")
	}
	return nil
}

// DeleteServer удаляет сервер по ID.
func (s *Store) DeleteServer(id string) error {
	if _, err := s.DB.Exec(`DELETE FROM servers WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete server: %w", err)
	}
	return nil
}

// GetSMTPSettings возвращает сохранённые SMTP‑настройки.
func (s *Store) GetSMTPSettings() (models.SMTPSettings, error) {
	var (
		host, username, password, fromEmail, toEmail sql.NullString
		port, enabled                                sql.NullInt64
	)
	err := s.DB.QueryRow(`
SELECT host, port, username, password, from_email, to_email, enabled
FROM smtp_settings WHERE id = 1`).Scan(&host, &port, &username, &password, &fromEmail, &toEmail, &enabled)
	if err == sql.ErrNoRows {
		return models.SMTPSettings{
			Host:               "",
			Port:               587,
			Enabled:            false,
			ReminderDaysBefore: 10,
		}, nil
	}
	if err != nil {
		return models.SMTPSettings{}, fmt.Errorf("get smtp settings: %w", err)
	}
	rd := 10
	var rdN sql.NullInt64
	_ = s.DB.QueryRow(`SELECT reminder_days_before FROM smtp_settings WHERE id = 1`).Scan(&rdN)
	if rdN.Valid && rdN.Int64 > 0 {
		rd = int(rdN.Int64)
	}
	return models.SMTPSettings{
		Host:               host.String,
		Port:               int(port.Int64),
		Username:           username.String,
		Password:           password.String,
		From:               fromEmail.String,
		To:                 toEmail.String,
		Enabled:            enabled.Int64 == 1,
		ReminderDaysBefore: rd,
	}, nil
}

// SaveSMTPSettings сохраняет SMTP‑настройки (upsert по id=1).
func (s *Store) SaveSMTPSettings(settings models.SMTPSettings) error {
	enabledInt := 0
	if settings.Enabled {
		enabledInt = 1
	}
	rd := settings.ReminderDaysBefore
	if rd <= 0 {
		rd = 10
	}

	_, err := s.DB.Exec(`
INSERT INTO smtp_settings (id, host, port, username, password, from_email, to_email, enabled, reminder_days_before)
VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	host = excluded.host,
	port = excluded.port,
	username = excluded.username,
	password = excluded.password,
	from_email = excluded.from_email,
	to_email = excluded.to_email,
	enabled = excluded.enabled,
	reminder_days_before = excluded.reminder_days_before
`,
		settings.Host,
		settings.Port,
		settings.Username,
		settings.Password,
		settings.From,
		settings.To,
		enabledInt,
		rd,
	)
	if err != nil {
		return fmt.Errorf("save smtp settings: %w", err)
	}
	return nil
}

// InsertPayment добавляет запись об оплате и сдвигает next_payment_date у сервера.
func (s *Store) InsertPayment(p models.Payment) error {
	_, err := s.DB.Exec(`
INSERT INTO payments (id, server_id, amount, paid_at, created_at) VALUES (?, ?, ?, ?, ?)`,
		p.ID, p.ServerID, p.Amount, p.PaidAt, p.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}
	// Обновить дату следующего платежа у сервера (следующий период от paid_at).
	var cycle string
	var nextDate string
	if err := s.DB.QueryRow(`SELECT billing_cycle, next_payment_date FROM servers WHERE id = ?`, p.ServerID).Scan(&cycle, &nextDate); err != nil {
		return nil
	}
	nextDate = nextPaymentFrom(p.PaidAt, cycle)
	_, _ = s.DB.Exec(`UPDATE servers SET next_payment_date = ? WHERE id = ?`, nextDate, p.ServerID)
	return nil
}

func nextPaymentFrom(paidAt, cycle string) string {
	t, err := time.Parse("2006-01-02", paidAt)
	if err != nil {
		return paidAt
	}
	switch cycle {
	case "monthly":
		t = t.AddDate(0, 1, 0)
	case "quarterly":
		t = t.AddDate(0, 3, 0)
	case "yearly":
		t = t.AddDate(1, 0, 0)
	}
	return t.Format("2006-01-02")
}

// ListPayments возвращает платежи по серверу и/или за период.
func (s *Store) ListPayments(serverID, from, to string) ([]models.Payment, error) {
	q := `SELECT id, server_id, amount, paid_at, created_at FROM payments WHERE 1=1`
	args := []any{}
	if serverID != "" {
		q += ` AND server_id = ?`
		args = append(args, serverID)
	}
	if from != "" {
		q += ` AND paid_at >= ?`
		args = append(args, from)
	}
	if to != "" {
		q += ` AND paid_at <= ?`
		args = append(args, to)
	}
	q += ` ORDER BY paid_at DESC`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()
	var out []models.Payment
	for rows.Next() {
		var p models.Payment
		if err := rows.Scan(&p.ID, &p.ServerID, &p.Amount, &p.PaidAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// ListReportRows возвращает строки отчёта: платежи с именами серверов за период.
func (s *Store) ListReportRows(serverID, from, to string) ([]models.ReportRow, error) {
	q := `
SELECT p.server_id, COALESCE(s.name, ''), p.paid_at, p.amount
FROM payments p
LEFT JOIN servers s ON s.id = p.server_id
WHERE 1=1`
	args := []any{}
	if serverID != "" {
		q += ` AND p.server_id = ?`
		args = append(args, serverID)
	}
	if from != "" {
		q += ` AND p.paid_at >= ?`
		args = append(args, from)
	}
	if to != "" {
		q += ` AND p.paid_at <= ?`
		args = append(args, to)
	}
	q += ` ORDER BY p.paid_at DESC`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("list report: %w", err)
	}
	defer rows.Close()
	var out []models.ReportRow
	for rows.Next() {
		var r models.ReportRow
		if err := rows.Scan(&r.ServerID, &r.ServerName, &r.PaidAt, &r.Amount); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

// ListServersDueForReminder возвращает сервера, по которым нужно отправить напоминание сегодня (срок через daysBefore дней).
func (s *Store) ListServersDueForReminder(daysBefore int) ([]models.Server, error) {
	targetDate := time.Now().AddDate(0, 0, daysBefore).Format("2006-01-02")
	rows, err := s.DB.Query(`
SELECT id, name, provider, provider_url, ip_address, location,
       monthly_cost, billing_cycle, next_payment_date, status, notes, created_at
FROM servers
WHERE status = 'active' AND next_payment_date = ?
AND (id, next_payment_date) NOT IN (SELECT server_id, for_payment_date FROM reminder_sent)`,
		targetDate)
	if err != nil {
		return nil, fmt.Errorf("list due for reminder: %w", err)
	}
	defer rows.Close()
	var result []models.Server
	for rows.Next() {
		var server models.Server
		var providerURL sql.NullString
		var notes sql.NullString
		if err := rows.Scan(
			&server.ID, &server.Name, &server.Provider, &providerURL,
			&server.IPAddress, &server.Location, &server.MonthlyCost,
			&server.BillingCycle, &server.NextPaymentDate, &server.Status,
			&notes, &server.CreatedAt,
		); err != nil {
			return nil, err
		}
		if providerURL.Valid {
			server.ProviderURL = &providerURL.String
		}
		if notes.Valid {
			server.Notes = &notes.String
		}
		result = append(result, server)
	}
	return result, nil
}

// RecordReminderSent фиксирует отправку напоминания по серверу за указанную дату платежа.
func (s *Store) RecordReminderSent(serverID, forPaymentDate string) error {
	_, err := s.DB.Exec(`
INSERT INTO reminder_sent (server_id, for_payment_date, sent_at) VALUES (?, ?, ?)`,
		serverID, forPaymentDate, time.Now().Format(time.RFC3339))
	return err
}

