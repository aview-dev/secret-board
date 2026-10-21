<!-- 패널 머리 줄. aviewops CardPanelHeader 축소판: 아이콘 + 제목 + 보조 문구, 오른쪽 actions. -->
<script lang="ts">
	import type { Component, Snippet } from 'svelte';

	let {
		icon: Icon,
		title,
		subtitle,
		actions
	}: { icon: Component; title: string; subtitle?: string; actions?: Snippet } = $props();
</script>

<div class="flex items-center justify-between gap-2 border-b bg-panel-header px-4 py-2.5">
	<div class="min-w-0">
		<div class="flex items-center gap-2 text-sm font-semibold">
			<Icon class="size-4 shrink-0 text-muted-foreground" />
			<span class="truncate">{title}</span>
		</div>
		{#if subtitle}
			<p class="mt-0.5 text-xs text-muted-foreground">{subtitle}</p>
		{/if}
	</div>
	{#if actions}
		<div class="flex shrink-0 items-center gap-2">{@render actions()}</div>
	{/if}
</div>
