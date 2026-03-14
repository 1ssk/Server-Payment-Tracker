package models

// VPNServer описывает один VPN‑сервер в системе.
type VPNServer struct {
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

// SMTPSettings хранит настройки SMTP‑сервера для email‑уведомлений.
type SMTPSettings struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
	Enabled  bool   `json:"enabled"`
}

