<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { authClient } from '$lib/auth-client';

	let { data } = $props();

	async function signOut() {
		await authClient.signOut();
		goto(resolve('/login'), { invalidateAll: true });
	}
</script>

<main class="mx-auto max-w-2xl px-4 py-10">
	<div class="flex items-center justify-between">
		<h1 class="text-xl font-semibold tracking-tight">secret-board</h1>
		<button class="text-sm text-neutral-500 hover:text-neutral-900" onclick={signOut}>
			로그아웃
		</button>
	</div>
	<p class="mt-6 text-sm text-neutral-500">
		{data.email} 로 로그인했습니다. 게시판은 다음 단계에서 만듭니다.
	</p>
</main>
