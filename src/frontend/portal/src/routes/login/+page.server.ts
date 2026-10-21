import { ALLOWED_DOMAIN, googleConfigured } from '$lib/server/auth';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => ({ googleConfigured, domain: ALLOWED_DOMAIN });
