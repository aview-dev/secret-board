# Google OAuth 설정

secret-board 는 Google 로그인만 지원하고, `ALLOWED_DOMAIN` (기본 `corelinesoft.com`) 계정만 받습니다.

## 도메인 제한 (세 겹)

1. **동의 화면 대상 "내부"**: Google 이 Workspace 조직 계정만 받는다.
2. **ID 토큰 `hd` 클레임 검증**: better-auth `google.hd` 옵션. 로그인 URL 의 `hd=` 는 힌트일 뿐이라 우회 가능하므로, 서버가 Google 이 서명한 토큰의 `hd` 를 다시 확인한다. 개인 gmail 은 `hd` 가 없어 거절된다.
3. **계정 생성 직전 이메일 검사**: `@corelinesoft.com` 으로 끝나고 `email_verified` 인지 확인 (`databaseHooks.user.create.before`). 요청마다 세션을 읽을 때도 도메인을 다시 본다 (`hooks.server.ts`).

## 1. 프로젝트 만들기

1. [Google Cloud Console](https://console.cloud.google.com) 에 **회사 계정으로** 로그인한다.
2. 왼쪽 위 프로젝트 선택 > **새 프로젝트**.
   - 프로젝트 이름: `secret-board`
   - **조직: `corelinesoft.com`**. 조직 밖에 만들면 다음 단계에서 "내부" 를 고를 수 없다.
3. 만든 뒤 왼쪽 위에서 해당 프로젝트가 선택되어 있는지 확인한다.

조직에 프로젝트를 만들 권한이 없으면 IT 관리자에게 요청한다.

## 2. OAuth 동의 화면 (Google 인증 플랫폼)

**API 및 서비스 > OAuth 동의 화면** (Google 인증 플랫폼) > **시작하기**.

| 단계 | 값 |
|---|---|
| 앱 정보 | 앱 이름 `secret-board`, 사용자 지원 이메일 = 본인 |
| 대상 | **내부** (회색이면 프로젝트가 조직 밖에 있다) |
| 연락처 정보 | 본인 메일 |
| 완료 | 정책 동의 > 만들기 |

데이터 액세스(범위)는 기본값 그대로 둔다. better-auth 는 `openid`, `email`, `profile` 만 요청하고, 내부 앱은 Google 심사가 필요 없다.

## 3. 클라이언트 만들기

**클라이언트 > 클라이언트 만들기**.

| 항목 | 개발 | 운영 |
|---|---|---|
| 유형 | 웹 애플리케이션 | 웹 애플리케이션 |
| 이름 | `secret-board dev` | `secret-board prod` |
| 승인된 JavaScript 원본 | `http://localhost:5173` | `${PORTAL_ORIGIN}` |
| 승인된 리디렉션 URI | `http://localhost:5173/api/auth/callback/google` | `${PORTAL_ORIGIN}/api/auth/callback/google` |

만들면 클라이언트 ID 와 보안 비밀이 표시된다. **보안 비밀은 생성 직후에만 볼 수 있을 수 있으니** 바로 복사하거나 JSON 을 내려받는다.

## 4. 값 넣기

- 개발: `src/frontend/portal/.env`
- 운영: `docker/deploy/secret-board/.env`

```
GOOGLE_CLIENT_ID=xxxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-xxxx
```

두 파일 모두 git 에 올라가지 않는다. 값을 바꾼 뒤에는 포탈을 다시 시작한다.

## 원격 서버에서 개발할 때

Google 은 `http://` 리디렉션을 localhost 에만 허용한다. 개발 서버가 원격이면 SSH 포트 포워딩으로 접속한다.

```bash
ssh -L 5173:localhost:5173 <서버주소>
```

## 자주 만나는 에러

| 증상 | 원인 |
|---|---|
| `redirect_uri_mismatch` | 리디렉션 URI 가 정확히 일치하지 않음 (scheme, 포트, 끝 `/`) |
| `403 org_internal` (개인 gmail) | 정상. "내부" 설정이 막은 것 |
| 로그인 버튼 대신 "설정되지 않았습니다" | `.env` 값이 비었거나 포탈을 재시작하지 않음 |
