<script lang="ts">
	import { page } from '$app/state';
	import { authClient } from '$lib/auth-client';

	let { data } = $props();
	let loading = $state(false);

	// better-auth는 로그인 실패 시 ?error=... 를 붙여 돌려보낸다.
	const error = $derived(page.url.searchParams.get('error'));

	async function signIn() {
		loading = true;
		// 같은 사이트 안의 경로만 허용해서 open redirect를 막는다.
		const raw = page.url.searchParams.get('next');
		const next = raw && raw.startsWith('/') && !raw.startsWith('//') ? raw : '/';
		await authClient.signIn.social({
			provider: 'google',
			callbackURL: next,
			errorCallbackURL: '/login'
		});
	}
</script>

<main class="flex min-h-svh items-center justify-center bg-neutral-50 px-6">
	<div class="w-full max-w-md rounded-xl border border-neutral-200 bg-white p-10 text-center">
		<h1 class="text-2xl font-semibold tracking-tight">secret-board</h1>
		<p class="mt-3 text-sm text-neutral-500">
			@{data.domain} 계정으로만 로그인할 수 있는 익명 게시판
		</p>

		{#if error}
			<p class="mt-6 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">
				로그인할 수 없습니다. @{data.domain} 계정인지 확인해 주세요.
			</p>
		{/if}

		<div class="mt-8">
			{#if data.googleConfigured}
				<button
					class="h-10 w-full rounded-md bg-neutral-900 text-sm font-medium text-white hover:bg-neutral-800 disabled:opacity-50"
					onclick={signIn}
					disabled={loading}
				>
					{loading ? '로그인 중…' : 'Google로 로그인'}
				</button>
			{:else}
				<p class="rounded-md border border-dashed px-3 py-4 text-sm text-neutral-500">
					GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET이 설정되지 않았습니다.
				</p>
			{/if}
		</div>
	</div>
</main>
