// 좋아요 / 싫어요의 voter 해시와 신고의 reporter 해시를 만드는 곳 (docs/anonymity.md 규칙 9, 11).
//
// backend 는 "같은 사람이 같은 글에 두 번 누르지 않았는지" 만 알면 된다. 그래서 userId 대신
// HMAC(키, userId:postId) 를 보낸다. postId 가 섞여 있어 같은 사람이라도 글마다 값이 달라,
// board.db 만으로는 한 사람의 표끼리 묶을 수도, 누구의 표인지 알 수도 없다.
// 키는 BETTER_AUTH_SECRET 에서 용도 이름을 섞어 유도한다 (로그인용 키를 그대로 쓰지 않는다).
// 투표와 신고는 용도 이름이 달라 키도 다르다. 그래서 같은 사람의 voter 와 reporter 를 맞춰 볼 수 없다.
// 이 키가 바뀌면 기존 표 / 신고와 연결이 끊겨 한 번 더 누를 수 있게 된다.

import { createHmac } from 'node:crypto';
import { env } from '$env/dynamic/private';

const keys = new Map<string, Buffer>();

function key(purpose: string): Buffer {
	let k = keys.get(purpose);
	if (!k) {
		if (!env.BETTER_AUTH_SECRET) throw new Error('BETTER_AUTH_SECRET is not set');
		k = createHmac('sha256', env.BETTER_AUTH_SECRET).update(`secret-board ${purpose} v1`).digest();
		keys.set(purpose, k);
	}
	return k;
}

function hash(purpose: string, userId: string, postId: number): string {
	return createHmac('sha256', key(purpose)).update(`${userId}:${postId}`).digest('hex');
}

// 'vote' 는 처음 만들 때의 용도 이름 그대로다. 바꾸면 기존 표가 전부 끊긴다.
export const voterHash = (userId: string, postId: number) => hash('vote', userId, postId);
export const reporterHash = (userId: string, postId: number) => hash('report', userId, postId);
