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

// ADMIN_EMAILS(쉼표로 구분)에 적힌 계정은 DB의 role 과 상관없이 항상 관리자다.
// 관리자 화면에서 강등할 수 없어서, 관리자가 한 명도 남지 않는 상황을 막는다.
const ADMIN_EMAILS = new Set(
	(env.ADMIN_EMAILS || '')
		.split(',')
		.map((e) => e.trim().toLowerCase())
		.filter(Boolean)
);

export function isFixedAdmin(email: string): boolean {
	return ADMIN_EMAILS.has(email.toLowerCase());
}

export const googleConfigured = Boolean(env.GOOGLE_CLIENT_ID && env.GOOGLE_CLIENT_SECRET);

export const db = building ? undefined : new DatabaseSync(env.AUTH_DB_PATH || 'auth.db');

const options = {
	baseURL: env.BETTER_AUTH_URL || undefined,
	secret: env.BETTER_AUTH_SECRET,
	// `vite build`는 서버 모듈을 import해서 분석한다. 그때 DB 파일이 생기지 않도록
	// 빌드 중에는 database를 비워 둔다 (better-auth가 메모리 어댑터로 대신한다).
	database: db,
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
	user: {
		// 역할. 기본은 user 이고, 관리자 화면에서만 바꾼다 (input: false 라서 가입 요청으로는 못 정한다).
		additionalFields: {
			role: { type: 'string' as const, required: false, defaultValue: 'user', input: false }
		}
	},
	databaseHooks: {
		user: {
			create: {
				before: async (user: { email: string; emailVerified: boolean }) => {
					if (!isAllowedEmail(user.email) || !user.emailVerified) {
						throw new APIError('FORBIDDEN', { message: 'domain_not_allowed' });
					}
				}
			}
		},
		// 세션에 User-Agent를 남기지 않는다. 로그인 시각과 기기 정보를 글 작성 시각과
		// 맞춰 보는 단서가 되기 때문이다 (docs/anonymity.md 규칙 5).
		session: {
			create: {
				before: async (session: Record<string, unknown>) => ({
					data: { ...session, ipAddress: null, userAgent: null }
				})
			}
		}
	},
	// IP도 수집하지 않는다. 이 때문에 better-auth의 IP 기준 rate limit은 동작하지 않는다.
	advanced: { ipAddress: { disableIpTracking: true } },
	session: { cookieCache: { enabled: true, maxAge: 60 } }
};

// 서버가 뜰 때 auth 테이블(user, session, account, verification)을 만들고, 추가된 컬럼(user.role)을 붙인다.
if (!building) {
	const { runMigrations } = await getMigrations(options);
	await runMigrations();
	// 위 설정 전에 만들어진 세션에 남아 있던 값을 지운다.
	db!.exec('UPDATE session SET ipAddress = NULL, userAgent = NULL');
}

export const auth = betterAuth(options);
