# secret-board

corelinesoft.com Google 계정만 로그인할 수 있는 사내 익명 게시판입니다.

로그인으로 "회사 사람인지" 만 확인하고, 글을 저장하는 backend 에는 사용자 신원을 절대 넘기지 않습니다. 자세한 설계와 한계는 [docs/anonymity.md](docs/anonymity.md) 를 봅니다.

## 아키텍처

```
브라우저
    | HTTPS
    v
portal :3000 (SvelteKit BFF, Google 로그인)   --- auth.db (누가 로그인했는지)
    | X-Portal-Secret, 글 내용만 (신원 없음)
    v
backend :8080 (Go)                             --- board.db (글, 댓글)
```

## 기술 스택

| 구분 | 스택 |
|---|---|
| backend | Go 1.25.8, stdlib net/http, log/slog |
| portal | SvelteKit 2 (Svelte 5, adapter-node), bun, tailwind v4, better-auth (Google OAuth) |
| DB | SQLite (portal: `node:sqlite`, backend: `modernc.org/sqlite`, 순수 Go) |
| 배포 | Docker 멀티 스테이지, docker compose, GitHub Actions (태그 push 시 GHCR) |

## 저장소 구조

```
secret-board/
|- Makefile              # 개발 태스크 진입점
|- mise.toml             # 툴체인 버전 고정 (go, bun, node)
|- VERSION               # 버전 단일 소스 (SECRET_BOARD_VERSION)
|- CLAUDE.md             # Claude Code 용 프로젝트 안내
|- .claude/docs/         # Claude Code 용 세부 규칙
|- .github/workflows/    # CI (test, build, docker)
|- docs/                 # 설계 문서 (anonymity, google-oauth, local-share)
|- scripts/              # 개발 스크립트 (dev.sh)
|- src/backend/          # Go 백엔드
|  |- cmd/server/        # API 서버
|  |- internal/board/    # 게시글 / 댓글 저장소 (board.db)
|  `- internal/httpapi/  # 라우팅, portal 시크릿 검사
|- src/frontend/portal/  # SvelteKit 포탈
`- docker/               # Dockerfile, 이미지 빌드 스크립트, 배포 compose
```

## 개발

```bash
mise install                                   # go / bun / node 버전 맞추기
make install                                   # portal 의존성
cp src/frontend/portal/.env.example src/frontend/portal/.env   # 값 채우기 (docs/google-oauth.md)

make dev                                       # backend :8080 + portal :5173 (Ctrl-C 로 둘 다 종료)
make test
```

## 배포

```bash
cd docker/deploy/secret-board
cp .env.example .env                           # 값 채우기
docker compose up -d --build                   # 이미지 빌드 + caddy(HTTPS :8443) / portal / backend 실행
```

회사 도메인 없이 내 PC 에서 띄워 동료와 쓰는 방법은 [docs/local-share.md](docs/local-share.md) 를 봅니다.

`VERSION` 을 올린 커밋에 `vX.Y.Z` 태그를 push 하면 GitHub Actions 가 `ghcr.io/<owner>/secret-board` 로 이미지를 올립니다.

## 로드맵

1. [x] 뼈대: SvelteKit 포탈, Go 백엔드, Docker, CI
2. [x] Google OAuth 로그인 + corelinesoft.com 도메인 제한 (실제 OAuth 키로 확인 필요)
3. [x] Go 백엔드: 게시글 / 댓글 API (작성자 정보 없는 스키마, 시각은 시간 단위로 뭉갬)
4. [x] portal 과 backend 연결 (BFF, 공유 시크릿): 글 목록 / 쓰기 / 상세 / 댓글
5. [ ] 익명성 강화: 시간 뭉개기, 글별 수정 / 삭제 토큰
6. [ ] 디자인 (hairline 테마: 1px 선과 면 단계로 깊이 표현)
7. [ ] 배포
