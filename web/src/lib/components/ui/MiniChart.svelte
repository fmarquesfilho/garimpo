<script>
	/**
	 * MiniChart — gráfico de barras compacto para séries temporais.
	 * @prop pontos — array de { data, valor }
	 * @prop altura — altura em px (default 40)
	 * @prop formatValor — função para formatar tooltip (default: String)
	 */
	let { pontos = [], altura = 40, formatValor = String } = $props();

	let max = $derived(Math.max(...pontos.map(p => p.valor), 1));
	let min = $derived(Math.min(...pontos.map(p => p.valor), 0));
	let range = $derived(max - min || 1);
</script>

{#if pontos.length > 1}
	<div class="chart" style="height: {altura}px">
		{#each pontos as ponto, i}
			{@const h = ((ponto.valor - min) / range) * 100}
			<div
				class="bar"
				class:down={i > 0 && ponto.valor < pontos[i-1].valor}
				class:up={i > 0 && ponto.valor > pontos[i-1].valor}
				style="height: {Math.max(h, 8)}%"
				title="{ponto.data}: {formatValor(ponto.valor)}"
			></div>
		{/each}
	</div>
	<div class="labels">
		<span>{formatValor(min)}</span>
		<span>{formatValor(max)}</span>
	</div>
{:else}
	<p class="sem-dados">Aguardando dados…</p>
{/if}

<style>
	.chart {
		display: flex;
		align-items: flex-end;
		gap: 2px;
	}
	.bar {
		flex: 1;
		background: var(--ouro-claro);
		border-radius: 2px 2px 0 0;
		min-height: 3px;
		transition: height 0.2s ease;
	}
	.bar.down { background: var(--sucesso-texto); }
	.bar.up { background: var(--erro-texto); }
	.labels {
		display: flex;
		justify-content: space-between;
		font-size: 0.65rem;
		color: var(--tinta-suave);
		margin-top: 2px;
	}
	.sem-dados { font-size: var(--text-sm); color: var(--tinta-suave); font-style: italic; margin: 0; }
</style>
