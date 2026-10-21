// Go backend 를 부르는 곳. portal 에서 backend 로 가는 요청은 전부 여기를 지난다.
//
// 익명성 (docs/anonymity.md 규칙 2): 사용자 정보는 헤더에도 본문에도 넣지 않는다.
// 그래서 SvelteKit 의 event.fetch (쿠키를 따라 보낼 수 있다) 대신 전역 fetch 를 쓰고,
// 이 모듈의 함수는 locals 나 event 를 인자로 받지 않는다.

import { env } from '$env/dynamic/private';
import type { Comment, Post } from '$lib/board';

const BASE_URL = env.SECRET_BOARD_BACKEND_URL || 'http://localhost:8080';

export class BackendError extends Error {
	constructor(
		public status: number,
		public code: string
	) {
		super(`backend ${status} ${code}`);
	}
}

async function call<T>(method: 'GET' | 'POST', path: string, body?: unknown): Promise<T> {
	const headers: Record<string, string> = {
		'X-Portal-Secret': env.SECRET_BOARD_BACKEND_SECRET || ''
	};
	if (body !== undefined) headers['Content-Type'] = 'application/json';

	const res = await fetch(BASE_URL + path, {
		method,
		headers,
		body: body === undefined ? undefined : JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	if (!res.ok) throw new BackendError(res.status, data.error ?? 'unknown');
	return data as T;
}

export const listPosts = () => call<{ posts: Post[] }>('GET', '/api/v1/posts');

export const getPost = (id: number) =>
	call<{ post: Post; comments: Comment[] }>('GET', `/api/v1/posts/${id}`);

export const createPost = (title: string, body: string) =>
	call<Post>('POST', '/api/v1/posts', { title, body });

export const createComment = (postId: number, body: string) =>
	call<Comment>('POST', `/api/v1/posts/${postId}/comments`, { body });
