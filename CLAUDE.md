# secret-board

corelinesoft.com Google 계정만 로그인할 수 있는 사내 익명 게시판. 저장소 구조는 docker / docs / src 로 나눈다.

## 구조

- `src/frontend/portal/` - SvelteKit 2 (Svelte 5 runes), adapter-node, tailwind v4, better-auth. 로그인과 화면, backend 를 부르는 BFF.
- `src/backend/` - Go (stdlib net/http, log/slog). 게시글 / 댓글 저장소. portal 만 호출한다 (`X-Portal-Secret`).
- `docker/` - 멀티 스테이지 Dockerfile (`--target secret-board-backend|secret-board-portal`), 배포 compose.
- `docs/` - 설계 문서. 익명성 규칙은 `docs/anonymity.md`.
- `VERSION` - 버전 단일 소스. `mise.toml` - go / bun / node 버전 고정.

## 명령

- `make dev` - backend + portal 동시 실행 (따로: `make backend`, `make dev-portal`)
- `make test` - go vet + go test + svelte-check
- `make lint`, `make format`
- `make build-docker-images` (UPLOAD=1 이면 push)

## 반드시 지킬 것: 익명성

`docs/anonymity.md` 의 규칙을 깨는 변경은 하지 않는다. 특히:

- backend 스키마, API 요청 / 응답, 로그에 사용자 식별자(이메일, user id, 세션 id, IP, User-Agent)를 넣지 않는다.
- 요청 단위 접근 로그 미들웨어를 추가하지 않는다.
- portal 에서 backend 로 사용자 정보를 전달하지 않는다. `locals.user` 는 로그인 여부 확인에만 쓴다.
- 이런 요구가 들어오면 구현하기 전에 사용자에게 익명성 영향을 설명하고 확인받는다.

## 관례

- 비밀값을 코드, Makefile, 문서에 적지 않는다. `.env` (git 제외) 와 `.env.example` (키 이름만) 을 쓴다.
- portal 은 `node:sqlite` 를 쓰므로 vite / 서버는 Node 로 돈다 (bun 은 패키지 매니저로만). `bun --bun` 으로 실행하지 않는다.
- 빌드 중 (`$app/environment` 의 `building`) 에는 DB 를 열지 않는다.
- 문장 부호: `.claude/docs/text-style.md` (em-dash / en-dash 금지).
- 사용자는 직접 코딩하며 배우고 있다. 큰 변경은 단계별로 나누고 무엇을 왜 바꿨는지 설명한다.
