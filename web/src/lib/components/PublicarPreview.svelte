<script>
	/**
	 * Preview inline da publicação — mostra como ficará no Telegram/WhatsApp.
	 * Sempre exibe foto (se disponível) + legenda + botão fake de compra.
	 */
	let { imagem = '', legenda = '', link = '' } = $props();
</script>

<div class="mb-5">
	<span class="text-xs font-semibold uppercase tracking-wide text-tinta-suave">Preview</span>
	<div
		class="mt-1.5 overflow-hidden rounded-md border border-border bg-[var(--branco)] shadow-[0_2px_8px_rgba(0,0,0,0.04)]"
	>
		{#if imagem}
			<img src={imagem} alt="preview" class="block max-h-[280px] w-full object-cover" />
		{/if}
		<!-- eslint-disable-next-line svelte/no-at-html-tags -- legenda is generated internally by templates, not user-provided raw HTML -->
		<div class="p-4 text-[0.92rem] leading-relaxed">{@html legenda.replace(/\n/g, '<br>')}</div>
		{#if link}
			<div class="px-4 pb-4">
				<span class="inline-block rounded-lg bg-[var(--ouro)] px-4 py-2 text-sm font-semibold text-white"
					>🛒 Comprar</span
				>
			</div>
		{/if}
	</div>
</div>
