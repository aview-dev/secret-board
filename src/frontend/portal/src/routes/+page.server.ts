import { error } from '@sveltejs/kit';
import { listPosts } from '$lib/server/backend';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	try {
		return await listPosts();
	} catch {
		error(502, '게시판 서버에 연결할 수 없습니다.');
	}
};
