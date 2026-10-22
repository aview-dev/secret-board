<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import ShieldCheck from '@lucide/svelte/icons/shield-check';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	import { MAX_BODY_LEN, MAX_TITLE_LEN } from '$lib/board';
	import { BUTTON_OUTLINE, BUTTON_PRIMARY, INPUT, PANEL } from '$lib/ui';

	let { form } = $props();
	let submitting = $state(false);

	// 글자 수 표시용. 실패해서 돌아오면 서버가 돌려준 값부터 시작한다.
	// svelte-ignore state_referenced_locally
	let title = $state(form?.title ?? '');
	// svelte-ignore state_referenced_locally
	let body = $state(form?.body ?? '');
</script>

<main class="mx-auto max-w-3xl space-y-4 px-4 py-8">
	<Breadcrumb item="글쓰기" />

	<form
		method="POST"
		use:enhance={() => {
			submitting = true;
			return async ({ update }) => {
				await update({ reset: false });
				submitting = false;
			};
		}}
		class={PANEL}
	>
		<div class="space-y-4 px-6 py-5">
			{#if form?.message}
				<p
					class="rounded-md border border-error-border bg-error-bg px-3 py-2 text-sm text-error-fg"
				>
					{form.message}
				</p>
			{/if}

			<div>
				<div class="mb-1.5 flex items-baseline justify-between">
					<label for="title" class="text-sm font-medium">제목</label>
					<span class="text-2xs text-muted-foreground tabular-nums">
						{title.length} / {MAX_TITLE_LEN}
					</span>
				</div>
				<input
					id="title"
					name="title"
					placeholder="제목을 적어 주세요"
					required
					maxlength={MAX_TITLE_LEN}
					bind:value={title}
					class={INPUT}
				/>
			</div>

			<div>
				<div class="mb-1.5 flex items-baseline justify-between">
					<label for="body" class="text-sm font-medium">내용</label>
					<span class="text-2xs text-muted-foreground tabular-nums">
						{body.length.toLocaleString()} / {MAX_BODY_LEN.toLocaleString()}
					</span>
				</div>
				<textarea
					id="body"
					name="body"
					placeholder="내용을 적어 주세요. 마크다운 (**굵게**, `코드`, - 목록, [링크](https://...)) 을 쓸 수 있습니다. 이미지는 보이지 않습니다."
					required
					maxlength={MAX_BODY_LEN}
					rows="12"
					bind:value={body}
					class="{INPUT} resize-y leading-relaxed"></textarea>
			</div>
		</div>

		<footer
			class="flex flex-wrap items-center justify-between gap-3 border-t bg-panel-header px-6 py-3"
		>
			<p class="flex items-center gap-1.5 text-xs text-muted-foreground">
				<ShieldCheck class="size-4 shrink-0" />
				작성자 정보는 저장되지 않습니다. 작성 시각은 시간 단위로만 남습니다.
			</p>
			<div class="flex shrink-0 items-center gap-2">
				<a href={resolve('/')} class={BUTTON_OUTLINE}>취소</a>
				<button disabled={submitting} class={BUTTON_PRIMARY}>올리기</button>
			</div>
		</footer>
	</form>
</main>
