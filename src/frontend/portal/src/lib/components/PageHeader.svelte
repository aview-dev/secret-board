<!-- 페이지 제목 줄. aviewops LayoutPageHeader 축소판: 제목, 보조 문구, 오른쪽 actions. -->
<script lang="ts">
	import type { Snippet } from 'svelte';

	let { title, description, actions }: { title: string; description?: string; actions?: Snippet } =
		$props();
</script>

<div class="flex items-end justify-between gap-4">
	<div class="min-w-0">
		<h1 class="truncate text-xl font-semibold tracking-tight">{title}</h1>
		{#if description}
			<p class="mt-1 text-sm text-muted-foreground">{description}</p>
		{/if}
	</div>
	{#if actions}
		<div class="flex shrink-0 items-center gap-2">{@render actions()}</div>
	{/if}
</div>
