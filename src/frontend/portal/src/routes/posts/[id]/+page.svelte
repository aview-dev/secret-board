<script lang="ts">
	import { enhance } from '$app/forms';
	import Clock from '@lucide/svelte/icons/clock';
	import MessageSquare from '@lucide/svelte/icons/message-square';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	import PanelHeader from '$lib/components/PanelHeader.svelte';
	import { MAX_COMMENT_LEN, formatHour } from '$lib/board';
	import { BUTTON_PRIMARY, INPUT, PANEL } from '$lib/ui';

	let { data, form } = $props();
	let submitting = $state(false);
</script>

<main class="mx-auto max-w-3xl space-y-4 px-4 py-8">
	<Breadcrumb item={data.post.title} />

	<article class={PANEL}>
		<header class="border-b px-6 py-5">
			<h1 class="text-xl font-semibold tracking-tight break-words">{data.post.title}</h1>
			<p class="mt-2 flex items-center gap-1.5 text-xs text-muted-foreground">
				<Clock class="size-3.5" />{formatHour(data.post.createdAt)}
			</p>
		</header>
		<!-- whitespace-pre-wrap: 줄바꿈을 그대로 보여준다. {@html} 을 쓰지 않으므로 HTML 은 글자로 나온다. -->
		<div class="px-6 py-5 text-base leading-relaxed break-words whitespace-pre-wrap">
			{data.post.body}
		</div>
	</article>

	<section class={PANEL}>
		<PanelHeader icon={MessageSquare} title="댓글 {data.comments.length}" />

		{#if data.comments.length > 0}
			<ul class="divide-y">
				{#each data.comments as c, i (c.id)}
					<li class="px-4 py-3">
						<p class="flex items-center gap-2 text-xs text-muted-foreground">
							<span class="rounded-sm border bg-muted px-1.5 py-0.5 font-medium text-foreground">
								익명 {i + 1}
							</span>
							<span>{formatHour(c.createdAt)}</span>
						</p>
						<p class="mt-2 text-sm break-words whitespace-pre-wrap">{c.body}</p>
					</li>
				{/each}
			</ul>
		{:else}
			<p class="px-4 py-6 text-center text-sm text-muted-foreground">첫 댓글을 남겨 보세요.</p>
		{/if}

		<form
			method="POST"
			action="?/comment"
			use:enhance={() => {
				submitting = true;
				return async ({ update }) => {
					// 성공하면 입력창을 비우고, 실패하면 적던 내용을 남긴다.
					await update();
					submitting = false;
				};
			}}
			class="border-t bg-panel-header px-4 py-3"
		>
			{#if form?.message}
				<p
					class="mb-2 rounded-md border border-error-border bg-error-bg px-3 py-2 text-xs text-error-fg"
				>
					{form.message}
				</p>
			{/if}
			<textarea
				name="body"
				placeholder="댓글을 적어 주세요"
				aria-label="댓글"
				required
				maxlength={MAX_COMMENT_LEN}
				rows="3"
				class="{INPUT} resize-y">{form?.body ?? ''}</textarea
			>
			<div class="mt-2 flex justify-end">
				<button disabled={submitting} class={BUTTON_PRIMARY}>댓글 달기</button>
			</div>
		</form>
	</section>
</main>
