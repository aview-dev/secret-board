#! /bin/bash
#
# Builds the secret-board images.
# Tags:
#   ${IMAGE_REPO}:<version>          backend
#   ${IMAGE_REPO}:<version>-portal   SvelteKit portal (adapter-node)
#
# Usage (from the docker/ directory, or via `make build-docker-images`):
#   ./build-docker-images.sh                 build only
#   ./build-docker-images.sh --upload        build and push
#   UPLOAD=1 ./build-docker-images.sh        same via env (1 / true / yes)
#   make build-docker-images UPLOAD=1        same via make
#
# IMAGE_REPO defaults to `secret-board`; set e.g. ghcr.io/<owner>/secret-board
# to push to a registry. The version comes from the repo-root VERSION file;
# override per-run with SECRET_BOARD_VERSION=... ./build-docker-images.sh.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE_REPO="${IMAGE_REPO:-secret-board}"

# version (single source of truth: /VERSION)
if [ -z "${SECRET_BOARD_VERSION}" ]; then
    SECRET_BOARD_VERSION="$(sed -n 's/^SECRET_BOARD_VERSION=//p' "${SCRIPT_DIR}/../VERSION")"
fi
if [ -z "${SECRET_BOARD_VERSION}" ]; then
    echo "Failed to resolve SECRET_BOARD_VERSION from ${SCRIPT_DIR}/../VERSION"
    exit 1
fi

VERSION="${SECRET_BOARD_VERSION}"

# --- flag / env parsing ----------------------------------------------
# Env vars accept truthy values (1, true, yes). CLI flags always win.
UPLOAD_FLAG=""
case "${UPLOAD:-}" in 1|true|yes|TRUE|YES) UPLOAD_FLAG=1 ;; esac
for arg in "$@"; do
    case "${arg}" in
        --upload) UPLOAD_FLAG=1 ;;
        *) echo "Unknown argument: ${arg}" >&2; exit 2 ;;
    esac
done

########################################################################################################################
# Build images

echo "Building ${IMAGE_REPO} with SECRET_BOARD_VERSION=${VERSION} (upload: ${UPLOAD_FLAG:-no})"

docker build --tag ${IMAGE_REPO}:${VERSION} --target secret-board-backend \
    --file "${SCRIPT_DIR}/Dockerfile" \
    --build-arg ARG_SECRET_BOARD_VERSION=${VERSION} \
    "${SCRIPT_DIR}/.."
if [ $? -ne 0 ]; then
    echo "Failed to build backend"
    exit 1
fi

docker build --tag ${IMAGE_REPO}:${VERSION}-portal --target secret-board-portal \
    --file "${SCRIPT_DIR}/Dockerfile" \
    --build-arg ARG_SECRET_BOARD_VERSION=${VERSION} \
    "${SCRIPT_DIR}/.."
if [ $? -ne 0 ]; then
    echo "Failed to build portal"
    exit 1
fi

########################################################################################################################
# Push images

if [ -n "${UPLOAD_FLAG}" ]; then
    docker push ${IMAGE_REPO}:${VERSION}
    if [ $? -ne 0 ]; then
        echo "Failed to push backend image"
        exit 1
    fi
    docker push ${IMAGE_REPO}:${VERSION}-portal
    if [ $? -ne 0 ]; then
        echo "Failed to push portal image"
        exit 1
    fi
fi

echo "Build completed successfully"
