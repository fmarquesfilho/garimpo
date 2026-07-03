<script>
	/**
	 * ProductCard — componente unificado de exibição de produto.
	 *
	 * Layouts:
	 *   "full"    — card vertical com imagem grande (busca/curadoria)
	 *   "compact" — linha horizontal com thumb pequena (lista de lojas)
	 *   "feed"    — card de feed com borda lateral (oportunidades)
	 *
	 * Props de dados:
	 *   produto — objeto com: nome, preco, comissao, vendas, avaliacao, imagem, link, loja, origem, categoria, score, etc.
	 *
	 * Props de configuração:
	 *   layout       — "full" | "compact" | "feed" (default: "full")
	 *   posicao      — número de posição no ranking (opcional, layout full)
	 *   destaque     — borda dourada (opcional, layout full)
	 *   variacao     — { tipo: "queda"|"alta"|"novo", pct, preco_anterior, preco_atual, detectado_em }
	 *   nomeLoja     — override do nome da loja (para feeds onde vem de contexto externo)
	 *   mostrarScore — exibe ScoreMeter (default: true em full, false em outros)
	 *   mostrarLink  — exibe botão de copiar link (default: true em full)
	 *   onpublicar   — callback ao clicar Publicar
	 *   onfavoritar  — callback ao clicar Favoritar (futuro)
	 */
	import ScoreMeter from './ScoreMeter.svelte';
	import { brl, pct, tempoAtras } from '$lib/formatters.js';
	import { favoritos } from '$lib/favoritos.js';

	let {
		produto,
		layout = 'full',
		posicao = null,
		destaque = false,
		variacao = null,
		nomeLoja = '',
		mostrarScore = undefined,
		mostrarLink = undefined,
		onpublicar = null,
		onfavoritar = null
	} = $props();

	// Defaults baseados no layout
	let exibirScore = $derived(mostrarScore ?? layout === 'full');
	let exibirLink = $derived(mostrarLink ?? layout === 'full');

	let loja = $derived(nomeLoja || produto.loja || '');
	let produtoId = $derived(produto.produto_id || produto.id || '');
	let isFav = $derived($favoritos.some((f) => f.produto_id === produtoId || f.id === produtoId));

	function tempoRestante(iso) {
		if (!iso) return '';
		const diff = new Date(iso).getTime() - Date.now();
		if (diff <= 0) return 'expirado';
		const horas = Math.floor(diff / 3600000);
		if (horas < 24) return `${horas}h`;
		const dias = Math.floor(horas / 24);
		return `${dias}d`;
	}

	let copiado = $state(false);
	async function copiarLink() {
		if (!produto.link) return;
		try {
			await navigator.clipboard.writeText(produto.link);
			copiado = true;
			setTimeout(() => (copiado = false), 1600);
		} catch {
			copiado = false;
		}
	}
</script>

{#if layout === 'full'}
	<!-- ═══ LAYOUT FULL — card vertical com imagem grande ═══ -->
	<article class="cartao" class:destaque>
		{#if produto.imagem}
			<a href={produto.link || '#'} target="_blank" rel="noopener" class="thumb-link">
				<img src={produto.imagem} alt={produto.nome} class="thumb" loading="lazy" />
			</a>
		{/if}
		<div class="corpo">
			{#if posicao != null}
				<span class="posicao">#{posicao}</span>
			{/if}
			<header>
				<h3>{produto.nome}</h3>
				<div class="meta">
					{#if loja}<span class="loja-nome">🏪 {loja}</span>{/if}
					{#if produto.origem}
						<span class="selo origem"
							>{#if produto.origem === 'Coreia'}🇰🇷{:else if produto.origem === 'Japão'}🇯🇵{:else if produto.origem === 'China'}🇨🇳{/if}
							{produto.origem}</span
						>
					{/if}
					{#if produto.desconto > 0 && produto.desconto <= 1}
						<span class="selo desconto">🔥 {Math.round(produto.desconto * 100)}% OFF</span>
					{:else if produto.desconto > 1 && produto.desconto <= 100}
						<span class="selo desconto">🔥 {Math.round(produto.desconto)}% OFF</span>
					{/if}
					{#if produto.oferta_expira}
						<span class="selo expira" title="Expira em {new Date(produto.oferta_expira).toLocaleDateString('pt-BR')}"
							>⏳ {tempoRestante(produto.oferta_expira)}</span
						>
					{/if}
					{#if produto.categoria}<span class="cat">{produto.categoria}</span>{/if}
					{#if produto.suspeito}<span
							class="selo alerta"
							title="Comissão alta com poucas vendas — pode ser produto sem tração real. Avalie antes de publicar."
							>⚠ suspeito</span
						>{/if}
				</div>
			</header>
			<div class="dados">
				<div class="dado-principal">
					<span class="preco">{brl(produto.preco)}</span>
					<span class="comissao">{pct(produto.comissao)}</span>
				</div>
				<div class="dado-secundario">
					<span>{produto.vendas?.toLocaleString('pt-BR') ?? 0} vendas</span>
					<span>★ {produto.avaliacao?.toLocaleString('pt-BR', { minimumFractionDigits: 1 }) ?? '—'}</span>
				</div>
			</div>
			{#if exibirScore && produto.score}
				<ScoreMeter score={produto.score} componentes={produto.componentes} animar={destaque} />
			{/if}
			<footer>
				{#if onpublicar}
					<button class="publicar-btn" onclick={() => onpublicar(produto)}>📤 Publicar</button>
				{/if}
				{#if onfavoritar}
					<button
						class="ghost"
						class:favoritado={isFav}
						onclick={() => onfavoritar(produto)}
						title={isFav ? 'Remover dos favoritos' : 'Favoritar'}>{isFav ? '★' : '☆'}</button
					>
				{/if}
				{#if exibirLink}
					<button class="ghost" onclick={copiarLink} disabled={!produto.link}>
						{copiado ? '✓ Copiado' : '🔗 Link'}
					</button>
				{/if}
			</footer>
		</div>
	</article>
{:else if layout === 'compact'}
	<!-- ═══ LAYOUT COMPACT — linha horizontal ═══ -->
	<div class="compact-card">
		{#if produto.imagem}
			<img src={produto.imagem} alt={produto.nome} class="compact-thumb" loading="lazy" />
		{/if}
		<div class="compact-info">
			<h4>{produto.nome}</h4>
			<div class="compact-dados">
				<span class="compact-preco">{brl(produto.preco)}</span>
				<span class="compact-comissao">{pct(produto.comissao)}</span>
				{#if produto.vendas}<span>{produto.vendas} vendas</span>{/if}
				{#if produto.avaliacao}<span>★ {produto.avaliacao?.toFixed(1)}</span>{/if}
			</div>
			{#if loja}<span class="compact-loja">🏪 {loja}</span>{/if}
		</div>
		<div class="compact-acoes">
			{#if onfavoritar}
				<button
					class="btn-mini"
					class:favoritado={isFav}
					onclick={() => onfavoritar(produto)}
					title={isFav ? 'Remover dos favoritos' : 'Favoritar'}>{isFav ? '★' : '☆'}</button
				>
			{/if}
			{#if onpublicar}
				<button class="btn-mini" onclick={() => onpublicar(produto)} title="Publicar">📤</button>
			{/if}
		</div>
	</div>
{:else if layout === 'feed'}
	<!-- ═══ LAYOUT FEED — card com borda lateral (oportunidades) ═══ -->
	<div class="feed-card {variacao?.tipo ?? 'novo'}">
		<div class="feed-header">
			{#if variacao?.tipo === 'queda'}
				<span class="badge badge-queda">↓ {Math.abs(variacao.pct * 100).toFixed(0)}%</span>
			{:else if variacao?.tipo === 'alta'}
				<span class="badge badge-alta">↑ {Math.abs(variacao.pct * 100).toFixed(0)}%</span>
			{:else}
				<span class="badge badge-novo">Novo</span>
			{/if}
			{#if loja}<span class="feed-loja">{loja}</span>{/if}
			{#if variacao?.detectado_em}<span class="feed-tempo">{tempoAtras(variacao.detectado_em)}</span>{/if}
		</div>

		<div class="feed-body">
			{#if produto.imagem}
				<img src={produto.imagem} alt={produto.nome} class="feed-thumb" loading="lazy" />
			{/if}
			<div class="feed-content">
				<h3 class="feed-nome">{produto.nome}</h3>
				<div class="feed-precos">
					{#if variacao?.tipo === 'queda'}
						<span class="preco-antes">{brl(variacao.preco_anterior)}</span>
						<span class="seta">→</span>
						<span class="preco-atual destaque-queda">{brl(variacao.preco_atual)}</span>
						<span class="economia">(-{brl(variacao.preco_anterior - variacao.preco_atual)})</span>
					{:else if variacao?.tipo === 'alta'}
						<span class="preco-antes">{brl(variacao.preco_anterior)}</span>
						<span class="seta">→</span>
						<span class="preco-atual destaque-alta">{brl(variacao.preco_atual)}</span>
					{:else}
						<span class="preco-atual">{brl(produto.preco)}</span>
						{#if produto.comissao > 0}<span class="feed-comissao">{pct(produto.comissao)}</span>{/if}
						{#if produto.vendas > 0}<span class="feed-vendas">{produto.vendas} vendas</span>{/if}
					{/if}
				</div>
			</div>
		</div>

		<div class="feed-acoes">
			{#if onpublicar}
				<button class="btn-publicar" onclick={() => onpublicar(produto)}>📤 Publicar</button>
			{/if}
			{#if onfavoritar}
				<button
					class="btn-fav"
					class:favoritado={isFav}
					onclick={() => onfavoritar(produto)}
					title={isFav ? 'Remover dos favoritos' : 'Favoritar'}>{isFav ? '★' : '☆'}</button
				>
			{/if}
		</div>
	</div>
{/if}

<style>
	/* ═══ FULL LAYOUT ═══ */
	.cartao {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		overflow: hidden;
		box-shadow: var(--sombra);
		transition:
			transform 0.15s ease,
			box-shadow 0.15s ease;
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
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.meta {
		display: flex;
		align-items: center;
		gap: 4px 6px;
		margin-top: 4px;
		flex-wrap: wrap;
	}
	.cat {
		font-size: var(--text-xs);
		font-weight: 600;
		color: var(--rosa);
		text-transform: lowercase;
	}
	.loja-nome {
		font-size: var(--text-xs);
		font-weight: 600;
		color: var(--tinta-suave);
		max-width: 140px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
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
	.selo.desconto {
		background: var(--erro-fundo);
		color: var(--erro-texto);
	}
	.selo.expira {
		background: var(--ouro-fundo);
		color: var(--ouro-escuro);
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
	.publicar-btn:hover {
		background: var(--ouro-claro);
	}
	.ghost {
		background: transparent;
		color: var(--tinta-suave);
		padding: 8px 10px;
	}
	.ghost:hover:not(:disabled) {
		color: var(--ouro);
	}
	.ghost:disabled {
		opacity: 0.3;
		cursor: not-allowed;
	}
	.ghost.favoritado {
		color: var(--ouro);
	}

	/* ═══ COMPACT LAYOUT ═══ */
	.compact-card {
		display: flex;
		gap: var(--r3);
		padding: var(--r3) var(--r4);
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		background: var(--branco);
		align-items: center;
	}
	.compact-thumb {
		width: 56px;
		height: 56px;
		border-radius: 8px;
		object-fit: cover;
		flex-shrink: 0;
	}
	.compact-info {
		flex: 1;
		min-width: 0;
	}
	.compact-info h4 {
		font-size: 0.9rem;
		margin: 0 0 4px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.compact-dados {
		display: flex;
		flex-wrap: wrap;
		gap: var(--r2);
		font-size: 0.78rem;
		color: var(--tinta-suave);
	}
	.compact-preco {
		font-weight: 700;
		color: var(--ouro);
	}
	.compact-comissao {
		font-weight: 600;
	}
	.compact-loja {
		font-size: 0.72rem;
		color: var(--tinta-suave);
	}
	.compact-acoes {
		display: flex;
		gap: 4px;
		flex-shrink: 0;
	}
	.btn-mini {
		border: 1px solid var(--linha);
		background: var(--porcelana);
		border-radius: 8px;
		width: 36px;
		height: 36px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		font-size: 1rem;
	}
	.btn-mini:hover {
		border-color: var(--ouro);
	}
	.btn-mini.favoritado {
		color: var(--ouro);
		border-color: var(--ouro);
	}

	/* ═══ FEED LAYOUT ═══ */
	.feed-card {
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r4);
		background: var(--branco);
		transition: border-color 0.15s;
	}
	.feed-card:hover {
		border-color: var(--ouro-claro);
	}
	.feed-card.queda {
		border-left: 3px solid var(--sucesso-texto);
	}
	.feed-card.alta {
		border-left: 3px solid var(--erro-texto);
	}
	.feed-card.novo {
		border-left: 3px solid var(--ouro);
	}
	.feed-header {
		display: flex;
		align-items: center;
		gap: var(--r2);
		margin-bottom: 6px;
	}
	.badge {
		padding: 2px 8px;
		border-radius: var(--raio-full);
		font-size: 0.72rem;
		font-weight: 700;
	}
	.badge-queda {
		background: var(--sucesso-fundo);
		color: var(--sucesso-texto);
	}
	.badge-alta {
		background: var(--erro-fundo);
		color: var(--erro-texto);
	}
	.badge-novo {
		background: var(--ouro-fundo);
		color: var(--ouro-escuro);
	}
	.feed-loja {
		font-size: 0.72rem;
		color: var(--tinta-suave);
		background: var(--porcelana);
		padding: 1px 6px;
		border-radius: 4px;
	}
	.feed-tempo {
		font-size: 0.72rem;
		color: var(--tinta-suave);
		margin-left: auto;
	}
	.feed-body {
		display: flex;
		gap: var(--r3);
	}
	.feed-thumb {
		width: 64px;
		height: 64px;
		border-radius: 8px;
		object-fit: cover;
		flex-shrink: 0;
	}
	.feed-content {
		flex: 1;
		min-width: 0;
	}
	.feed-nome {
		font-size: 0.95rem;
		font-weight: 600;
		margin: 0 0 8px;
		line-height: 1.3;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.feed-precos {
		display: flex;
		align-items: center;
		gap: var(--r2);
		font-size: 0.88rem;
		flex-wrap: wrap;
	}
	.preco-antes {
		text-decoration: line-through;
		color: var(--tinta-suave);
	}
	.seta {
		color: var(--tinta-suave);
		font-size: 0.8rem;
	}
	.preco-atual {
		font-weight: 700;
	}
	.destaque-queda {
		color: var(--sucesso-texto);
	}
	.destaque-alta {
		color: var(--erro-texto);
	}
	.economia {
		font-size: 0.78rem;
		color: var(--sucesso-texto);
		font-weight: 600;
	}
	.feed-comissao {
		font-size: 0.78rem;
		color: var(--ouro);
		font-weight: 600;
	}
	.feed-vendas {
		font-size: 0.78rem;
		color: var(--tinta-suave);
	}
	.feed-acoes {
		margin-top: 10px;
		display: flex;
		gap: var(--r2);
	}
	.btn-publicar {
		padding: 6px 14px;
		border: 1px solid var(--ouro-claro);
		background: var(--ouro-fundo);
		color: var(--ouro-escuro);
		border-radius: var(--raio-sm);
		font-size: 0.82rem;
		font-weight: 600;
		cursor: pointer;
	}
	.btn-publicar:hover {
		background: var(--ouro-claro);
	}
	.btn-fav {
		padding: 6px 10px;
		border: 1px solid var(--linha);
		background: transparent;
		border-radius: var(--raio-sm);
		cursor: pointer;
	}
	.btn-fav:hover {
		border-color: var(--ouro);
	}
	.btn-fav.favoritado {
		color: var(--ouro);
		border-color: var(--ouro);
	}

	@media (max-width: 420px) {
		.thumb {
			height: 140px;
		}
		.dados {
			flex-direction: column;
			gap: 4px;
		}
		.feed-thumb {
			width: 48px;
			height: 48px;
		}
	}
</style>
