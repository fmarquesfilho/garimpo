<script>
	/** Cartão de uma busca salva. Exibe keywords, badges e ações. */
	import { Badge, Button } from '$lib/components/ui';

	let { busca, buscaAtiva = '', onaplicar = null, onproximakw = null, onremover = null } = $props();
</script>

<div class="cartao-busca">
	<div class="cartao-topo">
		<div class="cartao-kws">
			{#each busca.keywords ?? [] as kw, i}
				<button
					type="button"
					class="kw-btn"
					class:kw-ativa={buscaAtiva === kw}
					onclick={() => onaplicar?.({ ...busca, keywords: busca.keywords.slice(i).concat(busca.keywords.slice(0, i)) })}
					title="Aplicar filtros com '{kw}'"
				>{kw}</button>
			{/each}
		</div>
		<div class="cartao-acoes">
			{#if (busca.keywords?.length ?? 0) > 1}
				<Button variant="ghost" size="sm" onclick={() => onproximakw?.(busca)} aria-label="Próxima keyword">→</Button>
			{/if}
			<Button variant="ghost" size="sm" onclick={() => onremover?.(busca.id)} aria-label="Remover busca">✕</Button>
		</div>
	</div>
	<div class="cartao-meta">
		{#if busca.fontes?.length}
			{#each busca.fontes as f}
				<Badge variant="gold">{f === 'curadoria' ? '🔍' : f === 'quedas' ? '📉' : f === 'novos' ? '🆕' : '⭐'} {f}</Badge>
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

<style>
	.cartao-busca {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r3) var(--r4);
		display: flex;
		flex-direction: column;
		gap: var(--r2);
	}
	.cartao-topo {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--r3);
	}
	.cartao-kws {
		display: flex;
		flex-wrap: wrap;
		gap: var(--r2);
		flex: 1;
	}
	.kw-btn {
		border: 1px solid var(--linha);
		background: var(--porcelana);
		color: var(--tinta);
		font-family: var(--ui);
		font-weight: var(--font-semi);
		font-size: var(--text-base);
		padding: var(--r1) var(--r3);
		border-radius: var(--raio-full);
		cursor: pointer;
		transition: background 0.15s ease, border-color 0.15s ease;
	}
	.kw-btn:hover, .kw-btn.kw-ativa {
		background: var(--ouro-fundo);
		border-color: var(--ouro);
		color: var(--ouro-escuro);
	}
	.cartao-acoes {
		display: flex;
		align-items: center;
		gap: var(--r1);
		flex-shrink: 0;
	}
	.cartao-meta {
		display: flex;
		flex-wrap: wrap;
		gap: var(--r2);
	}

	@media (prefers-reduced-motion: reduce) {
		.kw-btn { transition-duration: 0ms; }
	}
</style>
