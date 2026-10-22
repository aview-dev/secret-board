import { error, fail } from '@sveltejs/kit';
import { BackendError, clearReports, deletePost, listReported } from '$lib/server/backend';
import type { Actions, PageServerLoad } from './$types';

// 관리자 확인은 hooks.server.ts 의 /admin 가드가 한다 (load 와 action 모두).
// 이 화면에는 신고 개수와 사유별 개수만 있다. 누가 신고했는지는 backend 도 모른다 (규칙 11).
export const load: PageServerLoad = async () => {
	try {
		return await listReported();
	} catch {
		error(502, '게시판 서버에 연결할 수 없습니다.');
	}
};

function postId(form: FormData): number | null {
	const id = Number(form.get('id'));
	return Number.isSafeInteger(id) && id > 0 ? id : null;
}

export const actions: Actions = {
	// 글을 지우면 신고도 ON DELETE CASCADE 로 함께 지워진다.
	delete: async ({ request }) => {
		const id = postId(await request.formData());
		if (!id) return fail(400, { message: '잘못된 요청입니다.' });
		try {
			await deletePost(id);
		} catch (e) {
			// 그새 지워졌다면 목록만 다시 그리면 된다.
			if (e instanceof BackendError && e.status === 404)
				return { message: '이미 지워진 글입니다.' };
			return fail(502, { message: '글을 삭제할 수 없습니다.' });
		}
		return { message: '글을 삭제했습니다.' };
	},

	// 문제없다고 판단한 글. 신고만 지우고 글은 남긴다.
	clear: async ({ request }) => {
		const id = postId(await request.formData());
		if (!id) return fail(400, { message: '잘못된 요청입니다.' });
		try {
			await clearReports(id);
		} catch {
			return fail(502, { message: '신고를 초기화할 수 없습니다.' });
		}
		return { message: '신고를 초기화했습니다. 글은 그대로 남습니다.' };
	}
};
