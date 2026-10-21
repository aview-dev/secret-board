// 버튼 / 입력 공용 클래스. aviewops 의 shadcn Button variant 중 쓰는 것만 문자열로 옮겼다.
// 강조색(primary)은 화면당 주요 버튼 하나에만 쓰고, 나머지는 outline / ghost 로 둔다.

const BUTTON_BASE =
	'inline-flex h-8 shrink-0 items-center justify-center gap-1.5 rounded-md px-3 text-sm font-medium whitespace-nowrap transition-all outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 [&_svg]:size-4 [&_svg]:shrink-0';

export const BUTTON_PRIMARY = `${BUTTON_BASE} bg-primary text-primary-foreground hover:brightness-90`;
export const BUTTON_OUTLINE = `${BUTTON_BASE} border bg-control text-foreground hover:brightness-95`;
export const BUTTON_GHOST = `${BUTTON_BASE} text-muted-foreground hover:bg-muted hover:text-foreground`;

export const INPUT =
	'block w-full rounded-md border bg-background px-3 py-2 text-sm outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50';

// 바닥(app-ground) 위에 뜨는 흰 패널. 그림자 대신 1px 선으로 구분한다.
export const PANEL = 'overflow-hidden rounded-md border bg-panel';
