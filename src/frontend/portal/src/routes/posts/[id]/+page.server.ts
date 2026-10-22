import { error, fail, redirect } from '@sveltejs/kit';
import { BackendError, createComment, deletePost, getPost, vote } from '$lib/server/backend';
import { renderMarkdown } from '$lib/server/markdown';
import { voterHash } from '$lib/server/votes';
import { MAX_COMMENT_LEN } from '$lib/board';
import type { Actions, PageServerLoad } from './$types';

function parseId(raw: string): number {
	const id = Number(raw);
	if (!Number.isSafeInteger(id) || id <= 0) error(404, '글을 찾을 수 없습니다.');
	return id;
}

// 로그인은 hooks 가 보장한다. userId 는 voter 해시를 만들 때만 쓰고 backend 로 보내지 않는다 (규칙 9).
export const load: PageServerLoad = async ({ params, locals }) => {
	const id = parseId(params.id);
	let data;
	try {
		data = await getPost(id, voterHash(locals.user!.id, id));
	} catch (e) {
		if (e instanceof BackendError && e.status === 404) error(404, '글을 찾을 수 없습니다.');
		error(502, '게시판 서버에 연결할 수 없습니다.');
	}
	// 본문은 마크다운 원문 그대로 저장되어 있다. 보여줄 때마다 서버에서 HTML 로 바꾼다.
	return {
		...data,
		bodyHtml: renderMarkdown(data.post.body ?? ''),
		comments: data.comments.map((c) => ({ ...c, bodyHtml: renderMarkdown(c.body) }))
	};
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
	},

	// value: 1 = 좋아요, -1 = 싫어요, 0 = 취소. 같은 버튼을 다시 누르면 화면이 0 을 보낸다.
	vote: async ({ params, request, locals }) => {
		const id = parseId(params.id);
		const raw = (await request.formData()).get('value');
		const value = raw === null ? NaN : Number(raw);
		if (value !== 1 && value !== -1 && value !== 0)
			return fail(400, { voteMessage: '잘못된 요청입니다.' });

		try {
			await vote(id, voterHash(locals.user!.id, id), value);
		} catch (e) {
			if (e instanceof BackendError && e.status === 404) error(404, '글을 찾을 수 없습니다.');
			return fail(502, { voteMessage: '투표를 저장할 수 없습니다.' });
		}
		return { ok: true };
	},

	// 관리자만. 버튼을 숨기는 것만으로는 막을 수 없어서 (form 은 누구나 보낼 수 있다) 여기서 다시 확인한다.
	delete: async ({ params, locals }) => {
		const id = parseId(params.id);
		if (locals.role !== 'admin') error(403, '관리자만 삭제할 수 있습니다.');

		try {
			await deletePost(id);
		} catch (e) {
			if (e instanceof BackendError && e.status === 404) error(404, '글을 찾을 수 없습니다.');
			return fail(502, { deleteMessage: '글을 삭제할 수 없습니다.' });
		}
		redirect(303, '/');
	}
};
