// HTTP 라우팅. 포탈 BFF가 유일한 호출자이며, X-Portal-Secret 헤더로 확인한다.

package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
)

type Config struct {
	// 포탈과 공유하는 시크릿. 비어 있으면 검사를 건너뛴다 (개발용).
	PortalSecret string
	Version      string
}

func NewServer(cfg Config) http.Handler {
	mux := http.NewServeMux()

	// 헬스체크는 시크릿 없이 열어 둔다 (docker / 로드밸런서용).
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": cfg.Version})
	})

	api := http.NewServeMux()
	// TODO(다음 단계): 게시글 API
	//   GET  /api/v1/posts
	//   POST /api/v1/posts
	mux.Handle("/api/", requirePortalSecret(cfg.PortalSecret, api))

	return mux
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
