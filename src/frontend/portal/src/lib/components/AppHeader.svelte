<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { authClient } from '$lib/auth-client';
	import type { Role } from '$lib/roles';

	let { role }: { role: Role } = $props();

	async function signOut() {
		await authClient.signOut();
		goto(resolve('/login'), { invalidateAll: true });
	}
</script>

<header class="border-b border-neutral-200 bg-white">
	<div class="mx-auto flex h-14 max-w-2xl items-center justify-between px-4">
		<a href={resolve('/')} class="font-semibold tracking-tight">secret-board</a>
		<nav class="flex items-center gap-4 text-sm text-neutral-500">
			{#if role === 'admin'}
				<span class="rounded bg-neutral-900 px-1.5 py-0.5 text-xs font-medium text-white">
					관리자
				</span>
			{/if}
			{#if role === 'admin'}
				<a href={resolve('/admin/users')} class="hover:text-neutral-900">사용자 관리</a>
			{/if}
			<a href={resolve('/me')} class="hover:text-neutral-900">내 페이지</a>
			<button class="hover:text-neutral-900" onclick={signOut}>로그아웃</button>
		</nav>
	</div>
</header>
