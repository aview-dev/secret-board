// 브라우저와 서버가 같이 쓰는 게시판 타입과 제한값. 값은 backend (internal/board/store.go) 와 맞춘다.
export const MAX_TITLE_LEN = 100;
export const MAX_BODY_LEN = 10000;
export const MAX_COMMENT_LEN = 2000;

export type Post = {
	id: number;
	title: string;
	body?: string; // 목록에서는 빠진다
	createdAt: string; // 시간 단위로 뭉갠 시각 (docs/anonymity.md 규칙 4)
	commentCount: number;
};

export type Comment = {
	id: number;
	postId: number;
	body: string;
	createdAt: string;
};

// 시각은 시간 단위까지만 저장되므로 분은 보여주지 않는다.
const fmt = new Intl.DateTimeFormat('ko-KR', {
	month: 'long',
	day: 'numeric',
	hour: 'numeric'
});

export function formatHour(iso: string): string {
	return fmt.format(new Date(iso)) + '쯤';
}
