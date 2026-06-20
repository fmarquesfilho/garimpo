<script>
	// Barra de "teor" — assinatura visual. O preenchimento dourado encoda o
	// score; abaixo, a decomposição em componentes (explicabilidade).
	let { score = 0, componentes = {}, animar = true } = $props();

	const rotulos = {
		comissao: 'comissão',
		valor_esperado: 'valor esperado',
		avaliacao: 'avaliação',
		demanda: 'demanda'
	};
	const cores = {
		comissao: 'var(--ouro)',
		valor_esperado: 'var(--ouro-claro)',
		avaliacao: 'var(--rosa)',
		demanda: 'var(--ardosia)'
	};

	// componentes aditivos (exclui o multiplicador de nicho)
	const partes = $derived(
		Object.entries(componentes)
			.filter(([k]) => k !== 'multiplicador_nicho')
			.map(([k, v]) => ({ chave: k, rotulo: rotulos[k] ?? k, valor: Math.max(0, v), cor: cores[k] ?? 'var(--ardosia)' }))
	);
	const somaPartes = $derived(partes.reduce((s, p) => s + p.valor, 0) || 1);
	const mult = $derived(componentes.multiplicador_nicho);
	const largura = $derived(Math.min(100, Math.max(0, score * 100)));
</script>

<div class="teor">
	<div class="cabeca">
		<span class="rotulo">teor</span>
		<span class="valor dado">{score.toFixed(3)}</span>
		{#if mult && mult > 1}
			<span class="chip" title="bônus por estar no nicho">×{mult.toLocaleString('pt-BR')} nicho</span>
		{/if}
	</div>

	<div class="trilho" role="meter" aria-valuemin="0" aria-valuemax="1" aria-valuenow={score} aria-label="teor">
		<div class="ouro" class:animar style="--w: {largura}%"></div>
	</div>

	<div class="quebra" aria-hidden="true">
		{#each partes as p}
			<span class="seg" style="flex: {p.valor / somaPartes}; background: {p.cor}" title="{p.rotulo}: {p.valor.toFixed(3)}"></span>
		{/each}
	</div>
	<ul class="legenda">
		{#each partes as p}
			<li><span class="ponto" style="background: {p.cor}"></span>{p.rotulo}</li>
		{/each}
	</ul>
</div>

<style>
	.teor {
		display: flex;
		flex-direction: column;
		gap: var(--r2);
	}
	.cabeca {
		display: flex;
		align-items: baseline;
		gap: var(--r2);
	}
	.valor {
		font-size: 1.05rem;
		font-weight: 700;
		color: var(--ouro);
	}
	.chip {
		margin-left: auto;
		font-size: 0.68rem;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 999px;
		background: var(--ouro-fundo);
		color: #7a5a1e;
	}
	.trilho {
		height: 10px;
		border-radius: 999px;
		background: var(--linha);
		overflow: hidden;
	}
	.ouro {
		height: 100%;
		width: var(--w);
		border-radius: 999px;
		background: linear-gradient(90deg, var(--ouro-claro), var(--ouro));
	}
	.ouro.animar {
		animation: encher 0.7s cubic-bezier(0.2, 0.7, 0.2, 1) both;
	}
	@keyframes encher {
		from {
			width: 0;
		}
		to {
			width: var(--w);
		}
	}
	.quebra {
		display: flex;
		gap: 2px;
		height: 5px;
	}
	.seg {
		border-radius: 999px;
		min-width: 3px;
	}
	.legenda {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-wrap: wrap;
		gap: var(--r3);
		font-size: 0.72rem;
		color: var(--tinta-suave);
	}
	.legenda li {
		display: flex;
		align-items: center;
		gap: 5px;
	}
	.ponto {
		width: 8px;
		height: 8px;
		border-radius: 999px;
		display: inline-block;
	}
</style>
