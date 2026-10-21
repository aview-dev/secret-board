<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import { MAX_COMMENT_LEN, formatHour } from '$lib/board';

	let { data, form } = $props();
	let submitting = $state(false);
</script>

<main class="mx-auto max-w-2xl space-y-6 px-4 py-10">
	<a href={resolve('/')} class="text-sm text-neutral-500 hover:text-neutral-900">&larr; 목록</a>

	<article class="overflow-hidden rounded-xl border border-neutral-200 bg-white">
		<header class="border-b border-neutral-200 bg-neutral-50 px-6 py-4">
			<h1 class="text-lg font-semibold break-words text-neutral-900">{data.post.title}</h1>
			<p class="mt-0.5 text-xs text-neutral-500">{formatHour(data.post.createdAt)}</p>
		</header>
		<!-- whitespace-pre-wrap: 줄바꿈을 그대로 보여준다. {@html} 을 쓰지 않으므로 HTML 은 글자로 나온다. -->
		<div class="px-6 py-5 text-sm leading-relaxed break-words whitespace-pre-wrap text-neutral-900">
			{data.post.body}
		</div>
	</article>

	<section class="overflow-hidden rounded-xl border border-neutral-200 bg-white">
		<header class="border-b border-neutral-200 bg-neutral-50 px-6 py-3">
			<h2 class="text-sm font-semibold text-neutral-900">댓글 {data.comments.length}</h2>
		</header>

		{#if data.comments.length > 0}
			<ul class="divide-y divide-neutral-100">
				{#each data.comments as c, i (c.id)}
					<li class="px-6 py-3">
						<p class="text-xs text-neutral-400">익명 {i + 1} · {formatHour(c.createdAt)}</p>
						<p class="mt-1 text-sm break-words whitespace-pre-wrap text-neutral-900">{c.body}</p>
					</li>
				{/each}
			</ul>
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
			class="border-t border-neutral-200 px-6 py-4"
		>
			{#if form?.message}
				<p class="mb-2 text-xs text-red-600">{form.message}</p>
			{/if}
			<textarea
				name="body"
				placeholder="댓글을 적어 주세요"
				required
				maxlength={MAX_COMMENT_LEN}
				rows="3"
				class="block w-full resize-y rounded-md border border-neutral-300 px-3 py-2 text-sm outline-none placeholder:text-neutral-400 focus:border-neutral-500"
				>{form?.body ?? ''}</textarea
			>
			<div class="mt-2 flex justify-end">
				<button
					disabled={submitting}
					class="rounded-md bg-neutral-900 px-3 py-1.5 text-sm font-medium text-white hover:bg-neutral-700 disabled:opacity-50"
				>
					댓글 달기
				</button>
			</div>
		</form>
	</section>
</main>
