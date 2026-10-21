package httpapi

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"secret-board/backend/internal/board"
)

func newTestServer(t *testing.T, secret string) http.Handler {
	t.Helper()
	store, err := board.Open(filepath.Join(t.TempDir(), "board.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return NewServer(Config{PortalSecret: secret, Version: "test", Store: store})
}

func TestHealthzIsPublic(t *testing.T) {
	h := newTestServer(t, "s3cret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("healthz: got %d, want 200", rec.Code)
	}
}

func TestAPIRequiresPortalSecret(t *testing.T) {
	h := newTestServer(t, "s3cret")

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

func TestCreatePostAndComment(t *testing.T) {
	h := newTestServer(t, "")

	do := func(method, path, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
		return rec
	}

	if rec := do("POST", "/api/v1/posts", `{"title":"t","body":"b"}`); rec.Code != http.StatusCreated {
		t.Fatalf("create post: got %d %s", rec.Code, rec.Body)
	}
	if rec := do("POST", "/api/v1/posts", `{"title":"","body":"b"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("blank title: got %d", rec.Code)
	}
	if rec := do("POST", "/api/v1/posts/1/comments", `{"body":"c"}`); rec.Code != http.StatusCreated {
		t.Fatalf("create comment: got %d %s", rec.Code, rec.Body)
	}
	if rec := do("POST", "/api/v1/posts/42/comments", `{"body":"c"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("comment on missing post: got %d", rec.Code)
	}

	rec := do("GET", "/api/v1/posts/1", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"body":"c"`) {
		t.Fatalf("get post: got %d %s", rec.Code, rec.Body)
	}
}

func TestVote(t *testing.T) {
	h := newTestServer(t, "")
	voter := strings.Repeat("c", 64)

	do := func(method, path, body string, header map[string]string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		for k, v := range header {
			req.Header.Set(k, v)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}

	do("POST", "/api/v1/posts", `{"title":"t","body":"b"}`, nil)

	rec := do("PUT", "/api/v1/posts/1/vote", `{"voter":"`+voter+`","value":1}`, nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"likes":1`) {
		t.Fatalf("vote: got %d %s", rec.Code, rec.Body)
	}
	if rec := do("PUT", "/api/v1/posts/1/vote", `{"voter":"me@corelinesoft.com","value":1}`, nil); rec.Code != http.StatusBadRequest {
		t.Fatalf("raw email as voter: got %d", rec.Code)
	}
	if rec := do("PUT", "/api/v1/posts/42/vote", `{"voter":"`+voter+`","value":1}`, nil); rec.Code != http.StatusNotFound {
		t.Fatalf("vote on missing post: got %d", rec.Code)
	}

	rec = do("GET", "/api/v1/posts/1", "", map[string]string{"X-Voter": voter})
	if !strings.Contains(rec.Body.String(), `"myVote":1`) {
		t.Fatalf("get post with voter: %s", rec.Body)
	}
	rec = do("GET", "/api/v1/posts/1", "", nil)
	if !strings.Contains(rec.Body.String(), `"myVote":0`) {
		t.Fatalf("get post without voter: %s", rec.Body)
	}
}
