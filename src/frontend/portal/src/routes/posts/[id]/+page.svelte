<script lang="ts">
	import { enhance } from '$app/forms';
	import Clock from '@lucide/svelte/icons/clock';
	import MessageSquare from '@lucide/svelte/icons/message-square';
	import ThumbsDown from '@lucide/svelte/icons/thumbs-down';
	import ThumbsUp from '@lucide/svelte/icons/thumbs-up';
	import Trash2 from '@lucide/svelte/icons/trash-2';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	import PanelHeader from '$lib/components/PanelHeader.svelte';
	import { MAX_COMMENT_LEN, formatHour } from '$lib/board';
	import { BUTTON_GHOST, BUTTON_PRIMARY, INPUT, PANEL } from '$lib/ui';

	let { data, form } = $props();
	let submitting = $state(false);
	let voting = $state(false);
	let deleting = $state(false);

	// 고른 표는 색으로 구분한다. 좋아요는 강조색, 싫어요는 붉은 계열.
	const VOTE_BUTTON =
		'inline-flex h-8 items-center gap-1.5 rounded-md border px-3 text-sm font-medium tabular-nums transition-all disabled:opacity-50 [&_svg]:size-4';
	const VOTE_IDLE = 'bg-control text-muted-foreground hover:text-foreground hover:brightness-95';
	const LIKE_ON = 'border-primary bg-primary/10 text-primary';
	const DISLIKE_ON = 'border-error-border bg-error-bg text-error-fg';
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
		<!-- bodyHtml 은 lib/server/markdown.ts 가 만든 것만 온다 (원문 HTML 은 글자로 바뀐 상태). -->
		<div class="markdown px-6 py-5 text-base leading-relaxed">
			<!-- eslint-disable-next-line svelte/no-at-html-tags -->
			{@html data.bodyHtml}
		</div>
		<form
			method="POST"
			action="?/vote"
			use:enhance={() => {
				voting = true;
				return async ({ update }) => {
					await update();
					voting = false;
				};
			}}
			class="flex flex-wrap items-center gap-2 border-t px-6 py-3"
		>
			<!-- 이미 누른 버튼을 다시 누르면 0 (취소) 을 보낸다. -->
			<button
				name="value"
				value={data.myVote === 1 ? 0 : 1}
				disabled={voting}
				aria-pressed={data.myVote === 1}
				title={data.myVote === 1 ? '좋아요 취소' : '좋아요'}
				class="{VOTE_BUTTON} {data.myVote === 1 ? LIKE_ON : VOTE_IDLE}"
			>
				<ThumbsUp />{data.post.likes}
			</button>
			<button
				name="value"
				value={data.myVote === -1 ? 0 : -1}
				disabled={voting}
				aria-pressed={data.myVote === -1}
				title={data.myVote === -1 ? '싫어요 취소' : '싫어요'}
				class="{VOTE_BUTTON} {data.myVote === -1 ? DISLIKE_ON : VOTE_IDLE}"
			>
				<ThumbsDown />{data.post.dislikes}
			</button>
			{#if form?.voteMessage}
				<span class="text-xs text-error-fg">{form.voteMessage}</span>
			{:else}
				<span class="text-xs text-muted-foreground">누가 눌렀는지는 저장되지 않습니다.</span>
			{/if}
		</form>
		<!-- 관리자에게만 보인다. 버튼을 숨기는 건 화면 편의이고, 실제 권한 확인은 delete action 이 한다. -->
		{#if data.me?.role === 'admin'}
			<form
				method="POST"
				action="?/delete"
				use:enhance={({ cancel }) => {
					if (
						!confirm(
							'이 글을 삭제할까요? 댓글과 좋아요 / 싫어요도 함께 지워지고 되돌릴 수 없습니다.'
						)
					)
						return cancel();
					deleting = true;
					return async ({ update }) => {
						await update();
						deleting = false;
					};
				}}
				class="flex items-center justify-end gap-2 border-t px-6 py-2"
			>
				{#if form?.deleteMessage}
					<span class="text-xs text-error-fg">{form.deleteMessage}</span>
				{/if}
				<button disabled={deleting} class="{BUTTON_GHOST} hover:text-error-fg">
					<Trash2 />관리자 삭제
				</button>
			</form>
		{/if}
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
						<div class="markdown mt-2 text-sm">
							<!-- eslint-disable-next-line svelte/no-at-html-tags -->
							{@html c.bodyHtml}
						</div>
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
				placeholder="댓글을 적어 주세요 (마크다운 가능)"
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
