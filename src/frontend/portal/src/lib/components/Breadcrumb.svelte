<!-- 상세 화면 위 경로 표시. aviewops DetailPanelHeader 의 "목록 / 항목" 모양. 목록은 게시판 하나뿐이라 고정한다. -->
<script lang="ts">
	import { resolve } from '$app/paths';

	let { item }: { item: string } = $props();
</script>

<nav class="flex min-w-0 items-center gap-2 text-sm text-muted-foreground">
	<a href={resolve('/')} class="shrink-0 hover:text-foreground hover:underline">게시판</a>
	<span class="shrink-0">/</span>
	<span class="truncate font-medium text-foreground">{item}</span>
</nav>
