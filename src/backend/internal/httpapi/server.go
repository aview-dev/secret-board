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
	api.HandleFunc("DELETE /api/v1/posts/{id}", h.deletePost)
	api.HandleFunc("POST /api/v1/posts/{id}/comments", h.createComment)
	api.HandleFunc("DELETE /api/v1/posts/{id}/comments/{commentId}", h.deleteComment)
	api.HandleFunc("PUT /api/v1/posts/{id}/vote", h.vote)
	api.HandleFunc("PUT /api/v1/posts/{id}/report", h.report)
	api.HandleFunc("DELETE /api/v1/posts/{id}/reports", h.clearReports)
	api.HandleFunc("GET /api/v1/reports", h.listReported)
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
// X-Voter 헤더가 있으면 그 voter 의 표(myVote)도 붙인다. voter 는 글마다 다른 해시다 (규칙 9).
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
	myVote, err := h.store.MyVote(r.Context(), id, r.Header.Get("X-Voter"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"post": p, "comments": comments, "myVote": myVote})
}

// deletePost 는 관리자 삭제용이다. 관리자인지는 portal 이 확인하고, 여기에는 누가 지우는지 오지 않는다 (규칙 7).
func (h *handlers) deletePost(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.store.DeletePost(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

// deleteComment 는 관리자 삭제용이다. deletePost 와 마찬가지로 누가 지우는지는 오지 않는다 (규칙 7).
func (h *handlers) deleteComment(w http.ResponseWriter, r *http.Request) {
	postID, ok := pathInt(w, r, "id")
	if !ok {
		return
	}
	commentID, ok := pathInt(w, r, "commentId")
	if !ok {
		return
	}
	if err := h.store.DeleteComment(r.Context(), postID, commentID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// vote 는 좋아요(1) / 싫어요(-1) / 취소(0) 를 받는다. voter 는 portal 이 만든 해시만 온다.
func (h *handlers) vote(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		Voter string `json:"voter"`
		Value int    `json:"value"`
	}
	if !readJSON(w, r, &in) {
		return
	}
	v, err := h.store.Vote(r.Context(), id, in.Voter, in.Value)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// report 는 신고를 받는다. reporter 는 portal 이 만든 해시만 오고, reason 은 정해진 값 중 하나다 (규칙 11).
func (h *handlers) report(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		Reporter string `json:"reporter"`
		Reason   string `json:"reason"`
	}
	if !readJSON(w, r, &in) {
		return
	}
	if err := h.store.Report(r.Context(), id, in.Reporter, in.Reason); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listReported 와 clearReports 는 관리자 신고 목록용이다. 관리자인지는 portal 이 확인한다.
func (h *handlers) listReported(w http.ResponseWriter, r *http.Request) {
	posts, err := h.store.ListReported(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"posts": posts})
}

func (h *handlers) clearReports(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.store.ClearReports(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	return pathInt(w, r, "id")
}

// pathInt 는 경로의 {name} 을 양의 정수로 읽는다. 아니면 404 를 쓰고 false.
func pathInt(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
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
