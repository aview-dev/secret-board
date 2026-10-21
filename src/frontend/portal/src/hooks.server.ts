import { redirect, type Handle } from '@sveltejs/kit';
import { building } from '$app/environment';
import { sequence } from '@sveltejs/kit/hooks';
import { svelteKitHandler } from 'better-auth/svelte-kit';
import { auth, isAllowedEmail } from '$lib/server/auth';

const handleAuth: Handle = async ({ event, resolve }) => {
	const result = await auth.api.getSession({ headers: event.request.headers });
	// 세션이 있어도 도메인이 맞지 않으면 로그인하지 않은 것으로 본다 (ALLOWED_DOMAIN을 바꾼 경우 대비).
	const ok = result && isAllowedEmail(result.user.email);
	event.locals.user = ok ? result.user : null;
	event.locals.session = ok ? result.session : null;
	return resolve(event);
};

// 기본 거부. /login만 공개이고 나머지는 로그인이 필요하다.
// route.id가 없는 요청(/api/auth/* 등)은 아래 better-auth 핸들러로 넘어간다.
const handleGuard: Handle = async ({ event, resolve }) => {
	const id = event.route.id;
	if (!id) return resolve(event);

	const isPublic = id === '/login';
	if (!event.locals.user && !isPublic) {
		const next = encodeURIComponent(event.url.pathname + event.url.search);
		redirect(307, `/login?next=${next}`);
	}
	if (event.locals.user && isPublic) redirect(307, '/');
	return resolve(event);
};

const handleBetterAuth: Handle = ({ event, resolve }) =>
	svelteKitHandler({ event, resolve, auth, building });

export const handle = sequence(handleAuth, handleGuard, handleBetterAuth);
