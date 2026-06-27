<script>
	/** Cartão de uma busca salva. Exibe keywords, badges e ações. */
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
				<button type="button" class="btn-prox" onclick={() => onproximakw?.(busca)} title="Próxima keyword">→</button>
			{/if}
			<button type="button" class="x" onclick={() => onremover?.(busca.id)} aria-label="remover busca">✕</button>
		</div>
	</div>
	<div class="cartao-meta">
		{#if busca.fontes?.length}
			{#each busca.fontes as f}
				<span class="badge fonte">{f === 'curadoria' ? '🔍' : f === 'quedas' ? '📉' : f === 'novos' ? '🆕' : '⭐'} {f}</span>
			{/each}
		{:else}
			<span class="badge">{busca.estrategia ?? 'nicho'}</span>
		{/if}
		{#if busca.cron}
			<span class="badge cron" title="coleta periódica: {busca.cron}">⏱ agendada</span>
		{/if}
		{#if busca.categorias?.length}
			{#each busca.categorias as cat}
				<span class="badge cat">{cat}</span>
			{/each}
		{:else if busca.categoria}
			<span class="badge cat">{busca.categoria}</span>
		{/if}
		{#if busca.shop_ids?.length}
			<span class="badge shop" title="monitorando lojas">🏪 {busca.shop_ids.length} {busca.shop_ids.length === 1 ? 'loja' : 'lojas'}</span>
		{/if}
		{#if busca.dias_janela && busca.dias_janela !== 7}
			<span class="badge">janela: {busca.dias_janela}d</span>
		{/if}
	</div>
</div>

<style>
	.cartao-busca {
		background: var(--nevoa); border: 1px solid var(--linha);
		border-radius: var(--raio); padding: var(--r3) var(--r4);
		display: flex; flex-direction: column; gap: var(--r2);
	}
	.cartao-topo { display: flex; align-items: flex-start; justify-content: space-between; gap: var(--r3); }
	.cartao-kws { display: flex; flex-wrap: wrap; gap: var(--r2); flex: 1; }
	.kw-btn {
		border: 1px solid var(--linha); background: var(--porcelana);
		color: var(--tinta); font-weight: 600; font-size: 0.88rem;
		padding: 5px 12px; border-radius: 999px; cursor: pointer;
	}
	.kw-btn:hover, .kw-btn.kw-ativa { background: var(--ouro-fundo); border-color: var(--ouro); color: #7a5a1e; }
	.cartao-acoes { display: flex; align-items: center; gap: var(--r2); flex-shrink: 0; }
	.btn-prox {
		border: 1px solid var(--linha); background: transparent;
		color: var(--tinta-suave); font-size: 0.9rem; padding: 3px 8px;
		border-radius: 8px; cursor: pointer;
	}
	.btn-prox:hover { color: var(--tinta); border-color: var(--tinta-suave); }
	.cartao-meta { display: flex; flex-wrap: wrap; gap: var(--r2); }
	.badge {
		font-size: 0.72rem; padding: 2px 8px; border-radius: 999px;
		background: var(--porcelana); border: 1px solid var(--linha);
		color: var(--tinta-suave);
	}
	.badge.cron { color: color-mix(in srgb, var(--ouro) 70%, var(--tinta-suave)); }
	.badge.fonte { background: var(--ouro-fundo); color: var(--ouro-escuro); border-color: var(--ouro-claro); }
	.badge.cat { color: var(--rosa); border-color: color-mix(in srgb, var(--rosa) 30%, var(--linha)); }
	.badge.shop {
		color: var(--rosa);
		border-color: color-mix(in srgb, var(--rosa) 30%, var(--linha));
		background: color-mix(in srgb, var(--rosa) 8%, var(--porcelana));
	}
	.x {
		border: none; background: transparent; color: var(--tinta-suave);
		font-size: 0.72rem; cursor: pointer; padding: 2px 4px;
	}
	.x:hover { color: var(--erro-texto); }
</style>
