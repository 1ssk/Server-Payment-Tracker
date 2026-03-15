package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/anyx/serversanyx-backend/internal/mailer"
	"github.com/anyx/serversanyx-backend/internal/models"
	"github.com/anyx/serversanyx-backend/internal/storage"
)

// Server объединяет HTTP‑слой и хранилище.
type Server struct {
	store *storage.Store
	mux   *http.ServeMux
}

// New создаёт новый HTTP‑сервер поверх указанного хранилища.
func New(store *storage.Store) *Server {
	s := &Server{
		store: store,
		mux:   http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// Handler возвращает корневой http.Handler.
func (s *Server) Handler() http.Handler {
	return logRequest(s.mux)
}

// registerRoutes регистрирует все HTTP‑маршруты API.
func (s *Server) registerRoutes() {
	// Авторизация
	s.mux.HandleFunc("POST /api/login", s.handleLogin)

	// Операции с серверами (требуют авторизации)
	s.mux.HandleFunc("GET /api/servers", s.handleGetServers)
	s.mux.HandleFunc("POST /api/servers", s.handleCreateServer)
	s.mux.HandleFunc("PUT /api/servers", s.handleUpdateServer)
	s.mux.HandleFunc("DELETE /api/servers", s.handleDeleteServer)

	// Платежи и отчёты
	s.mux.HandleFunc("POST /api/servers/{id}/payments", s.handleConfirmPayment)
	s.mux.HandleFunc("GET /api/servers/{id}/payments", s.handleGetServerPayments)
	s.mux.HandleFunc("GET /api/reports", s.handleGetReports)

	// SMTP‑настройки
	s.mux.HandleFunc("GET /api/smtp-settings", s.handleGetSMTPSettings)
	s.mux.HandleFunc("PUT /api/smtp-settings", s.handleUpdateSMTPSettings)
	s.mux.HandleFunc("POST /api/smtp-settings/test", s.handleSendTestEmail)
}

// handleLogin выполняет проверку логина/пароля и возвращает токен.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Учётные данные только из переменных окружения (без дефолтов в коде).
	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")
	if username == "" || password == "" {
		http.Error(w, "auth not configured", http.StatusServiceUnavailable)
		return
	}
	if req.Username != username || req.Password != password {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token := authToken()
	writeJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{Token: token})
}

// handleGetServers возвращает список серверов.
func (s *Server) handleGetServers(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(w, r) {
		return
	}
	servers, err := s.store.ListServers()
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Servers []models.VPNServer `json:"servers"`
	}{Servers: servers})
}

// handleCreateServer создаёт новый сервер.
func (s *Server) handleCreateServer(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(w, r) {
		return
	}
	var server models.VPNServer
	if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if server.ID == "" {
		server.ID = time.Now().Format("20060102150405")
	}
	if server.CreatedAt == "" {
		server.CreatedAt = time.Now().Format("2006-01-02")
	}

	if err := s.store.InsertServer(server); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, server)
}

// handleUpdateServer обновляет существующий сервер.
func (s *Server) handleUpdateServer(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(w, r) {
		return
	}
	var server models.VPNServer
	if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if server.ID == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := s.store.UpdateServer(server); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, server)
}

// handleDeleteServer удаляет сервер по ID.
func (s *Server) handleDeleteServer(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(w, r) {
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := s.store.DeleteServer(id); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGetSMTPSettings возвращает сохранённые SMTP‑настройки.
func (s *Server) handleGetSMTPSettings(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(w, r) {
		return
	}
	settings, err := s.store.GetSMTPSettings()
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

// handleUpdateSMTPSettings сохраняет SMTP‑настройки.
func (s *Server) handleUpdateSMTPSettings(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(w, r) {
		return
	}
	var settings models.SMTPSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := s.store.SaveSMTPSettings(settings); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, settings)
}

// handleConfirmPayment — подтверждение оплаты по серверу (POST /api/servers/{id}/payments).
func (s *Server) handleConfirmPayment(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(w, r) {
		return
	}
	serverID := r.PathValue("id")
	if serverID == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	var req struct {
		PaidAt string  `json:"paidAt"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.PaidAt == "" {
		http.Error(w, "paidAt required", http.StatusBadRequest)
		return
	}
	if req.Amount <= 0 {
		// Подставить стоимость из сервера
		servers, _ := s.store.ListServers()
		for _, sv := range servers {
			if sv.ID == serverID {
				req.Amount = sv.MonthlyCost
				break
			}
		}
		if req.Amount <= 0 {
			http.Error(w, "amount required", http.StatusBadRequest)
			return
		}
	}
	p := models.Payment{
		ID:        time.Now().Format("20060102150405") + "-" + serverID,
		ServerID:  serverID,
		Amount:    req.Amount,
		PaidAt:    req.PaidAt,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}
	if err := s.store.InsertPayment(p); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

// handleGetServerPayments — история платежей по серверу.
func (s *Server) handleGetServerPayments(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(w, r) {
		return
	}
	serverID := r.PathValue("id")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	list, err := s.store.ListPayments(serverID, from, to)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Payments []models.Payment `json:"payments"`
	}{Payments: list})
}

// handleGetReports — отчёт по расходам (фильтр по серверу и периоду).
func (s *Server) handleGetReports(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(w, r) {
		return
	}
	serverID := r.URL.Query().Get("serverId")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	rows, err := s.store.ListReportRows(serverID, from, to)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Rows []models.ReportRow `json:"rows"`
	}{Rows: rows})
}

// handleSendTestEmail — отправить тестовое письмо (проверка SMTP).
func (s *Server) handleSendTestEmail(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(w, r) {
		return
	}
	settings, err := s.store.GetSMTPSettings()
	if err != nil || !settings.Enabled || settings.Host == "" || settings.To == "" {
		http.Error(w, "smtp not configured", http.StatusBadRequest)
		return
	}
	if err := mailer.SendTestEmail(settings, settings.To); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Ok bool `json:"ok"`
	}{Ok: true})
}

// writeJSON отправляет JSON‑ответ с нужным статус‑кодом.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode json: %v", err)
	}
}

// logRequest оборачивает handler и пишет простые HTTP‑логи.
func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// authToken возвращает "секретный" токен, который фронт должен передавать в заголовке Authorization.
func authToken() string {
	if v := os.Getenv("AUTH_TOKEN"); v != "" {
		return v
	}
	// Значение по умолчанию для простых установок.
	return "demo-token"
}

// checkAuth проверяет заголовок Authorization: Bearer <token>.
func checkAuth(w http.ResponseWriter, r *http.Request) bool {
	// Эндпоинт логина не требует авторизации.
	if r.URL.Path == "/api/login" {
		return true
	}

	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if len(h) < len(prefix) || h[:len(prefix)] != prefix {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}

	token := h[len(prefix):]
	if token != authToken() {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}


