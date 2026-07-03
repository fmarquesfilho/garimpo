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
	import { cn } from '$lib/utils.js';

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
	<article
		class={cn(
			'bg-card border border-border rounded-md overflow-hidden shadow-sm transition-[transform,box-shadow] duration-150 ease-out hover:-translate-y-0.5 hover:shadow-[var(--sombra)]',
			destaque && 'border-ouro-claro'
		)}
	>
		{#if produto.imagem}
			<a href={produto.link || '#'} target="_blank" rel="noopener" class="block no-underline hover:opacity-90">
				<img
					src={produto.imagem}
					alt={produto.nome}
					class="w-full h-[180px] object-cover block bg-porcelana max-sm:h-[140px]"
					loading="lazy"
				/>
			</a>
		{/if}
		<div class="p-4 flex flex-col gap-3">
			{#if posicao != null}
				<span class="text-xs font-bold text-tinta-suave opacity-60">#{posicao}</span>
			{/if}
			<header>
				<h3 class="text-base font-semibold leading-tight m-0 line-clamp-2">{produto.nome}</h3>
				<div class="flex items-center gap-1 gap-y-1 mt-1 flex-wrap">
					{#if loja}<span
							class="text-xs font-semibold text-tinta-suave max-w-[140px] overflow-hidden text-ellipsis whitespace-nowrap"
							>🏪 {loja}</span
						>{/if}
					{#if produto.origem}
						<span class="text-[0.65rem] font-bold px-1.5 py-px rounded-full bg-sucesso-fundo text-sucesso"
							>{#if produto.origem === 'Coreia'}🇰🇷{:else if produto.origem === 'Japão'}🇯🇵{:else if produto.origem === 'China'}🇨🇳{/if}
							{produto.origem}</span
						>
					{/if}
					{#if produto.desconto > 0 && produto.desconto <= 1}
						<span class="text-[0.65rem] font-bold px-1.5 py-px rounded-full bg-erro-fundo text-erro"
							>🔥 {Math.round(produto.desconto * 100)}% OFF</span
						>
					{:else if produto.desconto > 1 && produto.desconto <= 100}
						<span class="text-[0.65rem] font-bold px-1.5 py-px rounded-full bg-erro-fundo text-erro"
							>🔥 {Math.round(produto.desconto)}% OFF</span
						>
					{/if}
					{#if produto.oferta_expira}
						<span
							class="text-[0.65rem] font-bold px-1.5 py-px rounded-full bg-ouro-fundo text-ouro-escuro"
							title="Expira em {new Date(produto.oferta_expira).toLocaleDateString('pt-BR')}"
							>⏳ {tempoRestante(produto.oferta_expira)}</span
						>
					{/if}
					{#if produto.categoria}<span class="text-xs font-semibold text-rosa lowercase">{produto.categoria}</span>{/if}
					{#if produto.suspeito}<span
							class="text-[0.65rem] font-bold px-1.5 py-px rounded-full bg-erro-fundo text-erro"
							title="Comissão alta com poucas vendas — pode ser produto sem tração real. Avalie antes de publicar."
							>⚠ suspeito</span
						>{/if}
				</div>
			</header>
			<div class="flex justify-between items-baseline max-sm:flex-col max-sm:gap-1">
				<div class="flex items-baseline gap-2">
					<span class="text-lg font-bold font-mono">{brl(produto.preco)}</span>
					<span class="text-sm font-bold text-ouro">{pct(produto.comissao)}</span>
				</div>
				<div class="flex gap-3 text-xs text-tinta-suave">
					<span>{produto.vendas?.toLocaleString('pt-BR') ?? 0} vendas</span>
					<span>★ {produto.avaliacao?.toLocaleString('pt-BR', { minimumFractionDigits: 1 }) ?? '—'}</span>
				</div>
			</div>
			{#if exibirScore && produto.score}
				<ScoreMeter score={produto.score} componentes={produto.componentes} animar={destaque} />
			{/if}
			<footer class="flex gap-2 mt-2">
				{#if onpublicar}
					<button
						class="rounded-sm py-2 px-3.5 text-sm font-semibold border border-ouro-claro bg-ouro-fundo text-ouro-escuro flex-1 hover:bg-ouro-claro"
						onclick={() => onpublicar(produto)}>📤 Publicar</button
					>
				{/if}
				{#if onfavoritar}
					<button
						class={cn(
							'rounded-sm py-2 px-2.5 text-sm font-semibold border border-transparent bg-transparent text-tinta-suave hover:text-ouro',
							isFav && 'text-ouro'
						)}
						onclick={() => onfavoritar(produto)}
						title={isFav ? 'Remover dos favoritos' : 'Favoritar'}>{isFav ? '★' : '☆'}</button
					>
				{/if}
				{#if exibirLink}
					<button
						class="rounded-sm py-2 px-2.5 text-sm font-semibold border border-transparent bg-transparent text-tinta-suave hover:text-ouro disabled:opacity-30 disabled:cursor-not-allowed"
						onclick={copiarLink}
						disabled={!produto.link}
					>
						{copiado ? '✓ Copiado' : '🔗 Link'}
					</button>
				{/if}
			</footer>
		</div>
	</article>
{:else if layout === 'compact'}
	<!-- ═══ LAYOUT COMPACT — linha horizontal ═══ -->
	<div class="flex gap-3 px-4 py-3 border border-border rounded-sm bg-[var(--branco)] items-center">
		{#if produto.imagem}
			<img src={produto.imagem} alt={produto.nome} class="w-14 h-14 rounded-lg object-cover shrink-0" loading="lazy" />
		{/if}
		<div class="flex-1 min-w-0">
			<h4 class="text-sm m-0 mb-1 whitespace-nowrap overflow-hidden text-ellipsis">{produto.nome}</h4>
			<div class="flex flex-wrap gap-2 text-xs text-tinta-suave">
				<span class="font-bold text-ouro">{brl(produto.preco)}</span>
				<span class="font-semibold">{pct(produto.comissao)}</span>
				{#if produto.vendas}<span>{produto.vendas} vendas</span>{/if}
				{#if produto.avaliacao}<span>★ {produto.avaliacao?.toFixed(1)}</span>{/if}
			</div>
			{#if loja}<span class="text-[0.72rem] text-tinta-suave">🏪 {loja}</span>{/if}
		</div>
		<div class="flex gap-1 shrink-0">
			{#if onfavoritar}
				<button
					class={cn(
						'border border-border bg-porcelana rounded-lg w-9 h-9 flex items-center justify-center cursor-pointer text-base hover:border-ouro',
						isFav && 'text-ouro border-ouro'
					)}
					onclick={() => onfavoritar(produto)}
					title={isFav ? 'Remover dos favoritos' : 'Favoritar'}>{isFav ? '★' : '☆'}</button
				>
			{/if}
			{#if onpublicar}
				<button
					class="border border-border bg-porcelana rounded-lg w-9 h-9 flex items-center justify-center cursor-pointer text-base hover:border-ouro"
					onclick={() => onpublicar(produto)}
					title="Publicar">📤</button
				>
			{/if}
		</div>
	</div>
{:else if layout === 'feed'}
	<!-- ═══ LAYOUT FEED — card com borda lateral (oportunidades) ═══ -->
	<div
		class={cn(
			'border border-border rounded-md p-4 bg-[var(--branco)] transition-[border-color] duration-150 hover:border-ouro-claro',
			variacao?.tipo === 'queda' && 'border-l-[3px] border-l-sucesso',
			variacao?.tipo === 'alta' && 'border-l-[3px] border-l-erro',
			(!variacao?.tipo || variacao?.tipo === 'novo') && 'border-l-[3px] border-l-ouro'
		)}
	>
		<div class="flex items-center gap-2 mb-1.5">
			{#if variacao?.tipo === 'queda'}
				<span class="px-2 py-0.5 rounded-full text-[0.72rem] font-bold bg-sucesso-fundo text-sucesso"
					>↓ {Math.abs(variacao.pct * 100).toFixed(0)}%</span
				>
			{:else if variacao?.tipo === 'alta'}
				<span class="px-2 py-0.5 rounded-full text-[0.72rem] font-bold bg-erro-fundo text-erro"
					>↑ {Math.abs(variacao.pct * 100).toFixed(0)}%</span
				>
			{:else}
				<span class="px-2 py-0.5 rounded-full text-[0.72rem] font-bold bg-ouro-fundo text-ouro-escuro">Novo</span>
			{/if}
			{#if loja}<span class="text-[0.72rem] text-tinta-suave bg-porcelana px-1.5 py-px rounded">{loja}</span>{/if}
			{#if variacao?.detectado_em}<span class="text-[0.72rem] text-tinta-suave ml-auto"
					>{tempoAtras(variacao.detectado_em)}</span
				>{/if}
		</div>

		<div class="flex gap-3">
			{#if produto.imagem}
				<img
					src={produto.imagem}
					alt={produto.nome}
					class="w-16 h-16 rounded-lg object-cover shrink-0 max-sm:w-12 max-sm:h-12"
					loading="lazy"
				/>
			{/if}
			<div class="flex-1 min-w-0">
				<h3 class="text-[0.95rem] font-semibold mb-2 leading-tight line-clamp-2">{produto.nome}</h3>
				<div class="flex items-center gap-2 text-sm flex-wrap">
					{#if variacao?.tipo === 'queda'}
						<span class="line-through text-tinta-suave">{brl(variacao.preco_anterior)}</span>
						<span class="text-tinta-suave text-[0.8rem]">→</span>
						<span class="font-bold text-sucesso">{brl(variacao.preco_atual)}</span>
						<span class="text-xs text-sucesso font-semibold"
							>(-{brl(variacao.preco_anterior - variacao.preco_atual)})</span
						>
					{:else if variacao?.tipo === 'alta'}
						<span class="line-through text-tinta-suave">{brl(variacao.preco_anterior)}</span>
						<span class="text-tinta-suave text-[0.8rem]">→</span>
						<span class="font-bold text-erro">{brl(variacao.preco_atual)}</span>
					{:else}
						<span class="font-bold">{brl(produto.preco)}</span>
						{#if produto.comissao > 0}<span class="text-xs text-ouro font-semibold">{pct(produto.comissao)}</span>{/if}
						{#if produto.vendas > 0}<span class="text-xs text-tinta-suave">{produto.vendas} vendas</span>{/if}
					{/if}
				</div>
			</div>
		</div>

		<div class="mt-2.5 flex gap-2">
			{#if onpublicar}
				<button
					class="py-1.5 px-3.5 border border-ouro-claro bg-ouro-fundo text-ouro-escuro rounded-sm text-[0.82rem] font-semibold cursor-pointer hover:bg-ouro-claro"
					onclick={() => onpublicar(produto)}>📤 Publicar</button
				>
			{/if}
			{#if onfavoritar}
				<button
					class={cn(
						'py-1.5 px-2.5 border border-border bg-transparent rounded-sm cursor-pointer hover:border-ouro',
						isFav && 'text-ouro border-ouro'
					)}
					onclick={() => onfavoritar(produto)}
					title={isFav ? 'Remover dos favoritos' : 'Favoritar'}>{isFav ? '★' : '☆'}</button
				>
			{/if}
		</div>
	</div>
{/if}
