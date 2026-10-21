// auth.db 의 user 테이블을 직접 읽고 쓰는 곳. 관리자 화면과 역할 판단에 쓴다.
// 이메일, 역할, 가입일만 다룬다. 로그인 시각 같은 활동 기록은 보여주지 않는다 (docs/anonymity.md 규칙 8).

import { db, isFixedAdmin } from './auth';
import { isRole, type Role } from '$lib/roles';

export type UserRow = { id: string; email: string; role: Role; fixed: boolean; createdAt: string };

// 매 요청마다 DB에서 읽는다. 세션 쿠키 캐시(60초)에 든 값을 쓰면 강등이 늦게 반영되기 때문이다.
export function roleOf(user: { id: string; email: string }): Role {
	if (isFixedAdmin(user.email)) return 'admin';
	const row = db!.prepare('SELECT role FROM user WHERE id = ?').get(user.id) as
		{ role: string | null } | undefined;
	return row?.role === 'admin' ? 'admin' : 'user';
}

export function listUsers(): UserRow[] {
	const rows = db!
		.prepare('SELECT id, email, role, createdAt FROM user ORDER BY createdAt')
		.all() as {
		id: string;
		email: string;
		role: string | null;
		createdAt: string;
	}[];
	return rows.map((r) => ({
		id: r.id,
		email: r.email,
		role: isFixedAdmin(r.email) ? 'admin' : isRole(r.role) ? r.role : 'user',
		fixed: isFixedAdmin(r.email),
		createdAt: r.createdAt
	}));
}

export function setRole(id: string, role: Role): void {
	db!
		.prepare('UPDATE user SET role = ?, updatedAt = ? WHERE id = ?')
		.run(role, new Date().toISOString(), id);
}
