<script>
	import ScoreMeter from './ScoreMeter.svelte';
	import { buscarOrigemProduto } from '$lib/api.js';
	import { onMount } from 'svelte';

	let { candidato, posicao = null, destaque = false, onselecionar = null, onpublicar = null } = $props();

	const brl = (v) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
	const pct = (v) => `${(v * 100).toLocaleString('pt-BR', { maximumFractionDigits: 0 })}%`;

	// Origem do produto — busca automática via API pública se não veio preenchido
	let origemResolvida = $state(candidato.origem || '');
	let marcaResolvida = $state('');
	let buscandoOrigem = $state(false);

	onMount(() => {
		if (!origemResolvida && candidato.id) {
			resolverOrigem();
		}
	});

	async function resolverOrigem() {
		// Extrai shopId do link do produto (formato: -i.SHOPID.ITEMID)
		const shopId = extrairShopId(candidato);
		if (!shopId) return;

		buscandoOrigem = true;
		try {
			const r = await buscarOrigemProduto({ itemId: candidato.id, shopId });
			if (r.origem) origemResolvida = r.origem;
			if (r.marca) marcaResolvida = r.marca;
		} catch {
			// Falha silenciosa — badge simplesmente não aparece
		} finally {
			buscandoOrigem = false;
		}
	}

	function extrairShopId(c) {
		// Campo direto da API
		if (c.loja_id) return String(c.loja_id);
		// Tenta extrair do link (shopee.com.br/...-i.SHOPID.ITEMID)
		if (c.link) {
			const m = c.link.match(/-i\.(\d+)\.\d+/);
			if (m) return m[1];
			// Formato alternativo: shopid no query param
			try {
				const u = new URL(c.link, 'https://shopee.com.br');
				if (u.searchParams.has('shop_id')) return u.searchParams.get('shop_id');
			} catch { /* URL inválida */ }
		}
		// Se veio como campo separado
		if (c.shop_id) return String(c.shop_id);
		return null;
	}

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
		<a href={candidato.link || '#'} target="_blank" rel="noopener" class="thumb-link">
			<img src={candidato.imagem} alt={candidato.nome} class="thumb" loading="lazy" />
		</a>
	{/if}

	<div class="corpo">
		{#if posicao != null}
			<span class="posicao dado">#{posicao}</span>
		{/if}

		<header>
			<h3>{candidato.nome}</h3>
			<div class="meta">
				{#if candidato.loja}
					<span class="loja">🏪 {candidato.loja}</span>
				{/if}
				{#if origemResolvida}
					<span class="selo origem">{#if origemResolvida === 'Coreia'}🇰🇷{:else if origemResolvida === 'Japão'}🇯🇵{:else if origemResolvida === 'China'}🇨🇳{/if} {origemResolvida}</span>
				{:else if buscandoOrigem}
					<span class="selo origem-loading">⏳</span>
				{/if}
				{#if marcaResolvida}
					<span class="selo marca">🏷️ {marcaResolvida}</span>
				{/if}
				{#if candidato.categoria}
					<span class="cat">{candidato.categoria}</span>
				{/if}
				{#if candidato.suspeito}
					<span class="selo alerta">⚠ suspeito</span>
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
				<button class="publicar-btn" onclick={() => onpublicar(candidato)}>📤 Publicar</button>
			{/if}
			<button class="ghost" onclick={copiarLink} disabled={!candidato.link}>
				{copiado ? '✓ Copiado' : '🔗 Link'}
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
	.thumb-link {
		display: block;
		text-decoration: none;
	}
	.thumb-link:hover .thumb {
		opacity: 0.9;
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
	.loja {
		font-size: var(--text-xs);
		font-weight: 600;
		color: var(--tinta-suave);
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
	.selo.origem {
		background: var(--sucesso-fundo);
		color: var(--sucesso-texto);
	}
	.selo.origem-loading {
		background: var(--porcelana);
		color: var(--tinta-suave);
		animation: pulse 1.2s ease-in-out infinite;
	}
	.selo.marca {
		background: var(--ouro-fundo);
		color: var(--ouro-escuro);
	}
	@keyframes pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.4; }
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
	.publicar-btn {
		background: var(--ouro-fundo);
		border-color: var(--ouro-claro);
		color: var(--ouro-escuro);
		flex: 1;
	}
	.publicar-btn:hover { background: var(--ouro-claro); }
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
