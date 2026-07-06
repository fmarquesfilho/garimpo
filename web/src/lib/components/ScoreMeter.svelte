<script>
	// Barra de "teor" — assinatura visual. O preenchimento dourado encoda o
	// score; abaixo, a decomposição em componentes (explicabilidade).
	let { score = 0, componentes = {}, animar = true } = $props();
	let mostraAjuda = $state(false);

	const rotulos = {
		comissao: 'comissão',
		valor_esperado: 'valor esperado',
		avaliacao: 'avaliação',
		demanda: 'demanda'
	};
	const cores = {
		comissao: 'var(--color-primary)',
		valor_esperado: 'var(--color-ouro-claro)',
		avaliacao: 'var(--color-destructive)',
		demanda: 'var(--color-muted-foreground)'
	};

	// componentes aditivos (exclui o multiplicador de nicho)
	const partes = $derived(
		Object.entries(componentes)
			.filter(([k]) => k !== 'multiplicador_nicho')
			.map(([k, v]) => ({
				chave: k,
				rotulo: rotulos[k] ?? k,
				valor: Math.max(0, v),
				cor: cores[k] ?? 'var(--color-muted-foreground)'
			}))
	);
	const somaPartes = $derived(partes.reduce((s, p) => s + p.valor, 0) || 1);
	const mult = $derived(componentes.multiplicador_nicho);
	const largura = $derived(Math.min(100, Math.max(0, score * 100)));
</script>

<div class="flex flex-col gap-2">
	<div class="flex items-baseline gap-2">
		<span class="rotulo">teor</span>
		<button
			class="flex h-4 w-4 cursor-pointer items-center justify-center self-center rounded-full border border-border bg-card p-0 text-[0.7rem] font-bold leading-none text-muted-foreground hover:border-primary hover:text-primary"
			type="button"
			aria-label="O que é o teor?"
			onclick={() => (mostraAjuda = !mostraAjuda)}>?</button
		>
		<span class="dado text-lg font-bold text-primary">{score.toFixed(3)}</span>
		{#if mult && mult > 1}
			<span
				class="ml-auto rounded-full bg-accent px-2 py-0.5 text-xs font-semibold text-accent-foreground"
				title="bônus por estar no nicho">×{mult.toLocaleString('pt-BR')} nicho</span
			>
		{/if}
	</div>

	{#if mostraAjuda}
		<div
			class="rounded-[10px] border border-border border-l-[3px] border-l-primary bg-muted px-4 py-3 text-[0.8rem] leading-snug text-foreground"
		>
			<p class="mb-2 mt-0">
				<strong>Teor</strong> é o "grau de ouro" da pepita — um número de 0 a 1 que mede
				<em>o quanto o produto rende pelo esforço de divulgar</em>. Quanto maior, melhor a aposta.
			</p>
			<p class="mb-2 mt-0">É a soma de três sinais, cada um comparado aos outros produtos do dia:</p>
			<ul class="mb-2 mt-0 pl-4">
				<li class="mb-0.5"><strong>comissão</strong> — quanto da venda volta pra você;</li>
				<li class="mb-0.5">
					<strong>valor esperado</strong> — comissão × preço × vendas: o retorno provável, não só a porcentagem;
				</li>
				<li class="mb-0.5"><strong>avaliação</strong> — a nota, como sinal de confiança.</li>
			</ul>
			<p class="mb-2 mt-0">
				Na estratégia <strong>nicho</strong>, produtos de cosméticos/perfumaria/bem-estar ganham um bônus (o
				<span class="dado">×nicho</span>). A barra colorida abaixo mostra o peso de cada sinal nesta pepita.
			</p>
			<button
				class="cursor-pointer rounded-sm border-none bg-primary px-3 py-1 text-sm font-semibold text-primary-foreground"
				type="button"
				onclick={() => (mostraAjuda = false)}>entendi</button
			>
		</div>
	{/if}

	<div
		class="h-2.5 overflow-hidden rounded-full bg-border"
		role="meter"
		aria-valuemin="0"
		aria-valuemax="1"
		aria-valuenow={score}
		aria-label="teor"
	>
		<div
			class="h-full rounded-full bg-primary"
			class:animate-[encher_0.7s_cubic-bezier(0.2,0.7,0.2,1)_both]={animar}
			style="width: {largura}%"
		></div>
	</div>

	<div class="flex h-[5px] gap-0.5" aria-hidden="true">
		{#each partes as p}
			<span
				class="min-w-[3px] rounded-full"
				style="flex: {p.valor / somaPartes}; background: {p.cor}"
				title="{p.rotulo}: {p.valor.toFixed(3)}"
			></span>
		{/each}
	</div>
	<ul class="m-0 flex list-none flex-wrap gap-3 p-0 text-xs text-muted-foreground">
		{#each partes as p}
			<li class="flex items-center gap-1.5">
				<span class="inline-block h-2 w-2 rounded-full" style="background: {p.cor}"></span>{p.rotulo}
			</li>
		{/each}
	</ul>
</div>

<style>
	@keyframes encher {
		from {
			width: 0;
		}
		to {
			width: var(--w);
		}
	}
</style>
