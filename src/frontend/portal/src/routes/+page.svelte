<script lang="ts">
	import { resolve } from '$app/paths';
	import { formatHour } from '$lib/board';

	let { data } = $props();
</script>

<main class="mx-auto max-w-2xl space-y-6 px-4 py-10">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-semibold tracking-tight">게시판</h1>
		<a
			href={resolve('/posts/new')}
			class="rounded-md bg-neutral-900 px-3 py-1.5 text-sm font-medium text-white hover:bg-neutral-700"
		>
			글쓰기
		</a>
	</div>

	<section class="overflow-hidden rounded-xl border border-neutral-200 bg-white">
		{#if data.posts.length === 0}
			<p class="px-6 py-10 text-center text-sm text-neutral-500">
				아직 글이 없습니다. 첫 글을 남겨 보세요.
			</p>
		{:else}
			<ul class="divide-y divide-neutral-100">
				{#each data.posts as p (p.id)}
					<li>
						<a
							href={resolve('/posts/[id]', { id: String(p.id) })}
							class="flex items-center justify-between gap-4 px-6 py-3 hover:bg-neutral-50"
						>
							<span class="min-w-0 truncate text-sm font-medium text-neutral-900">
								{p.title}
								{#if p.commentCount > 0}
									<span class="ml-1 text-xs text-neutral-400">[{p.commentCount}]</span>
								{/if}
							</span>
							<span class="shrink-0 text-xs text-neutral-400">{formatHour(p.createdAt)}</span>
						</a>
					</li>
				{/each}
			</ul>
		{/if}
	</section>
</main>
