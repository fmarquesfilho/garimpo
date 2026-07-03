<script>
	/** Cartão de uma busca salva. Exibe keywords, badges e ações. */
	import { Badge, Button } from '$lib/components/ui';

	let { busca, buscaAtiva = '', onaplicar = null, onproximakw = null, onremover = null } = $props();
</script>

<div class="flex flex-col gap-2 rounded-md border border-border bg-card px-4 py-3">
	<div class="flex items-start justify-between gap-3">
		<div class="flex flex-1 flex-wrap gap-2">
			{#each busca.keywords ?? [] as kw, i}
				<button
					type="button"
					class="cursor-pointer rounded-full border border-border bg-porcelana px-3 py-1 font-[var(--ui)] font-semibold text-foreground transition-[background,border-color] duration-150 ease-linear hover:border-ouro hover:bg-ouro-fundo hover:text-ouro-escuro motion-reduce:transition-none"
					class:!bg-ouro-fundo={buscaAtiva === kw}
					class:!border-ouro={buscaAtiva === kw}
					class:!text-ouro-escuro={buscaAtiva === kw}
					onclick={() =>
						onaplicar?.({ ...busca, keywords: busca.keywords.slice(i).concat(busca.keywords.slice(0, i)) })}
					title="Aplicar filtros com '{kw}'">{kw}</button
				>
			{/each}
		</div>
		<div class="flex shrink-0 items-center gap-1">
			{#if (busca.keywords?.length ?? 0) > 1}
				<Button variant="ghost" size="sm" onclick={() => onproximakw?.(busca)} aria-label="Próxima keyword">→</Button>
			{/if}
			<Button variant="ghost" size="sm" onclick={() => onremover?.(busca.id)} aria-label="Remover busca">✕</Button>
		</div>
	</div>
	<div class="flex flex-wrap gap-2">
		{#if busca.fontes?.length}
			{#each busca.fontes as f}
				<Badge variant="gold"
					>{f === 'curadoria' ? '🔍' : f === 'quedas' ? '📉' : f === 'novos' ? '🆕' : '⭐'} {f}</Badge
				>
			{/each}
		{:else}
			<Badge>{busca.estrategia ?? 'nicho'}</Badge>
		{/if}
		{#if busca.cron}
			<Badge variant="gold">⏱ agendada</Badge>
		{/if}
		{#if busca.categorias?.length}
			{#each busca.categorias as cat}
				<Badge variant="pink">{cat}</Badge>
			{/each}
		{:else if busca.categoria}
			<Badge variant="pink">{busca.categoria}</Badge>
		{/if}
		{#if busca.shop_ids?.length}
			<Badge variant="pink">🏪 {busca.shop_ids.length} {busca.shop_ids.length === 1 ? 'loja' : 'lojas'}</Badge>
		{/if}
		{#if busca.dias_janela && busca.dias_janela !== 7}
			<Badge>janela: {busca.dias_janela}d</Badge>
		{/if}
	</div>
</div>
