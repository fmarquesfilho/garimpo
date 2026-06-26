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
	{#if candidato.imagem}
		<img src={candidato.imagem} alt={candidato.nome} class="thumb" loading="lazy" />
	{/if}

	<div class="corpo">
		{#if posicao != null}
			<span class="posicao dado">#{posicao}</span>
		{/if}

		<header>
			<h3>{candidato.nome}</h3>
			<div class="meta">
				<span class="cat">{candidato.categoria}</span>
				{#if candidato.suspeito}
					<span class="selo alerta">⚠ suspeito</span>
				{/if}
				{#if candidato.exploracao}
					<span class="selo explor">✦ exploração</span>
				{/if}
			</div>
		</header>

		<div class="dados">
			<div class="dado-principal">
				<span class="preco">{brl(candidato.preco)}</span>
				<span class="comissao">{pct(candidato.comissao)}</span>
			</div>
			<div class="dado-secundario">
				<span>{candidato.vendas.toLocaleString('pt-BR')} vendas</span>
				<span>★ {candidato.avaliacao.toLocaleString('pt-BR', { minimumFractionDigits: 1 })}</span>
			</div>
		</div>

		<ScoreMeter score={candidato.score} componentes={candidato.componentes} animar={destaque} />

		<footer>
			{#if onpublicar}
				<button class="primario" onclick={() => onpublicar(candidato)}>📤 Publicar</button>
			{/if}
			{#if onselecionar}
				<button class="secundario" onclick={() => onselecionar(candidato)}>Garimpar</button>
			{/if}
			<button class="ghost" onclick={copiarLink} disabled={!candidato.link}>
				{copiado ? '✓' : '🔗'}
			</button>
		</footer>
	</div>
</article>

<style>
	.cartao {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		overflow: hidden;
		box-shadow: var(--sombra);
		transition: transform 0.15s ease, box-shadow 0.15s ease;
	}
	.cartao:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 20px -8px rgba(46, 34, 38, 0.2);
	}
	.destaque {
		border-color: var(--ouro-claro);
	}

	.thumb {
		width: 100%;
		height: 180px;
		object-fit: cover;
		display: block;
		background: var(--porcelana);
	}

	.corpo {
		padding: var(--r4);
		display: flex;
		flex-direction: column;
		gap: var(--r3);
	}

	.posicao {
		font-size: var(--text-xs);
		font-weight: 700;
		color: var(--tinta-suave);
		opacity: 0.6;
	}

	header h3 {
		font-size: 1rem;
		font-weight: 600;
		line-height: 1.3;
		margin: 0;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.meta {
		display: flex;
		align-items: center;
		gap: 6px;
		margin-top: 4px;
		flex-wrap: wrap;
	}
	.cat {
		font-size: var(--text-xs);
		font-weight: 600;
		color: var(--rosa);
		text-transform: lowercase;
	}
	.selo {
		font-size: 0.65rem;
		font-weight: 700;
		padding: 1px 6px;
		border-radius: var(--raio-full);
	}
	.selo.alerta {
		background: var(--erro-fundo);
		color: var(--erro-texto);
	}
	.selo.explor {
		background: var(--porcelana);
		color: var(--tinta-suave);
	}

	.dados {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
	}
	.dado-principal {
		display: flex;
		align-items: baseline;
		gap: var(--r2);
	}
	.preco {
		font-size: 1.1rem;
		font-weight: 700;
		font-family: var(--mono);
	}
	.comissao {
		font-size: var(--text-sm);
		font-weight: 700;
		color: var(--ouro);
	}
	.dado-secundario {
		display: flex;
		gap: var(--r3);
		font-size: var(--text-xs);
		color: var(--tinta-suave);
	}

	footer {
		display: flex;
		gap: var(--r2);
		margin-top: var(--r2);
	}
	footer button {
		border-radius: var(--raio-sm);
		padding: 8px 14px;
		font-size: var(--text-sm);
		font-weight: 600;
		border: 1px solid transparent;
	}
	.primario {
		background: var(--ouro);
		color: var(--branco);
		flex: 1;
	}
	.primario:hover { background: var(--ouro-hover); }
	.secundario {
		background: transparent;
		border-color: var(--linha);
		color: var(--tinta);
	}
	.secundario:hover { border-color: var(--ouro); color: var(--ouro); }
	.ghost {
		background: transparent;
		color: var(--tinta-suave);
		padding: 8px 10px;
	}
	.ghost:hover:not(:disabled) { color: var(--ouro); }
	.ghost:disabled { opacity: 0.3; cursor: not-allowed; }

	@media (max-width: 420px) {
		.thumb { height: 140px; }
		.dados { flex-direction: column; gap: 4px; }
	}
</style>
