<script>
	import { quadro, COLUNAS } from '$lib/board.js';

	const brl = (v) => (v ?? 0).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });

	// índice da coluna para saber "próxima" e "anterior"
	function vizinhas(idColuna) {
		const i = COLUNAS.findIndex((c) => c.id === idColuna);
		return {
			anterior: i > 0 ? COLUNAS[i - 1] : null,
			proxima: i < COLUNAS.length - 1 ? COLUNAS[i + 1] : null
		};
	}
</script>

<section class="intro">
	<p class="rotulo">operação do dia</p>
	<h1>Quadro</h1>
	<p class="sub">
		Puxe cada pepita pelo fluxo: do que você selecionou até publicado e em análise. Os limites de
		WIP existem pra você focar — se uma coluna estoura, o gargalo está ali.
	</p>
	<button class="limpar" onclick={() => quadro.limpar()}>Limpar quadro</button>
</section>

<div class="quadro">
	{#each COLUNAS as col}
		{@const cards = $quadro[col.id]}
		{@const estourou = col.wip != null && cards.length > col.wip}
		<section class="coluna">
			<header class="cab-col">
				<h2>{col.titulo}</h2>
				<span class="contagem dado" class:estourou>
					{cards.length}{#if col.wip != null}/{col.wip}{/if}
				</span>
			</header>

			{#if cards.length === 0}
				<p class="vazio-col">Arraste pra cá puxando do passo anterior.</p>
			{/if}

			<div class="cards">
				{#each cards as card (card.id)}
					{@const v = vizinhas(col.id)}
					<article class="mini">
						<div class="mini-cab">
							<span class="cat">{card.categoria}</span>
							{#if card.estrategia}
								<span class="tag {card.estrategia}">{card.estrategia}</span>
							{/if}
						</div>
						<h3>{card.nome}</h3>
						<p class="mini-dado dado">{brl(card.preco)} · {(card.comissao * 100).toFixed(0)}% · teor {card.score?.toFixed(2)}</p>
						<div class="acoes">
							{#if v.anterior}
								<button class="passo" title="Voltar para {v.anterior.titulo}" onclick={() => quadro.mover(card.id, col.id, v.anterior.id)}>←</button>
							{/if}
							{#if card.link}
								<a class="passo link" href={card.link} target="_blank" rel="noreferrer" title="Abrir link">↗</a>
							{/if}
							<button class="passo" title="Remover" onclick={() => quadro.remover(card.id, col.id)}>✕</button>
							{#if v.proxima}
								<button class="passo avancar" title="Avançar para {v.proxima.titulo}" onclick={() => quadro.mover(card.id, col.id, v.proxima.id)}>→</button>
							{/if}
						</div>
					</article>
				{/each}
			</div>
		</section>
	{/each}
</div>

<style>
	.intro {
		max-width: 42rem;
		margin-bottom: var(--r8);
	}
	h1 {
		font-size: clamp(2rem, 6vw, 3rem);
		margin: var(--r2) 0 var(--r4);
	}
	.sub {
		color: var(--tinta-suave);
		margin: 0 0 var(--r4);
	}
	.limpar {
		background: transparent;
		border: 1px solid var(--linha);
		color: var(--tinta-suave);
		font-size: 0.82rem;
		font-weight: 600;
		padding: 7px 14px;
		border-radius: var(--raio-full);
	}
	.limpar:hover {
		border-color: var(--erro-texto);
		color: var(--erro-texto);
	}
	.quadro {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: var(--r4);
	}
	@media (max-width: 860px) {
		.quadro {
			grid-auto-flow: column;
			grid-template-columns: none;
			grid-auto-columns: 78%;
			overflow-x: auto;
			scroll-snap-type: x mandatory;
			padding-bottom: var(--r4);
		}
		.coluna {
			scroll-snap-align: start;
		}
	}
	.coluna {
		background: color-mix(in srgb, var(--porcelana) 60%, white);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r4);
		min-height: 160px;
	}
	.cab-col {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--r4);
	}
	.cab-col h2 {
		font-size: 1.05rem;
	}
	.contagem {
		font-size: 0.8rem;
		color: var(--tinta-suave);
		background: var(--nevoa);
		padding: 2px 8px;
		border-radius: var(--raio-full);
		border: 1px solid var(--linha);
	}
	.contagem.estourou {
		color: var(--branco);
		background: var(--erro-texto);
		border-color: var(--erro-texto);
	}
	.vazio-col {
		font-size: 0.8rem;
		color: var(--tinta-suave);
		opacity: 0.7;
		margin: var(--r2) 0;
	}
	.cards {
		display: flex;
		flex-direction: column;
		gap: var(--r3);
	}
	.mini {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		padding: var(--r3);
		box-shadow: var(--sombra);
	}
	.mini-cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: var(--r2);
	}
	.cat {
		font-size: 0.66rem;
		font-weight: 600;
		color: var(--rosa);
	}
	.tag {
		font-size: 0.62rem;
		font-weight: 700;
		padding: 1px 7px;
		border-radius: var(--raio-full);
		text-transform: lowercase;
	}
	.tag.nicho {
		background: color-mix(in srgb, var(--rosa) 18%, white);
		color: var(--rosa);
	}
	.tag.diversificada {
		background: color-mix(in srgb, var(--tinta-suave) 18%, white);
		color: var(--tinta-suave);
	}
	.mini h3 {
		font-size: 0.98rem;
		margin: 6px 0;
		line-height: 1.2;
	}
	.mini-dado {
		font-size: 0.74rem;
		color: var(--tinta-suave);
		margin: 0 0 var(--r3);
	}
	.acoes {
		display: flex;
		gap: 5px;
		align-items: center;
	}
	.passo {
		width: 30px;
		height: 30px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border: 1px solid var(--linha);
		background: var(--porcelana);
		border-radius: 8px;
		font-size: 0.85rem;
		text-decoration: none;
		color: var(--tinta);
	}
	.passo:hover {
		border-color: var(--tinta-suave);
	}
	.avancar {
		margin-left: auto;
		background: var(--tinta);
		color: var(--branco);
		border-color: var(--tinta);
	}
	.avancar:hover {
		background: var(--tinta);
	}
</style>
