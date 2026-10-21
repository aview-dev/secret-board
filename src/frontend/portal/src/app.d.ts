// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
import type { auth } from '$lib/server/auth';
import type { Role } from '$lib/roles';

type Session = typeof auth.$Infer.Session;

declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			user: Session['user'] | null;
			session: Session['session'] | null;
			// 관리자 여부. backend 에는 관리자 전용 요청에만 전달한다 (docs/anonymity.md 규칙 7).
			role: Role | null;
		}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
