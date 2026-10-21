import { error, fail } from '@sveltejs/kit';
import { BackendError, createComment, getPost } from '$lib/server/backend';
import { MAX_COMMENT_LEN } from '$lib/board';
import type { Actions, PageServerLoad } from './$types';

function parseId(raw: string): number {
	const id = Number(raw);
	if (!Number.isSafeInteger(id) || id <= 0) error(404, '글을 찾을 수 없습니다.');
	return id;
}

export const load: PageServerLoad = async ({ params }) => {
	const id = parseId(params.id);
	try {
		return await getPost(id);
	} catch (e) {
		if (e instanceof BackendError && e.status === 404) error(404, '글을 찾을 수 없습니다.');
		error(502, '게시판 서버에 연결할 수 없습니다.');
	}
};

export const actions: Actions = {
	comment: async ({ params, request }) => {
		const id = parseId(params.id);
		const body = String((await request.formData()).get('body') ?? '').trim();

		if (!body) return fail(400, { body, message: '댓글 내용을 적어 주세요.' });
		if (body.length > MAX_COMMENT_LEN) return fail(400, { body, message: '댓글이 너무 깁니다.' });

		try {
			await createComment(id, body);
		} catch (e) {
			if (e instanceof BackendError && e.status === 404) error(404, '글을 찾을 수 없습니다.');
			return fail(502, { body, message: '댓글을 저장할 수 없습니다.' });
		}
		// 성공하면 load 가 다시 돌아 새 댓글이 보인다.
		return { ok: true };
	}
};
