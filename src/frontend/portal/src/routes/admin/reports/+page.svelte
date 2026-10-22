<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import Flag from '@lucide/svelte/icons/flag';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	import PanelHeader from '$lib/components/PanelHeader.svelte';
	import { REPORT_REASONS, REPORT_REASON_LABEL, formatHour } from '$lib/board';
	import { BUTTON_GHOST, BUTTON_OUTLINE, PANEL } from '$lib/ui';

	let { data, form } = $props();
	// 처리 중인 글 id. 그 줄의 버튼만 잠근다.
	let busy = $state<number | null>(null);

	// 두 버튼 모두 되돌릴 수 없으니 한 번 더 묻는다.
	function confirmed(action: string | null) {
		return action?.endsWith('delete')
			? confirm(
					'이 글을 삭제할까요? 댓글, 좋아요 / 싫어요, 신고도 함께 지워지고 되돌릴 수 없습니다.'
				)
			: confirm('신고를 초기화할까요? 글은 남고, 이후 다시 신고될 수 있습니다.');
	}
</script>

<main class="mx-auto max-w-3xl space-y-4 px-4 py-8">
	<Breadcrumb item="신고 목록" />

	{#if form?.message}
		<p class="rounded-md border bg-panel px-3 py-2 text-sm">{form.message}</p>
	{/if}

	<section class={PANEL}>
		<PanelHeader
			icon={Flag}
			title="신고된 글 {data.posts.length}"
			subtitle="신고가 많은 순. 누가 신고했는지는 저장되지 않아 관리자도 알 수 없습니다."
		/>

		{#if data.posts.length > 0}
			<ul class="divide-y">
				{#each data.posts as p (p.id)}
					<li class="flex flex-wrap items-center gap-3 px-4 py-3">
						<div class="min-w-0 flex-1">
							<a
								href={resolve('/posts/[id]', { id: String(p.id) })}
								class="block truncate font-medium hover:underline"
							>
								{p.title}
							</a>
							<p class="mt-1 flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground">
								<span class="font-medium text-error-fg tabular-nums">신고 {p.reports}</span>
								{#each REPORT_REASONS as r (r)}
									{#if p.reasons[r]}
										<span class="rounded-sm border bg-muted px-1.5 py-0.5">
											{REPORT_REASON_LABEL[r]}
											{p.reasons[r]}
										</span>
									{/if}
								{/each}
								<span>· {formatHour(p.createdAt)}</span>
							</p>
						</div>
						<form
							method="POST"
							use:enhance={({ action, cancel }) => {
								if (!confirmed(action.search)) return cancel();
								busy = p.id;
								return async ({ update }) => {
									await update();
									busy = null;
								};
							}}
							class="flex shrink-0 items-center gap-1"
						>
							<input type="hidden" name="id" value={p.id} />
							<button formaction="?/clear" disabled={busy === p.id} class={BUTTON_OUTLINE}>
								문제없음
							</button>
							<button
								formaction="?/delete"
								disabled={busy === p.id}
								class="{BUTTON_GHOST} hover:text-error-fg"
							>
								삭제
							</button>
						</form>
					</li>
				{/each}
			</ul>
		{:else}
			<p class="px-4 py-6 text-center text-sm text-muted-foreground">신고된 글이 없습니다.</p>
		{/if}
	</section>
</main>
