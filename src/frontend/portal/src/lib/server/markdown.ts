// 글 / 댓글 본문을 마크다운에서 HTML 로 바꾸는 곳. 서버에서만 돌려 브라우저로 라이브러리를 보내지 않는다.
//
// 안전 (XSS): html: false 라 본문에 섞인 <script> 같은 HTML 은 실행되지 않고 글자로 나온다.
// markdown-it 은 javascript: / vbscript: / data: 링크도 기본으로 막는다 (validateLink).
// 그래서 이 함수의 결과만 {@html} 로 넣는다. 다른 문자열을 {@html} 에 넣지 않는다.
//
// 익명성 (docs/anonymity.md 규칙 10): 이미지 문법은 끈다. 외부 이미지를 허용하면 글쓴이가 자기 서버
// 이미지를 넣어 두고 읽는 사람의 IP 와 읽은 시각을 모을 수 있다 (추적 픽셀).
// 링크는 noreferrer 로 열어 게시판 주소가 상대 사이트로 넘어가지 않게 한다.

import MarkdownIt from 'markdown-it';

const md = new MarkdownIt({
	html: false,
	linkify: true, // 맨 URL 도 링크로
	breaks: true // 엔터 한 번도 줄바꿈으로 (지금까지 whitespace-pre-wrap 으로 보이던 모양과 맞춘다)
}).disable('image');

md.renderer.rules.link_open = (tokens, idx, options, _env, self) => {
	tokens[idx].attrSet('target', '_blank');
	tokens[idx].attrSet('rel', 'noopener noreferrer nofollow');
	return self.renderToken(tokens, idx, options);
};

export function renderMarkdown(src: string): string {
	return md.render(src);
}
