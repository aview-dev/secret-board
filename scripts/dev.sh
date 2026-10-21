#! /bin/bash
#
# Runs the Go backend and the portal dev server together (`make dev`).
# Each line is prefixed with [backend] / [portal]. Ctrl-C stops both, and if
# either one exits (crash, port in use) the other is stopped too.

set -uo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${SECRET_BOARD_VERSION:-dev}"

# Kill the whole process group (go run's child binary, vite) on any exit.
trap 'trap - EXIT; kill 0 2>/dev/null' EXIT INT TERM

(
    cd "${ROOT_DIR}/src/backend" &&
        go run -ldflags "-X main.version=${VERSION}" ./cmd/server 2>&1 |
        sed -u 's/^/[backend] /'
) &

(
    cd "${ROOT_DIR}/src/frontend/portal" &&
        bun run dev 2>&1 |
        sed -u 's/^/[portal]  /'
) &

# Return as soon as either side stops; the EXIT trap takes the other one down.
wait -n
