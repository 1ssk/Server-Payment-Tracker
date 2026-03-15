package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHandleLogin_NoCredentials(t *testing.T) {
	os.Unsetenv("ADMIN_USERNAME")
	os.Unsetenv("ADMIN_PASSWORD")
	defer func() {
		_ = os.Unsetenv("ADMIN_USERNAME")
		_ = os.Unsetenv("ADMIN_PASSWORD")
	}()

	s := New(nil)
	handler := s.Handler()
	body := bytes.NewBufferString(`{"username":"a","password":"b"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/login", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when auth not configured, got %d", rr.Code)
	}
}

func TestHandleLogin_WithCredentials(t *testing.T) {
	os.Setenv("ADMIN_USERNAME", "testuser")
	os.Setenv("ADMIN_PASSWORD", "testpass")
	defer func() {
		_ = os.Unsetenv("ADMIN_USERNAME")
		_ = os.Unsetenv("ADMIN_PASSWORD")
	}()

	s := New(nil)
	handler := s.Handler()
	body := bytes.NewBufferString(`{"username":"testuser","password":"testpass"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/login", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 on valid login, got %d body=%s", rr.Code, rr.Body.String())
	}
}
