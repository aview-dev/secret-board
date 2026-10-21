import type { PageServerLoad } from './$types';

// hooks의 가드 덕분에 여기 오는 사람은 항상 로그인한 회사 계정이다.
export const load: PageServerLoad = ({ locals }) => ({ email: locals.user!.email });
