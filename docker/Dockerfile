# secret-board images (multi-stage, one file,
# --target per image). The backend is pure Go (CGO off, static binary); the
# portal is SvelteKit adapter-node built with bun and run with plain node.
#
# Build from the repo root context:
#   docker build --target secret-board-backend -f docker/Dockerfile .
#   docker build --target secret-board-portal  -f docker/Dockerfile .
# or just run docker/build-docker-images.sh (tags + push).

########################################################################################################################
# build backend (Go, static binary)
########################################################################################################################

FROM golang:1.25 AS backend-build

WORKDIR /secret-board/backend

# warm the module cache before copying sources so code edits reuse this layer.
# go.su[m]: go.sum appears only once the module has dependencies.
COPY src/backend/go.mod src/backend/go.su[m] ./
RUN go mod download

COPY src/backend/ ./
ARG ARG_SECRET_BOARD_VERSION="unknown"
RUN CGO_ENABLED=0 go build -v -trimpath \
        -ldflags "-s -w -X main.version=${ARG_SECRET_BOARD_VERSION}" \
        -o /out/secret-board-server ./cmd/server

########################################################################################################################
# build portal (SvelteKit adapter-node)
########################################################################################################################

# node base, not oven/bun: the portal uses node's built-in `node:sqlite`, which
# bun does not provide, and `vite build` imports every server module to read
# its page options. bun is still the package manager (bun.lock).
FROM node:24-slim AS portal-build

RUN npm install -g bun@1.2.21

WORKDIR /secret-board/src/frontend/portal

COPY src/frontend/portal/package.json src/frontend/portal/bun.lock ./
RUN bun install --frozen-lockfile

COPY src/frontend/portal/ ./

# better-auth refuses to construct on the default secret. This placeholder only
# satisfies the build-time import; the runtime image (a separate stage) reads
# the real value through $env/dynamic/private.
ENV BETTER_AUTH_SECRET=build-time-placeholder-not-a-secret

RUN bun run build

########################################################################################################################
# portal production deps (adapter-node needs package.json + prod node_modules)
########################################################################################################################

FROM portal-build AS portal-prod-deps

RUN rm -rf node_modules && bun install --frozen-lockfile --production

########################################################################################################################
# secret-board backend image (posts / comments API, never sees user identity)
########################################################################################################################

FROM gcr.io/distroless/static-debian12:nonroot AS secret-board-backend

ARG ARG_SECRET_BOARD_VERSION="unknown"
ENV SECRET_BOARD_VERSION=$ARG_SECRET_BOARD_VERSION
ENV RUNNING_IN_DOCKER=true

WORKDIR /secret-board

COPY --from=backend-build /out/secret-board-server ./secret-board-server

EXPOSE 8080

ENTRYPOINT ["./secret-board-server"]

########################################################################################################################
# secret-board portal image (SvelteKit BFF + Google sign-in)
########################################################################################################################

FROM node:24-slim AS secret-board-portal

ARG ARG_SECRET_BOARD_VERSION="unknown"
ENV SECRET_BOARD_VERSION=$ARG_SECRET_BOARD_VERSION
ENV RUNNING_IN_DOCKER=true
ENV NODE_ENV=production
# auth.db (who signed in). Mounted from ./vol/portal in the deploy compose.
ENV AUTH_DB_PATH=/var/secret-board/portal/auth.db

WORKDIR /secret-board/portal

COPY --from=portal-build     /secret-board/src/frontend/portal/build        ./build
COPY --from=portal-build     /secret-board/src/frontend/portal/package.json ./package.json
COPY --from=portal-prod-deps /secret-board/src/frontend/portal/node_modules ./node_modules

RUN mkdir -p /var/secret-board/portal && chown node:node /var/secret-board/portal
USER node

# adapter-node runtime knobs: PORT / ORIGIN come from the environment
# (see deploy/secret-board/docker-compose.yml).
EXPOSE 3000

CMD ["node", "build"]
