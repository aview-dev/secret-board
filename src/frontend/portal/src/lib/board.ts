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
	likes: number;
	dislikes: number;
};

// 투표 결과. mine 은 내가 준 표: 1 = 좋아요, -1 = 싫어요, 0 = 안 누름.
export type Votes = { likes: number; dislikes: number; mine: number };

export type Comment = {
	id: number;
	postId: number;
	body: string;
	createdAt: string;
};

// 신고 사유. backend 의 ReportReasons / schema CHECK 와 맞춘다. 자유 입력은 받지 않는다 (규칙 11).
export const REPORT_REASONS = ['spam', 'abuse', 'privacy', 'other'] as const;
export type ReportReason = (typeof REPORT_REASONS)[number];

export const REPORT_REASON_LABEL: Record<ReportReason, string> = {
	spam: '스팸 / 광고',
	abuse: '욕설 / 비방',
	privacy: '개인정보 노출',
	other: '기타'
};

export function isReportReason(v: unknown): v is ReportReason {
	return typeof v === 'string' && (REPORT_REASONS as readonly string[]).includes(v);
}

// 관리자 신고 목록의 한 줄. 누가 신고했는지는 없고 개수만 있다.
export type ReportedPost = {
	id: number;
	title: string;
	createdAt: string;
	reports: number;
	reasons: Partial<Record<ReportReason, number>>;
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
