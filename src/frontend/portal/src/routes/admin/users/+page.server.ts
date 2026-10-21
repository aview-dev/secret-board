import { fail } from '@sveltejs/kit';
import { isRole } from '$lib/roles';
import { listUsers, setRole } from '$lib/server/users';
import type { Actions, PageServerLoad } from './$types';

// 관리자 확인은 hooks.server.ts 의 /admin 가드가 한다.
export const load: PageServerLoad = ({ locals }) => ({
	users: listUsers(),
	myId: locals.user!.id
});

export const actions: Actions = {
	setRole: async ({ request, locals }) => {
		const form = await request.formData();
		const id = form.get('id');
		const role = form.get('role');
		if (typeof id !== 'string' || !isRole(role))
			return fail(400, { message: '잘못된 요청입니다.' });

		const target = listUsers().find((u) => u.id === id);
		if (!target) return fail(404, { message: '사용자를 찾을 수 없습니다.' });
		if (target.fixed)
			return fail(400, { message: 'ADMIN_EMAILS 에 있는 계정은 바꿀 수 없습니다.' });
		// 스스로 강등하면 되돌릴 사람이 없을 수 있어서 막는다.
		if (target.id === locals.user!.id) return fail(400, { message: '내 권한은 바꿀 수 없습니다.' });

		setRole(id, role);
		return {
			message: `${target.email} 을(를) ${role === 'admin' ? '관리자' : '일반'}로 바꿨습니다.`
		};
	}
};
