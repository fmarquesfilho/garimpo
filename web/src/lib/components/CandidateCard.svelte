<script>
	import ScoreMeter from './ScoreMeter.svelte';

	let { candidato, posicao = null, destaque = false, onselecionar = null, onpublicar = null } = $props();

	const brl = (v) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
	const pct = (v) => `${(v * 100).toLocaleString('pt-BR', { maximumFractionDigits: 0 })}%`;

	let copiado = $state(false);
	async function copiarLink() {
		if (!candidato.link) return;
		try {
			await navigator.clipboard.writeText(candidato.link);
			copiado = true;
			setTimeout(() => (copiado = false), 1600);
		} catch {
			copiado = false;
		}
	}
</script>

<article class="cartao" class:destaque>
	{#if posicao != null}
		<div class="posicao dado">{posicao}</div>
	{/if}

	<div class="cartao-corpo">
		{#if candidato.imagem}
			<img src={candidato.imagem} alt={candidato.nome} class="thumb" loading="lazy" />
		{/if}
		<div class="cartao-info">
			<header>
				<span class="cat">{candidato.categoria}</span>
				<h3>{candidato.nome}</h3>
				{#if candidato.suspeito || candidato.exploracao}
					<div class="selos">
						{#if candidato.suspeito}
							<span class="selo alerta" title="Comissão alta, mas sem vendas/nota — pode ser produto-fantasma">
								⚠ suspeito
							</span>
						{/if}
						{#if candidato.exploracao}
							<span class="selo explor" title="Sorteado fora do topo para testar o que converte (exploração)">
								✦ exploração
							</span>
						{/if}
					</div>
				{/if}
			</header>

			<dl class="laudo">
				<div>
					<dt class="rotulo">preço</dt>
					<dd class="dado">{brl(candidato.preco)}</dd>
				</div>
				<div>
					<dt class="rotulo">comissão</dt>
					<dd class="dado ouro">{pct(candidato.comissao)}</dd>
				</div>
				<div>
					<dt class="rotulo">vendas</dt>
					<dd class="dado">{candidato.vendas.toLocaleString('pt-BR')}</dd>
				</div>
				<div>
					<dt class="rotulo">nota</dt>
					<dd class="dado">{candidato.avaliacao.toLocaleString('pt-BR', { minimumFractionDigits: 1 })}</dd>
				</div>
			</dl>
		</div>
	</div>

	<ScoreMeter score={candidato.score} componentes={candidato.componentes} animar={destaque} />

	<footer>
		{#if onselecionar}
			<button class="primario" onclick={() => onselecionar(candidato)}>Garimpar</button>
		{/if}
		{#if onpublicar}
			<button class="publicar" onclick={() => onpublicar(candidato)}>Publicar</button>
		{/if}
		<button class="secundario" onclick={copiarLink} disabled={!candidato.link}>
			{copiado ? 'Copiado' : 'Copiar link'}
		</button>
	</footer>
</article>

<style>
	.cartao {
		position: relative;
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r6);
		display: flex;
		flex-direction: column;
		gap: var(--r4);
		box-shadow: var(--sombra);
		transition:
			transform 0.18s ease,
			box-shadow 0.18s ease;
	}
	.cartao:hover {
		transform: translateY(-3px);
		box-shadow: 0 1px 2px rgba(43, 29, 46, 0.05), 0 18px 40px -18px rgba(43, 29, 46, 0.3);
	}
	.destaque {
		border-color: var(--ouro-claro);
		background: linear-gradient(180deg, #fffaf1, var(--nevoa));
	}
	.cartao-corpo {
		display: flex;
		gap: var(--r4);
	}
	.thumb {
		width: 80px;
		height: 80px;
		border-radius: var(--raio-sm);
		object-fit: cover;
		flex-shrink: 0;
		background: var(--porcelana);
	}
	.cartao-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: var(--r3);
	}
	.posicao {
		position: absolute;
		top: var(--r4);
		right: var(--r4);
		font-size: 0.85rem;
		font-weight: 700;
		color: var(--tinta-suave);
		opacity: 0.5;
	}
	.selos {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin-top: 6px;
	}
	.selo {
		font-size: 0.68rem;
		font-weight: 700;
		padding: 2px 8px;
		border-radius: 999px;
	}
	.selo.alerta {
		background: color-mix(in srgb, var(--erro-texto) 14%, var(--porcelana));
		color: var(--erro-texto);
	}
	.selo.explor {
		background: color-mix(in srgb, var(--tinta-suave) 14%, var(--porcelana));
		color: var(--tinta-suave);
	}
	.cat {
		font-size: 0.72rem;
		font-weight: 600;
		letter-spacing: 0.04em;
		color: var(--rosa);
		text-transform: lowercase;
	}
	h3 {
		font-size: 1.35rem;
		margin-top: 4px;
		max-width: 22ch;
	}
	.destaque h3 {
		font-size: 1.7rem;
	}
	.laudo {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: var(--r3);
		margin: 0;
		padding: var(--r3) 0;
		border-top: 1px solid var(--linha);
		border-bottom: 1px solid var(--linha);
	}
	.laudo div {
		display: flex;
		flex-direction: column;
		gap: 3px;
	}
	.laudo dd {
		margin: 0;
		font-size: 0.95rem;
		font-weight: 700;
	}
	.laudo dd.ouro {
		color: var(--ouro);
	}
	footer {
		display: flex;
		flex-wrap: wrap;
		gap: var(--r2);
		margin-top: auto;
	}
	button {
		border-radius: 10px;
		padding: 10px 16px;
		font-size: 0.9rem;
		font-weight: 600;
		border: 1px solid transparent;
		transition: background 0.15s ease, border-color 0.15s ease;
	}
	.primario {
		background: var(--ouro);
		color: #fff;
		flex: 1;
	}
	.primario:hover {
		background: #a3782f;
	}
	.publicar {
		background: var(--rosa);
		color: #fff;
	}
	.publicar:hover {
		background: #8f4c62;
	}
	.secundario {
		background: transparent;
		border-color: var(--linha);
		color: var(--tinta);
	}
	.secundario:hover:not(:disabled) {
		border-color: var(--tinta-suave);
	}
	.secundario:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}
	@media (max-width: 420px) {
		.laudo {
			grid-template-columns: repeat(2, 1fr);
		}
	}
</style>
