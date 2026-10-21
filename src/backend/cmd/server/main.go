// secret-board backend. 게시글 / 댓글 저장소이며 포탈(SvelteKit BFF)만 호출한다.
//
// 익명성 규칙 (docs/anonymity.md):
//   - 요청자 신원(이메일, 세션, IP)을 받지도, 저장하지도, 로그에 남기지도 않는다.
//   - 접근 로그(access log) 미들웨어를 붙이지 않는다.

package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"secret-board/backend/internal/board"
	"secret-board/backend/internal/httpapi"
)

var version = "dev"

func main() {
	addr := envOr("SECRET_BOARD_SERVER_ADDRESS", ":8080")
	secret := os.Getenv("SECRET_BOARD_BACKEND_SECRET")
	if secret == "" {
		slog.Warn("SECRET_BOARD_BACKEND_SECRET is empty, portal secret check is disabled (dev only)")
	}

	dbPath := envOr("SECRET_BOARD_DB_PATH", "board.db")
	store, err := board.Open(dbPath)
	if err != nil {
		slog.Error("open board db failed", "path", dbPath, "err", err)
		os.Exit(1)
	}
	defer store.Close()

	srv := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewServer(httpapi.Config{PortalSecret: secret, Version: version, Store: store}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("secret-board backend listening", "addr", addr, "version", version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
