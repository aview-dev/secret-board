// HTTP 라우팅. 포탈 BFF가 유일한 호출자이며, X-Portal-Secret 헤더로 확인한다.

package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"secret-board/backend/internal/board"
)

type Config struct {
	// 포탈과 공유하는 시크릿. 비어 있으면 검사를 건너뛴다 (개발용).
	PortalSecret string
	Version      string
	Store        *board.Store
}

// 요청 본문 상한. 본문 10000 자가 UTF-8 로 최대 40KB 쯤이라 넉넉히 잡는다.
const maxRequestBytes = 64 << 10

func NewServer(cfg Config) http.Handler {
	mux := http.NewServeMux()

	// 헬스체크는 시크릿 없이 열어 둔다 (docker / 로드밸런서용).
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": cfg.Version})
	})

	api := http.NewServeMux()
	h := &handlers{store: cfg.Store}
	api.HandleFunc("GET /api/v1/posts", h.listPosts)
	api.HandleFunc("POST /api/v1/posts", h.createPost)
	api.HandleFunc("GET /api/v1/posts/{id}", h.getPost)
	api.HandleFunc("POST /api/v1/posts/{id}/comments", h.createComment)
	mux.Handle("/api/", requirePortalSecret(cfg.PortalSecret, api))

	return mux
}

type handlers struct {
	store *board.Store
}

func (h *handlers) listPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.store.ListPosts(r.Context(), 100)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"posts": posts})
}

func (h *handlers) createPost(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if !readJSON(w, r, &in) {
		return
	}
	p, err := h.store.CreatePost(r.Context(), in.Title, in.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

// getPost 는 글 하나와 댓글 전체를 함께 돌려준다 (상세 화면이 한 번에 그리도록).
func (h *handlers) getPost(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	p, err := h.store.GetPost(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	comments, err := h.store.ListComments(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"post": p, "comments": comments})
}

func (h *handlers) createComment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		Body string `json:"body"`
	}
	if !readJSON(w, r, &in) {
		return
	}
	c, err := h.store.CreateComment(r.Context(), id, in.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return 0, false
	}
	return id, true
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad_request"})
		return false
	}
	return true
}

// writeError 는 저장소 에러를 상태 코드로 바꾼다.
// 예상하지 못한 에러만 로그에 남기고, 요청자 정보는 넣지 않는다 (규칙 3).
func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, board.ErrInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid"})
	case errors.Is(err, board.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
	default:
		slog.Error("store failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal"})
	}
}

func requirePortalSecret(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if secret != "" && subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Portal-Secret")), []byte(secret)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
