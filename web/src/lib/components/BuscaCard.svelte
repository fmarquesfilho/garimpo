<script>
	/**
	 * BuscaCard — cartão de uma busca salva. Dividido em seções (palavras-chave,
	 * categorias, lojas, marketplaces) exibidas só quando presentes, mais a
	 * informação de agendamento. Ações: rodar, editar (edit mode) e remover.
	 *
	 * Consome o formato de config (payloadToConfig): keywords[], categorias[],
	 * shopIds[], shopNomes{}, marketplaces (string|string[]), cron, id.
	 *
	 * @prop busca — config da busca salva
	 * @prop editando — true quando este card está em edit mode
	 * @prop onrodar, oneditar, onremover — (busca) => void
	 */
	import { cronLabel } from '$lib/busca-engine.svelte.js';
	import { cn } from '$lib/utils';

	let { busca, editando = false, selecionada = false, onrodar = null, oneditar = null, onremover = null } = $props();

	let keywords = $derived(busca.keywords ?? []);
	let categorias = $derived(busca.categorias ?? []);
	let lojas = $derived((busca.shopIds ?? []).map((id) => busca.shopNomes?.[id] || id));
	let marketplaces = $derived(
		Array.isArray(busca.marketplaces) ? busca.marketplaces : busca.marketplaces ? [busca.marketplaces] : []
	);
</script>

<div
	class={cn(
		'relative flex min-w-[250px] max-w-[340px] flex-1 flex-col rounded-sm border border-border bg-card px-3 py-2',
		editando && '!border-primary ring-2 ring-ring/20',
		selecionada && !editando && '!border-primary/60 bg-[var(--ouro-fundo)]'
	)}
>
	<div class="absolute right-1.5 top-1.5 flex items-center gap-0.5">
		<button
			type="button"
			class="rounded px-1 text-xs text-muted-foreground hover:bg-accent hover:text-primary"
			onclick={() => oneditar?.(busca)}
			aria-label="Editar busca"
			title="Editar">✎</button
		>
		<button
			type="button"
			class="rounded px-1 text-sm leading-none text-muted-foreground hover:text-destructive"
			onclick={() => onremover?.(busca)}
			aria-label="Remover busca"
			title="Remover">✕</button
		>
	</div>

	<div class="flex flex-col gap-1 pr-11 text-sm">
		{#if keywords.length}
			<div class="flex items-center gap-1.5">
				<span class="flex flex-wrap gap-1">
					{#each keywords as k (k)}<span
							class="rounded-full border border-[var(--ouro-claro)] bg-[var(--ouro-fundo)] px-2 py-0.5 text-xs font-semibold text-[var(--ouro-escuro)]"
							>{k}</span
						>{/each}
				</span>
				{#if busca.cron}
					<span
						class="inline-flex items-center gap-1 rounded-full border border-[var(--aviso-borda)] bg-[var(--aviso-fundo)] px-2 py-px font-[var(--mono)] text-[0.68rem] text-[var(--aviso-texto)]"
						>⏱ {cronLabel(busca.cron)}</span
					>
				{/if}
			</div>
		{:else if !categorias.length && !lojas.length}
			<div class="flex items-center gap-1.5">
				<span class="text-xs text-muted-foreground">(sem keywords)</span>
				{#if busca.cron}
					<span
						class="inline-flex items-center gap-1 rounded-full border border-[var(--aviso-borda)] bg-[var(--aviso-fundo)] px-2 py-px font-[var(--mono)] text-[0.68rem] text-[var(--aviso-texto)]"
						>⏱ {cronLabel(busca.cron)}</span
					>
				{/if}
			</div>
		{/if}
		{#if categorias.length}
			<div class="flex items-center gap-1.5">
				<span class="flex flex-wrap gap-1">
					{#each categorias as c (c)}<span class="rounded-full border border-border bg-muted px-2 py-0.5 text-xs"
							>{c}</span
						>{/each}
				</span>
				{#if !keywords.length && busca.cron}
					<span
						class="inline-flex items-center gap-1 rounded-full border border-[var(--aviso-borda)] bg-[var(--aviso-fundo)] px-2 py-px font-[var(--mono)] text-[0.68rem] text-[var(--aviso-texto)]"
						>⏱ {cronLabel(busca.cron)}</span
					>
				{/if}
			</div>
		{/if}
		{#if lojas.length}
			<div class="flex items-center gap-1.5">
				<span class="flex flex-wrap gap-1">
					{#each lojas as l (l)}<span class="rounded-full border border-border bg-muted px-2 py-0.5 text-xs"
							>🏪 {l}</span
						>{/each}
				</span>
			</div>
		{/if}
		{#if marketplaces.length}
			<div class="flex items-center gap-1.5">
				<span class="flex flex-wrap gap-1">
					{#each marketplaces as m (m)}<span class="rounded-full border border-border bg-muted px-2 py-0.5 text-xs"
							>{m}</span
						>{/each}
				</span>
			</div>
		{/if}
	</div>

	<div class="mt-1.5 flex items-center justify-between border-t border-border pt-1.5 text-xs text-muted-foreground">
		{#if editando}
			<span class="font-semibold text-[var(--ouro-escuro)]">✎ editando — altere e salve</span>
		{:else}
			<span>{busca.cron ? 'coleta periódica' : 'busca manual salva'}</span>
			<button
				type="button"
				class="rounded border border-primary bg-[var(--ouro-fundo)] px-3 py-1 text-xs font-semibold text-[var(--ouro-escuro)] hover:bg-primary hover:text-primary-foreground"
				onclick={() => onrodar?.(busca)}>↻ rodar</button
			>
		{/if}
	</div>
</div>
