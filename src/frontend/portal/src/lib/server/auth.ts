// better-auth 서버 설정. Google OAuth가 유일한 로그인 수단이고,
// ALLOWED_DOMAIN(기본 corelinesoft.com) 계정만 들어올 수 있다.
//
// 도메인 검사는 두 겹이다 (Google 동의 화면 "내부" 설정까지 치면 세 겹).
//   1. google.hd: Google ID 토큰의 `hd` 클레임(Workspace 도메인)을 서버에서 검증.
//      개인 gmail 계정은 hd가 없어서 여기서 거절된다.
//   2. databaseHooks: 계정이 만들어지기 직전에 이메일 도메인과 인증 여부를 한 번 더 확인.
//
// 익명성: 이 DB(auth.db)에는 "누가 로그인했는지"만 남는다. 게시글은 Go 백엔드의
// 별도 DB에 저장되고, 거기로는 사용자 신원이 절대 넘어가지 않는다 (docs/anonymity.md).

import { DatabaseSync } from 'node:sqlite';
import { betterAuth, APIError } from 'better-auth';
import { getMigrations } from 'better-auth/db/migration';
import { building } from '$app/environment';
import { env } from '$env/dynamic/private';

export const ALLOWED_DOMAIN = (env.ALLOWED_DOMAIN || 'corelinesoft.com').toLowerCase();

export function isAllowedEmail(email: string | null | undefined): boolean {
	return !!email && email.toLowerCase().endsWith('@' + ALLOWED_DOMAIN);
}

export const googleConfigured = Boolean(env.GOOGLE_CLIENT_ID && env.GOOGLE_CLIENT_SECRET);

const options = {
	baseURL: env.BETTER_AUTH_URL || undefined,
	secret: env.BETTER_AUTH_SECRET,
	// `vite build`는 서버 모듈을 import해서 분석한다. 그때 DB 파일이 생기지 않도록
	// 빌드 중에는 database를 비워 둔다 (better-auth가 메모리 어댑터로 대신한다).
	database: building ? undefined : new DatabaseSync(env.AUTH_DB_PATH || 'auth.db'),
	socialProviders: googleConfigured
		? {
				google: {
					clientId: env.GOOGLE_CLIENT_ID!,
					clientSecret: env.GOOGLE_CLIENT_SECRET!,
					hd: ALLOWED_DOMAIN,
					// 이름과 프로필 사진은 저장하지 않는다. 익명 게시판에 필요 없는 정보는 애초에 모으지 않는다.
					mapProfileToUser: () => ({ name: '', image: undefined })
				}
			}
		: {},
	databaseHooks: {
		user: {
			create: {
				before: async (user: { email: string; emailVerified: boolean }) => {
					if (!isAllowedEmail(user.email) || !user.emailVerified) {
						throw new APIError('FORBIDDEN', { message: 'domain_not_allowed' });
					}
				}
			}
		}
	},
	rateLimit: { window: 60, max: 10 },
	session: { cookieCache: { enabled: true, maxAge: 60 } }
};

// 서버가 뜰 때 auth 테이블(user, session, account, verification)을 만든다.
if (!building) {
	const { runMigrations } = await getMigrations(options);
	await runMigrations();
}

export const auth = betterAuth(options);
