<script>
	/**
	 * MiniChart — gráfico de barras compacto para séries temporais.
	 * @prop pontos — array de { data, valor }
	 * @prop altura — altura em px (default 40)
	 * @prop formatValor — função para formatar tooltip
	 */
	/** @type {(v: number) => string} */
	let formatDefault = (v) => String(v);
	let { pontos = [], altura = 40, formatValor = formatDefault, ...rest } = $props();

	let max = $derived(Math.max(...pontos.map((p) => p.valor), 1));
	let min = $derived(Math.min(...pontos.map((p) => p.valor), 0));
	let range = $derived(max - min || 1);
</script>

{#if pontos.length > 1}
	<div class="flex items-end gap-0.5" style="height: {altura}px" {...rest}>
		{#each pontos as ponto, i}
			{@const h = ((ponto.valor - min) / range) * 100}
			<div
				class="min-h-[3px] flex-1 rounded-t-sm bg-primary/30 transition-[height] duration-200 ease-linear motion-reduce:transition-none"
				class:!bg-sucesso={i > 0 && ponto.valor < pontos[i - 1].valor}
				class:!bg-destructive={i > 0 && ponto.valor > pontos[i - 1].valor}
				style="height: {Math.max(h, 8)}%"
				title="{ponto.data}: {formatValor(ponto.valor)}"
			></div>
		{/each}
	</div>
	<div class="mt-0.5 flex justify-between text-xs text-muted-foreground">
		<span>{formatValor(min)}</span>
		<span>{formatValor(max)}</span>
	</div>
{:else}
	<p class="m-0 text-sm italic text-muted-foreground">Aguardando dados…</p>
{/if}
