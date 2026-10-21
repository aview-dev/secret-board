import { fail, redirect } from '@sveltejs/kit';
import { BackendError, createPost } from '$lib/server/backend';
import { MAX_BODY_LEN, MAX_TITLE_LEN } from '$lib/board';
import type { Actions } from './$types';

// 로그인 확인은 hooks.server.ts 의 가드가 한다. 여기서는 locals.user 를 쓰지 않는다 (규칙 2).
export const actions: Actions = {
	default: async ({ request }) => {
		const form = await request.formData();
		const title = String(form.get('title') ?? '').trim();
		const body = String(form.get('body') ?? '').trim();

		// 화면에 바로 알려 주려고 먼저 검사한다. 최종 검사는 backend 가 한 번 더 한다.
		if (!title || !body)
			return fail(400, { title, body, message: '제목과 본문을 모두 적어 주세요.' });
		if (title.length > MAX_TITLE_LEN || body.length > MAX_BODY_LEN)
			return fail(400, { title, body, message: '글이 너무 깁니다.' });

		let id: number;
		try {
			id = (await createPost(title, body)).id;
		} catch (e) {
			const message =
				e instanceof BackendError && e.status === 400
					? '글을 저장할 수 없습니다. 내용을 확인해 주세요.'
					: '게시판 서버에 연결할 수 없습니다.';
			return fail(502, { title, body, message });
		}
		redirect(303, `/posts/${id}`);
	}
};
