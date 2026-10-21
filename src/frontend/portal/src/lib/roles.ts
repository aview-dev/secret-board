// 브라우저와 서버가 같이 쓰는 역할 정의. 서버 전용 코드는 $lib/server 에 둔다.
export const ROLES = ['admin', 'user'] as const;
export type Role = (typeof ROLES)[number];

export const ROLE_LABEL: Record<Role, string> = { admin: '관리자', user: '일반' };

export function isRole(v: unknown): v is Role {
	return typeof v === 'string' && (ROLES as readonly string[]).includes(v);
}
