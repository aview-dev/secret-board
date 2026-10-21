<script lang="ts">
	let { data } = $props();

	const fmt = new Intl.DateTimeFormat('ko-KR', { dateStyle: 'long' });
</script>

<main class="mx-auto max-w-2xl space-y-6 px-4 py-10">
	<h1 class="text-2xl font-semibold tracking-tight">내 페이지</h1>

	<section class="overflow-hidden rounded-xl border border-neutral-200 bg-white">
		<header class="border-b border-neutral-200 bg-neutral-50 px-6 py-4">
			<h2 class="text-base font-semibold text-neutral-900">계정 정보</h2>
			<p class="mt-0.5 text-xs text-neutral-500">
				로그인 확인에만 쓰이고, 게시글과는 연결되지 않습니다.
			</p>
		</header>
		<dl class="divide-y divide-neutral-100 text-sm">
			<div class="flex items-center px-6 py-3">
				<dt class="w-24 shrink-0 text-neutral-500">이메일</dt>
				<dd class="font-medium text-neutral-900">{data.me?.email}</dd>
			</div>
			<div class="flex items-center px-6 py-3">
				<dt class="w-24 shrink-0 text-neutral-500">권한</dt>
				<dd>
					{#if data.me?.role === 'admin'}
						<span class="rounded bg-neutral-900 px-2 py-0.5 text-xs font-medium text-white">
							관리자
						</span>
					{:else}
						<span class="rounded bg-neutral-100 px-2 py-0.5 text-xs font-medium text-neutral-700">
							일반
						</span>
					{/if}
				</dd>
			</div>
			<div class="flex items-center px-6 py-3">
				<dt class="w-24 shrink-0 text-neutral-500">가입일</dt>
				<dd class="text-neutral-900">{fmt.format(new Date(data.joinedAt))}</dd>
			</div>
		</dl>
	</section>
</main>
