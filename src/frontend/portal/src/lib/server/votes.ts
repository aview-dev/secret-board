// 좋아요 / 싫어요의 voter 해시를 만드는 곳 (docs/anonymity.md 규칙 9).
//
// backend 는 "같은 사람이 같은 글에 두 번 누르지 않았는지" 만 알면 된다. 그래서 userId 대신
// HMAC(투표 키, userId:postId) 를 보낸다. postId 가 섞여 있어 같은 사람이라도 글마다 값이 달라,
// board.db 만으로는 한 사람의 표끼리 묶을 수도, 누구의 표인지 알 수도 없다.
// 투표 키는 BETTER_AUTH_SECRET 에서 용도 이름을 섞어 유도한다 (로그인용 키를 그대로 쓰지 않는다).
// 이 키가 바뀌면 기존 표와 연결이 끊겨 한 번 더 누를 수 있게 된다.

import { createHmac } from 'node:crypto';
import { env } from '$env/dynamic/private';

let voteKey: Buffer | undefined;

function key(): Buffer {
	if (!voteKey) {
		if (!env.BETTER_AUTH_SECRET) throw new Error('BETTER_AUTH_SECRET is not set');
		voteKey = createHmac('sha256', env.BETTER_AUTH_SECRET).update('secret-board vote v1').digest();
	}
	return voteKey;
}

export function voterHash(userId: string, postId: number): string {
	return createHmac('sha256', key()).update(`${userId}:${postId}`).digest('hex');
}
