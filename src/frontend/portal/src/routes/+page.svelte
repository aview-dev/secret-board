<script lang="ts">
	import { resolve } from '$app/paths';
	import List from '@lucide/svelte/icons/list';
	import MessageSquare from '@lucide/svelte/icons/message-square';
	import PenLine from '@lucide/svelte/icons/pen-line';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PanelHeader from '$lib/components/PanelHeader.svelte';
	import { formatHour } from '$lib/board';
	import { BUTTON_PRIMARY, PANEL } from '$lib/ui';

	let { data } = $props();
</script>

<main class="mx-auto max-w-3xl space-y-4 px-4 py-8">
	<PageHeader title="게시판" description="작성자가 저장되지 않는 익명 게시판입니다.">
		{#snippet actions()}
			<a href={resolve('/posts/new')} class={BUTTON_PRIMARY}><PenLine />글쓰기</a>
		{/snippet}
	</PageHeader>

	<section class={PANEL}>
		<PanelHeader icon={List} title="전체 글 {data.posts.length}" />
		{#if data.posts.length === 0}
			<div
				class="m-4 rounded-md border border-dashed p-8 text-center text-sm text-muted-foreground"
			>
				<p class="font-medium text-foreground">아직 글이 없습니다</p>
				<p class="mt-1">첫 글을 남겨 보세요.</p>
			</div>
		{:else}
			<ul class="divide-y">
				{#each data.posts as p (p.id)}
					<li>
						<!-- aviewops 목록 행 관용구: hover 하면 왼쪽 3px 막대와 옅은 배경. 막대 폭은 늘 잡아 둬서 행이 흔들리지 않는다. -->
						<a
							href={resolve('/posts/[id]', { id: String(p.id) })}
							class="flex items-center justify-between gap-4 border-l-[3px] border-l-transparent px-4 py-3 hover:border-l-row-accent hover:bg-row-hover"
						>
							<span class="min-w-0 truncate text-sm font-medium">{p.title}</span>
							<span
								class="flex shrink-0 items-center gap-3 text-xs text-muted-foreground tabular-nums"
							>
								{#if p.commentCount > 0}
									<span class="flex items-center gap-1" title="댓글 {p.commentCount}개">
										<MessageSquare class="size-3.5" />{p.commentCount}
									</span>
								{/if}
								<span>{formatHour(p.createdAt)}</span>
							</span>
						</a>
					</li>
				{/each}
			</ul>
		{/if}
	</section>
</main>
