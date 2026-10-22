// Go backend 를 부르는 곳. portal 에서 backend 로 가는 요청은 전부 여기를 지난다.
//
// 익명성 (docs/anonymity.md 규칙 2): 사용자 정보는 헤더에도 본문에도 넣지 않는다.
// 그래서 SvelteKit 의 event.fetch (쿠키를 따라 보낼 수 있다) 대신 전역 fetch 를 쓰고,
// 이 모듈의 함수는 locals 나 event 를 인자로 받지 않는다.
// 예외는 투표용 voter 해시와 신고용 reporter 해시뿐이다 (규칙 9, 11, lib/server/votes.ts). userId 자체는 받지 않는다.

import { env } from '$env/dynamic/private';
import type { Comment, Post, ReportReason, ReportedPost, Votes } from '$lib/board';

const BASE_URL = env.SECRET_BOARD_BACKEND_URL || 'http://localhost:8080';

export class BackendError extends Error {
	constructor(
		public status: number,
		public code: string
	) {
		super(`backend ${status} ${code}`);
	}
}

async function call<T>(
	method: 'GET' | 'POST' | 'PUT' | 'DELETE',
	path: string,
	body?: unknown,
	extraHeaders?: Record<string, string>
): Promise<T> {
	const headers: Record<string, string> = {
		...extraHeaders,
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

export const getPost = (id: number, voter: string) =>
	call<{ post: Post; comments: Comment[]; myVote: number }>(
		'GET',
		`/api/v1/posts/${id}`,
		undefined,
		{ 'X-Voter': voter }
	);

export const createPost = (title: string, body: string) =>
	call<Post>('POST', '/api/v1/posts', { title, body });

// 관리자 삭제. 관리자 확인은 부르는 쪽(action)이 하고, 누가 지우는지는 보내지 않는다 (규칙 7).
// 성공하면 204 라 본문이 없다. call 의 res.json() 실패는 {} 로 넘어간다.
export const deletePost = (id: number) => call<unknown>('DELETE', `/api/v1/posts/${id}`);

export const createComment = (postId: number, body: string) =>
	call<Comment>('POST', `/api/v1/posts/${postId}/comments`, { body });

// 관리자 삭제. deletePost 와 같다. 글 id 도 함께 보내 다른 글의 댓글이 지워지지 않게 한다.
export const deleteComment = (postId: number, commentId: number) =>
	call<unknown>('DELETE', `/api/v1/posts/${postId}/comments/${commentId}`);

export const vote = (postId: number, voter: string, value: number) =>
	call<Votes>('PUT', `/api/v1/posts/${postId}/vote`, { voter, value });

// 신고. reporter 는 lib/server/votes.ts 의 reporterHash 만 넘긴다 (규칙 11).
export const reportPost = (postId: number, reporter: string, reason: ReportReason) =>
	call<unknown>('PUT', `/api/v1/posts/${postId}/report`, { reporter, reason });

// 관리자 신고 목록 / 신고 초기화. 관리자 확인은 부르는 쪽이 한다 (규칙 7).
export const listReported = () => call<{ posts: ReportedPost[] }>('GET', '/api/v1/reports');

export const clearReports = (postId: number) =>
	call<unknown>('DELETE', `/api/v1/posts/${postId}/reports`);
