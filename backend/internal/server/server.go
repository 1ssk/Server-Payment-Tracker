package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

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
	s.mux.HandleFunc("GET /api/servers", s.handleGetServers)
	s.mux.HandleFunc("POST /api/servers", s.handleCreateServer)
	s.mux.HandleFunc("PUT /api/servers", s.handleUpdateServer)
	s.mux.HandleFunc("DELETE /api/servers", s.handleDeleteServer)

	s.mux.HandleFunc("GET /api/smtp-settings", s.handleGetSMTPSettings)
	s.mux.HandleFunc("PUT /api/smtp-settings", s.handleUpdateSMTPSettings)
}

// handleGetServers возвращает список серверов.
func (s *Server) handleGetServers(w http.ResponseWriter, r *http.Request) {
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
	settings, err := s.store.GetSMTPSettings()
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

// handleUpdateSMTPSettings сохраняет SMTP‑настройки.
func (s *Server) handleUpdateSMTPSettings(w http.ResponseWriter, r *http.Request) {
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

