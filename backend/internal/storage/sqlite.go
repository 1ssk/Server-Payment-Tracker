package storage

import (
	"database/sql"
	"errors"
	"fmt"

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
	enabled INTEGER NOT NULL DEFAULT 0
);
`
	if _, err := s.DB.Exec(schema); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}

// ListServers возвращает все сервера.
func (s *Store) ListServers() ([]models.VPNServer, error) {
	rows, err := s.DB.Query(`
SELECT id, name, provider, provider_url, ip_address, location,
       monthly_cost, billing_cycle, next_payment_date, status, notes, created_at
FROM servers
ORDER BY created_at ASC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("query servers: %w", err)
	}
	defer rows.Close()

	var result []models.VPNServer
	for rows.Next() {
		var server models.VPNServer
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
func (s *Store) InsertServer(server models.VPNServer) error {
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
func (s *Store) UpdateServer(server models.VPNServer) error {
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
		port                                        sql.NullInt64
		enabled                                     sql.NullInt64
	)

	err := s.DB.QueryRow(`
SELECT host, port, username, password, from_email, to_email, enabled
FROM smtp_settings
WHERE id = 1`,
	).Scan(&host, &port, &username, &password, &fromEmail, &toEmail, &enabled)

	if err == sql.ErrNoRows {
		// Возвращаем значения по умолчанию.
		return models.SMTPSettings{
			Host:    "",
			Port:    587,
			Enabled: false,
		}, nil
	}
	if err != nil {
		return models.SMTPSettings{}, fmt.Errorf("get smtp settings: %w", err)
	}

	return models.SMTPSettings{
		Host:     host.String,
		Port:     int(port.Int64),
		Username: username.String,
		Password: password.String,
		From:     fromEmail.String,
		To:       toEmail.String,
		Enabled:  enabled.Int64 == 1,
	}, nil
}

// SaveSMTPSettings сохраняет SMTP‑настройки (upsert по id=1).
func (s *Store) SaveSMTPSettings(settings models.SMTPSettings) error {
	enabledInt := 0
	if settings.Enabled {
		enabledInt = 1
	}

	_, err := s.DB.Exec(`
INSERT INTO smtp_settings (id, host, port, username, password, from_email, to_email, enabled)
VALUES (1, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	host = excluded.host,
	port = excluded.port,
	username = excluded.username,
	password = excluded.password,
	from_email = excluded.from_email,
	to_email = excluded.to_email,
	enabled = excluded.enabled
`,
		settings.Host,
		settings.Port,
		settings.Username,
		settings.Password,
		settings.From,
		settings.To,
		enabledInt,
	)
	if err != nil {
		return fmt.Errorf("save smtp settings: %w", err)
	}

	return nil
}

