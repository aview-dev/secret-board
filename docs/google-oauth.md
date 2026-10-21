# Google OAuth 설정

secret-board 는 Google 로그인만 지원하고, `ALLOWED_DOMAIN` (기본 `corelinesoft.com`) 계정만 받습니다.

## 도메인 제한 (세 겹)

1. **동의 화면 사용자 유형 "내부(Internal)"**: Google 이 Workspace 조직 계정만 받는다.
2. **ID 토큰 `hd` 클레임 검증**: better-auth `google.hd` 옵션. 로그인 URL 의 `hd=` 는 힌트일 뿐이라 우회 가능하므로, 서버가 Google 이 서명한 토큰의 `hd` 를 다시 확인한다. 개인 gmail 은 `hd` 가 없어 거절된다.
3. **계정 생성 직전 이메일 검사**: `@corelinesoft.com` 으로 끝나고 `email_verified` 인지 확인 (`databaseHooks.user.create.before`). 요청마다 세션을 읽을 때도 도메인을 다시 본다 (`hooks.server.ts`).

## 클라이언트 만들기

1. [Google Cloud Console](https://console.cloud.google.com) 에 **회사 계정으로** 들어가 프로젝트를 만든다.
2. **API 및 서비스 > OAuth 동의 화면**: 사용자 유형 **내부**. 권한이 없으면 IT 관리자에게 요청한다.
3. **사용자 인증 정보 > OAuth 클라이언트 ID**: 유형 **웹 애플리케이션**.
   - 승인된 리디렉션 URI
     - 개발: `http://localhost:5173/api/auth/callback/google`
     - 운영: `${PORTAL_ORIGIN}/api/auth/callback/google`
4. 클라이언트 ID / 시크릿을
   - 개발: `src/frontend/portal/.env`
   - 운영: `docker/deploy/secret-board/.env`
   에 넣는다. 두 파일 모두 git 에 올라가지 않는다.
