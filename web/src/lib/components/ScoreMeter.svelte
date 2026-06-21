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
		<button
			class="ajuda"
			type="button"
			aria-label="O que é o teor?"
			onclick={() => (mostraAjuda = !mostraAjuda)}>?</button
		>
		<span class="valor dado">{score.toFixed(3)}</span>
		{#if mult && mult > 1}
			<span class="chip" title="bônus por estar no nicho">×{mult.toLocaleString('pt-BR')} nicho</span>
		{/if}
	</div>

	{#if mostraAjuda}
		<div class="popover">
			<p>
				<strong>Teor</strong> é o "grau de ouro" da pepita — um número de 0 a 1 que mede
				<em>o quanto o produto rende pelo esforço de divulgar</em>. Quanto maior, melhor a aposta.
			</p>
			<p>É a soma de três sinais, cada um comparado aos outros produtos do dia:</p>
			<ul>
				<li><strong>comissão</strong> — quanto da venda volta pra você;</li>
				<li>
					<strong>valor esperado</strong> — comissão × preço × vendas: o retorno provável, não só a
					porcentagem;
				</li>
				<li><strong>avaliação</strong> — a nota, como sinal de confiança.</li>
			</ul>
			<p>
				Na estratégia <strong>nicho</strong>, produtos de cosméticos/perfumaria/bem-estar ganham um
				bônus (o <span class="dado">×nicho</span>). A barra colorida abaixo mostra o peso de cada
				sinal nesta pepita.
			</p>
			<button class="fechar-pop" type="button" onclick={() => (mostraAjuda = false)}>entendi</button>
		</div>
	{/if}

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
	.ajuda {
		width: 16px;
		height: 16px;
		border-radius: 999px;
		border: 1px solid var(--linha);
		background: var(--nevoa);
		color: var(--tinta-suave);
		font-size: 0.7rem;
		font-weight: 700;
		line-height: 1;
		cursor: pointer;
		padding: 0;
		align-self: center;
	}
	.ajuda:hover {
		border-color: var(--ouro);
		color: var(--ouro);
	}
	.popover {
		font-size: 0.8rem;
		line-height: 1.45;
		color: var(--tinta);
		background: var(--porcelana);
		border: 1px solid var(--linha);
		border-left: 3px solid var(--ouro);
		border-radius: 10px;
		padding: var(--r3) var(--r4);
	}
	.popover p {
		margin: 0 0 var(--r2);
	}
	.popover ul {
		margin: 0 0 var(--r2);
		padding-left: 1.1rem;
	}
	.popover li {
		margin-bottom: 2px;
	}
	.fechar-pop {
		border: none;
		background: var(--ouro);
		color: #fff;
		font-weight: 600;
		font-size: 0.78rem;
		padding: 5px 12px;
		border-radius: 8px;
		cursor: pointer;
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
