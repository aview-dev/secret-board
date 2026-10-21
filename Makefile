# secret-board: Go backend (src/backend) and SvelteKit portal (src/frontend/portal).
# Keep this Makefile thin and delegate real logic to scripts once a recipe grows.
#
# 비밀값(Google OAuth, BETTER_AUTH_SECRET)은 여기에 적지 않는다. 포탈은
# src/frontend/portal/.env 를 읽는다 (.env.example 참고).

ROOT_DIR := $(PWD)
SRC_DIR := $(ROOT_DIR)/src
BACKEND_DIR := $(SRC_DIR)/backend
PORTAL_DIR := $(SRC_DIR)/frontend/portal

# Colors
INFO_COLOR := \033[0;36m
RESET_COLOR := \033[0m

# --- Backend env (dev defaults; production values come from deploy compose) ---
# Shared secret between portal BFF and backend. Empty = check disabled (dev).
export SECRET_BOARD_BACKEND_SECRET ?=
export SECRET_BOARD_SERVER_ADDRESS ?= :8080
# board.db (posts / comments). Relative to src/backend when run via make.
export SECRET_BOARD_DB_PATH ?= board.db

# Version single-source-of-truth - see /VERSION. Env var override wins for CI.
include VERSION
export SECRET_BOARD_VERSION

GO_LDFLAGS := -s -w -X main.version=$(SECRET_BOARD_VERSION)

.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "secret-board - available targets:"
	@echo "  install             - Install portal dependencies (bun install)"
	@echo "  dev                 - Run backend + portal together (Ctrl-C stops both)"
	@echo "  backend             - Run the Go backend (:8080)"
	@echo "  dev-portal          - Run the portal in dev mode (Vite HMR, :5173)"
	@echo "  tidy                - go mod tidy"
	@echo "  build               - Build backend and portal"
	@echo "  build-backend       - go build over src/backend into build/"
	@echo "  build-portal        - Production build (adapter-node) into portal/build"
	@echo "  typecheck           - svelte-check over the portal"
	@echo "  test                - Backend go vet + go test, portal typecheck"
	@echo "  test-backend        - go vet + go test over src/backend"
	@echo "  test-frontend       - Portal typecheck"
	@echo "  format              - prettier --write"
	@echo "  lint                - prettier --check + eslint"
	@echo "  build-docker-images - Build backend/portal docker images (UPLOAD=1 pushes them)"

.PHONY: dev
dev:
	@bash $(ROOT_DIR)/scripts/dev.sh

.PHONY: backend
backend:
	cd $(BACKEND_DIR) && go run -ldflags "$(GO_LDFLAGS)" ./cmd/server

.PHONY: tidy
tidy:
	cd $(BACKEND_DIR) && go mod tidy

.PHONY: install
install:
	cd $(PORTAL_DIR) && bun install

.PHONY: dev-portal
dev-portal:
	cd $(PORTAL_DIR) && bun run dev

.PHONY: build build-backend build-portal
build: build-backend build-portal

build-backend:
	@echo "$(INFO_COLOR)>> building backend (go build)$(RESET_COLOR)"
	cd $(BACKEND_DIR) && CGO_ENABLED=0 go build -trimpath -ldflags "$(GO_LDFLAGS)" -o $(ROOT_DIR)/build/secret-board-server ./cmd/server

build-portal:
	@echo "$(INFO_COLOR)>> building portal (adapter-node)$(RESET_COLOR)"
	cd $(PORTAL_DIR) && bun install
	cd $(PORTAL_DIR) && bun run build

.PHONY: typecheck
typecheck:
	cd $(PORTAL_DIR) && bun run check

.PHONY: test test-backend test-frontend
test: test-backend test-frontend

test-backend:
	@echo "$(INFO_COLOR)>> backend go vet + go test$(RESET_COLOR)"
	cd $(BACKEND_DIR) && go vet ./... && go test ./...

test-frontend: typecheck

.PHONY: format
format:
	cd $(PORTAL_DIR) && bun run format

.PHONY: lint
lint:
	cd $(PORTAL_DIR) && bun run lint

# UPLOAD=1 pushes the images after the build.
.PHONY: build-docker-images
build-docker-images:
	@echo "$(INFO_COLOR)>> Building Docker images...$(RESET_COLOR)"
	@cd $(ROOT_DIR)/docker && UPLOAD=$(UPLOAD) bash build-docker-images.sh
