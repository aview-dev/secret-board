# 내 PC 에서 띄워 동료와 함께 쓰기

회사 도메인이 아직 없을 때, 개인 PC (예: `10.77.4.34`) 에서 docker compose 로 띄우고 사내 동료가 접속하게 하는 방법입니다.

```
동료 브라우저 -> https://10-77-4-34.sslip.io -> 내 PC :443
                                                 `- caddy (HTTPS, tls internal) -> portal :3000 -> backend :8080
```

- `sslip.io` 는 `<IP 를 - 로 이은 것>.sslip.io` 를 그 IP 로 해석해 주는 공개 DNS 입니다. Google 은 IP 주소를 리디렉션 URI 로 받지 않지만, 이 이름은 공개 도메인이라 등록할 수 있습니다.
- 인증서는 caddy 가 자체 CA 로 서명합니다 (`tls internal`). 브라우저가 처음 한 번 경고를 띄우며, "고급 > 계속" 으로 들어갑니다.

## 준비물

- Docker (Docker Desktop 또는 Docker Engine + compose 플러그인)
- PC 의 80, 443 포트가 비어 있고, 방화벽에서 사내 접속을 허용

## 1. Google OAuth 클라이언트에 주소 추가

기존 클라이언트에 추가하거나 `secret-board share` 클라이언트를 새로 만듭니다 ([google-oauth.md](google-oauth.md)).

- 승인된 JavaScript 원본: `https://10-77-4-34.sslip.io`
- 승인된 리디렉션 URI: `https://10-77-4-34.sslip.io/api/auth/callback/google`

## 2. 설정과 실행

```bash
git clone https://github.com/aview-dev/secret-board.git
cd secret-board/docker/deploy/secret-board
cp .env.example .env
# .env 편집: SITE_ADDRESS, BETTER_AUTH_SECRET, SECRET_BOARD_BACKEND_SECRET,
#           GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET
docker compose up -d --build
```

비밀값은 `openssl rand -base64 32` 로 만듭니다 (Windows 는 Git Bash 에서 실행).

## 3. 확인

- 내 PC: `https://10-77-4-34.sslip.io` 접속 > 경고 통과 > Google 로그인
- 동료 PC: 같은 주소. 접속이 안 되면 PC 방화벽의 443 인바운드 허용을 확인

## 운영

| 할 일 | 명령 |
|---|---|
| 로그 보기 | `docker compose logs -f portal` |
| 코드 갱신 후 다시 띄우기 | `git pull && docker compose up -d --build` |
| 끄기 | `docker compose down` (데이터는 볼륨에 남음) |
| 데이터까지 삭제 | `docker compose down -v` |

## 주의

- PC 를 끄면 서비스도 꺼집니다. PC IP 가 바뀌면 `SITE_ADDRESS` 와 Google 클라이언트 주소를 함께 바꿉니다.
- 회사가 개인 PC 에서 서비스를 띄우는 것을 허용하는지 확인합니다.
- 운영자 (PC 주인) 는 DB 에 접근할 수 있습니다. 익명성의 한계는 [anonymity.md](anonymity.md) 를 참고합니다.
- ngrok, Cloudflare Tunnel 같은 외부 터널은 쓰지 않습니다. 외부 업체가 글 내용과 접속 IP 를 볼 수 있습니다.
