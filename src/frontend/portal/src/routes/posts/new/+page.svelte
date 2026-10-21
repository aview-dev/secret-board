<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import { MAX_BODY_LEN, MAX_TITLE_LEN } from '$lib/board';

	let { form } = $props();
	let submitting = $state(false);
</script>

<main class="mx-auto max-w-2xl space-y-6 px-4 py-10">
	<h1 class="text-2xl font-semibold tracking-tight">글쓰기</h1>

	{#if form?.message}
		<p class="rounded-md bg-neutral-100 px-3 py-2 text-sm text-neutral-700">{form.message}</p>
	{/if}

	<form
		method="POST"
		use:enhance={() => {
			submitting = true;
			return async ({ update }) => {
				await update({ reset: false });
				submitting = false;
			};
		}}
		class="overflow-hidden rounded-xl border border-neutral-200 bg-white"
	>
		<input
			name="title"
			placeholder="제목"
			required
			maxlength={MAX_TITLE_LEN}
			value={form?.title ?? ''}
			class="w-full border-b border-neutral-200 px-6 py-4 text-base font-medium outline-none placeholder:text-neutral-400"
		/>
		<textarea
			name="body"
			placeholder="내용을 적어 주세요"
			required
			maxlength={MAX_BODY_LEN}
			rows="12"
			class="block w-full resize-y px-6 py-4 text-sm outline-none placeholder:text-neutral-400"
			>{form?.body ?? ''}</textarea
		>
		<footer
			class="flex items-center justify-between border-t border-neutral-200 bg-neutral-50 px-6 py-3"
		>
			<p class="text-xs text-neutral-500">
				작성자 정보는 저장되지 않습니다. 작성 시각은 시간 단위로만 남습니다.
			</p>
			<div class="flex shrink-0 items-center gap-3">
				<a href={resolve('/')} class="text-sm text-neutral-500 hover:text-neutral-900">취소</a>
				<button
					disabled={submitting}
					class="rounded-md bg-neutral-900 px-3 py-1.5 text-sm font-medium text-white hover:bg-neutral-700 disabled:opacity-50"
				>
					올리기
				</button>
			</div>
		</footer>
	</form>
</main>
