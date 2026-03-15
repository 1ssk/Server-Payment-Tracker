package models

// Server описывает один сервер в системе учёта.
type Server struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Provider        string  `json:"provider"`
	ProviderURL     *string `json:"providerUrl,omitempty"`
	IPAddress       string  `json:"ipAddress"`
	Location        string  `json:"location"`
	MonthlyCost     float64 `json:"monthlyCost"`
	BillingCycle    string  `json:"billingCycle"`
	NextPaymentDate string  `json:"nextPaymentDate"`
	Status          string  `json:"status"`
	Notes           *string `json:"notes,omitempty"`
	CreatedAt       string  `json:"createdAt"`
}

// VPNServer — алиас для совместимости API (фронт ожидает те же поля).
type VPNServer = Server

// Payment — запись о подтверждённой оплате по серверу.
type Payment struct {
	ID        string  `json:"id"`
	ServerID  string  `json:"serverId"`
	Amount    float64 `json:"amount"`
	PaidAt    string  `json:"paidAt"`
	CreatedAt string  `json:"createdAt"`
}

// SMTPSettings хранит настройки SMTP и напоминаний.
type SMTPSettings struct {
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	From               string `json:"from"`
	To                 string `json:"to"`
	Enabled            bool   `json:"enabled"`
	ReminderDaysBefore int    `json:"reminderDaysBefore"` // за сколько дней до срока присылать напоминание (по умолчанию 10)
}

// ReportRow — строка отчёта по расходам.
type ReportRow struct {
	ServerID   string  `json:"serverId"`
	ServerName string  `json:"serverName"`
	PaidAt     string  `json:"paidAt"`
	Amount     float64 `json:"amount"`
}

