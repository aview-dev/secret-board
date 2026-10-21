<script lang="ts">
	import { enhance } from '$app/forms';
	import { ROLES, ROLE_LABEL } from '$lib/roles';

	let { data, form } = $props();

	const fmt = new Intl.DateTimeFormat('ko-KR', { dateStyle: 'medium' });
	const adminCount = $derived(data.users.filter((u) => u.role === 'admin').length);
</script>

<main class="mx-auto max-w-2xl space-y-6 px-4 py-10">
	<h1 class="text-2xl font-semibold tracking-tight">사용자 관리</h1>

	{#if form?.message}
		<p class="rounded-md bg-neutral-100 px-3 py-2 text-sm text-neutral-700">{form.message}</p>
	{/if}

	<section class="overflow-hidden rounded-xl border border-neutral-200 bg-white">
		<header class="border-b border-neutral-200 bg-neutral-50 px-6 py-4">
			<h2 class="text-base font-semibold text-neutral-900">사용자 목록</h2>
			<p class="mt-0.5 text-xs text-neutral-500">
				전체 {data.users.length}명 · 관리자 {adminCount}명. 로그인 기록은 익명성을 위해 보여주지
				않습니다.
			</p>
		</header>
		<ul class="divide-y divide-neutral-100 text-sm">
			{#each data.users as u (u.id)}
				<li class="flex items-center justify-between gap-4 px-6 py-3">
					<div class="min-w-0">
						<p class="truncate font-medium text-neutral-900">
							{u.email}
							{#if u.id === data.myId}<span class="ml-1 text-xs text-neutral-400">(나)</span>{/if}
						</p>
						<p class="text-xs text-neutral-400">{fmt.format(new Date(u.createdAt))} 가입</p>
					</div>
					{#if u.fixed || u.id === data.myId}
						<span
							class="shrink-0 rounded bg-neutral-900 px-2 py-0.5 text-xs font-medium text-white"
							title={u.fixed ? 'ADMIN_EMAILS 로 고정된 관리자' : ''}
						>
							{ROLE_LABEL[u.role]}{u.fixed ? ' (고정)' : ''}
						</span>
					{:else}
						<form method="POST" action="?/setRole" use:enhance class="shrink-0">
							<input type="hidden" name="id" value={u.id} />
							<select
								name="role"
								value={u.role}
								class="rounded-md border border-neutral-300 bg-white px-2 py-1 text-xs"
								onchange={(e) => e.currentTarget.form?.requestSubmit()}
							>
								{#each ROLES as r (r)}
									<option value={r}>{ROLE_LABEL[r]}</option>
								{/each}
							</select>
						</form>
					{/if}
				</li>
			{/each}
		</ul>
	</section>
</main>
