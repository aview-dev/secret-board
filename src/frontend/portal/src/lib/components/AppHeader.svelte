<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import LogOut from '@lucide/svelte/icons/log-out';
	import User from '@lucide/svelte/icons/user';
	import Users from '@lucide/svelte/icons/users';
	import { authClient } from '$lib/auth-client';
	import type { Role } from '$lib/roles';

	let { role }: { role: Role } = $props();

	async function signOut() {
		await authClient.signOut();
		goto(resolve('/login'), { invalidateAll: true });
	}

	// 바 위에서는 페이지 색이 안 보이므로 topbar-* 토큰만 쓴다.
	const link =
		'flex items-center gap-1.5 rounded-md px-2 py-1 text-topbar-muted hover:text-topbar-foreground [&_svg]:size-4';
</script>

<header class="border-b border-topbar-edge bg-topbar text-topbar-foreground">
	<div class="mx-auto flex h-12 max-w-3xl items-center justify-between px-4">
		<a href={resolve('/')} class="flex items-center gap-2">
			<span class="h-4 w-[3px] rounded-full bg-topbar-mark" aria-hidden="true"></span>
			<span class="font-semibold tracking-tight">secret-board</span>
		</a>
		<nav class="flex items-center gap-1 text-sm">
			{#if role === 'admin'}
				<span
					class="mr-1 rounded-sm border border-topbar-border px-1.5 py-0.5 text-xs text-topbar-muted"
				>
					관리자
				</span>
				<a href={resolve('/admin/users')} class={link} aria-label="사용자 관리">
					<Users /><span class="hidden sm:inline">사용자 관리</span>
				</a>
			{/if}
			<a href={resolve('/me')} class={link} aria-label="내 페이지">
				<User /><span class="hidden sm:inline">내 페이지</span>
			</a>
			<button class={link} onclick={signOut} aria-label="로그아웃">
				<LogOut /><span class="hidden sm:inline">로그아웃</span>
			</button>
		</nav>
	</div>
</header>
