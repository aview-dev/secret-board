import type { LayoutServerLoad } from './$types';

// 상단 바에 필요한 것만 내려준다. 로그인 전(/login)에는 null.
export const load: LayoutServerLoad = ({ locals }) => ({
	me: locals.user ? { email: locals.user.email, role: locals.role! } : null
});
