package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthzIsPublic(t *testing.T) {
	h := NewServer(Config{PortalSecret: "s3cret", Version: "test"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("healthz: got %d, want 200", rec.Code)
	}
}

func TestAPIRequiresPortalSecret(t *testing.T) {
	h := NewServer(Config{PortalSecret: "s3cret"})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("without secret: got %d, want 401", rec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil)
	req.Header.Set("X-Portal-Secret", "s3cret")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code == http.StatusUnauthorized {
		t.Fatalf("with secret: got 401")
	}
}
