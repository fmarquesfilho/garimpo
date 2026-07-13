<script>
	/**
	 * BuscaUnificada — View da página Descobrir (v4, omnibox). Renderiza estado da
	 * BuscaEngine e despacha events. Zero lógica de negócio — tudo na engine.
	 *
	 * Layout v4 (omnibox substitui as raias):
	 *   0. Omnibox: input unificado (keyword + @loja + #categoria + !marketplace)
	 *   1. Filtros numéricos: fontes · comissão · vendas · marketplaces (controles diretos)
	 *   2. Escopo ativo: cards de loja/categoria selecionadas (removíveis)
	 *   3. Buscas salvas: painel colapsável (atalhos para configurações)
	 *
	 * As lanes (Lane.svelte) e os autocompletes separados de loja/categoria foram
	 * removidos — o Omnibox concentra a composição da busca.
	 */
	import { onMount } from 'svelte';
	import { buscasSalvas } from '$lib/buscas.js';
	import { favoritos } from '$lib/favoritos.js';
	import { BuscaEngine } from '$lib/busca-engine.svelte.js';
	import { criarEffects } from '$lib/busca-engine-effects.js';
	import Omnibox from './Omnibox.svelte';
	import StoreCard from './StoreCard.svelte';
	import CategoriaCard from './CategoriaCard.svelte';
	import LojaCard from './LojaCard.svelte';
	import MarketplaceFilter from './MarketplaceFilter.svelte';
	import BuscasSalvasPanel from './BuscasSalvasPanel.svelte';
	import { Input, Select } from '$lib/components/ui';

	let { onresultados = null, oncarregando = null, onerro = null } = $props();

	const effects = criarEffects({
		getBuscasSalvas: () => $buscasSalvas,
		getFavoritos: () => $favoritos,
		sincronizarStore: () => buscasSalvas.sincronizarDoServidor()
	});
	const engine = new BuscaEngine(effects);

	onMount(() => engine.send({ type: 'INICIALIZAR' }));

	// Propaga estado para o parent
	$effect(() => onresultados?.(engine.ctx.resultados));
	$effect(() => oncarregando?.(engine.loading));
	$effect(() => onerro?.(engine.ctx.error ? new Error(engine.ctx.error) : null));

	// ── Fontes: toggles Novos/Quedas/Favoritos ────────────────────────────────
	const TOGGLES = [
		{ key: 'novos', label: '🆕 Novos' },
		{ key: 'quedas', label: '📉 Quedas' },
		{ key: 'favoritos', label: '⭐ Favoritos' }
	];
	function alternarFonte(key) {
		const f = { ...engine.ctx.fontes, [key]: !engine.ctx.fontes[key] };
		engine.send({ type: 'MUDAR_FONTES', fontes: f });
	}

	const comissaoOpcoes = [0.05, 0.07, 0.1, 0.15].map((c) => ({ value: String(c), label: `${Math.round(c * 100)}%` }));

	function limparTudo() {
		engine.send({ type: 'LIMPAR' });
	}
</script>

<div class="mb-4 space-y-3">
	<!-- ══ Console: omnibox + filtros ══ -->
	<div class="space-y-3 rounded-md border border-border bg-card p-3.5 shadow-sm">
		<Omnibox {engine} />

		<!-- Filtros numéricos (fora do input) -->
		<div class="flex flex-wrap items-end gap-3">
			<div class="flex items-center gap-2">
				{#each TOGGLES as t (t.key)}
					{@const on = engine.ctx.fontes[t.key]}
					<button
						type="button"
						class="rounded-full border px-3 py-1.5 text-sm transition-colors {on
							? 'border-primary bg-[var(--ouro-fundo)] font-semibold text-[var(--ouro-escuro)]'
							: 'border-border bg-card text-muted-foreground hover:border-primary'}"
						onclick={() => alternarFonte(t.key)}>{t.label}</button
					>
				{/each}
			</div>
			<div class="mx-1 hidden self-stretch border-l border-border sm:block"></div>
			<Select
				label="comissão mín."
				value={String(engine.ctx.comissaoMin)}
				onchange={(v) => engine.send({ type: 'MUDAR_FILTRO', comissaoMin: Number(v) })}
				options={comissaoOpcoes}
				size="sm"
				class="w-24"
			/>
			<div class="flex flex-col gap-1">
				<span class="text-xs font-semibold text-muted-foreground">vendas mín.</span>
				<Input
					type="number"
					min="0"
					value={String(engine.ctx.vendasMin)}
					oninput={(e) => engine.send({ type: 'MUDAR_FILTRO', vendasMin: Number(e.currentTarget.value || 0) })}
					size="sm"
					class="w-20"
				/>
			</div>
			<span class="flex-1"></span>

			<!-- Indicador de modo (sutil) -->
			{#if engine.modo === 'vinculada'}
				<span
					class="rounded-full border border-primary/40 bg-[var(--ouro-fundo)] px-2 py-0.5 text-[0.68rem] font-semibold text-[var(--ouro-escuro)]"
					>↻ busca ativa</span
				>
			{:else if engine.modo === 'editando'}
				<span
					class="rounded-full border border-primary bg-[var(--ouro-fundo)] px-2 py-0.5 text-[0.68rem] font-bold text-[var(--ouro-escuro)]"
					>✎ editando</span
				>
			{/if}

			<button
				type="button"
				class="rounded-md px-2 py-1.5 text-xs text-muted-foreground hover:bg-[var(--rosa-fundo)] hover:text-destructive"
				onclick={limparTudo}>✕ limpar tudo</button
			>
		</div>

		<!-- Marketplaces (filtro de escopo) -->
		<div class="rounded-sm border border-border bg-background p-3">
			<MarketplaceFilter
				marketplaces={engine.ctx.marketplacesFiltro}
				onchange={(mkts) => engine.send({ type: 'MUDAR_MARKETPLACES', marketplaces: mkts })}
			/>
		</div>

		<!-- Buscas salvas (colapsável) -->
		<div class="flex items-center gap-2">
			<button
				type="button"
				class="flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm transition-colors
					{engine.buscasPainelAberto
					? 'border border-primary bg-[var(--ouro-fundo)] font-semibold text-[var(--ouro-escuro)]'
					: 'border border-border text-muted-foreground hover:border-primary hover:text-foreground'}"
				onclick={() => (engine.buscasPainelAberto = !engine.buscasPainelAberto)}
			>
				💾 {engine.contadorBuscas}
				{engine.contadorBuscas === 1 ? 'salva' : 'salvas'}
				<span class="text-xs leading-none">{engine.buscasPainelAberto ? '▴' : '▾'}</span>
			</button>
		</div>
		<BuscasSalvasPanel {engine} bind:open={engine.buscasPainelAberto} />
	</div>

	<!-- ══ Escopo ativo: lojas + categorias selecionadas ══ -->
	{#if engine.lojaCards.length || engine.categoriaCards.length}
		<div class="flex flex-wrap gap-2.5">
			{#each engine.lojaCards as l (l.id)}
				<LojaCard
					nome={l.nome}
					marketplace={l.marketplace}
					origem={l.origem}
					monitorada={l.monitorada}
					cron={l.cron}
					onremover={() => engine.send({ type: 'REMOVER_LOJA', shopId: l.id })}
				/>
			{/each}
			{#each engine.categoriaCards as c (c.nome)}
				<CategoriaCard
					nome={c.nome}
					marketplaces={c.marketplaces}
					onremover={() => engine.send({ type: 'REMOVER_CATEGORIA', nome: c.nome })}
				/>
			{/each}
		</div>
	{/if}

	{#if engine.ctx.lojaResolvendo}<span class="block text-xs text-muted-foreground">resolvendo loja…</span>{/if}
	{#if engine.ctx.lojaErro}<span class="block text-xs text-destructive">{engine.ctx.lojaErro}</span>{/if}

	<!-- ══ Resultados: Store Cards (modo lojas) ══ -->
	{#if engine.modoResultados === 'lojas'}
		<div class="space-y-2">
			{#if engine.resultadosLojas.length > 0}
				{#each engine.resultadosLojas as loja (loja.id)}
					<StoreCard {loja} {engine} />
				{/each}
			{:else if engine.status === 'results'}
				<div class="rounded-md border border-border bg-card p-4 text-center text-sm text-muted-foreground">
					<p>Nenhuma loja encontrada com esse termo.</p>
					<p class="mt-1 text-xs">Tente colar o link da loja para adicioná-la ao registro.</p>
				</div>
			{/if}
		</div>
	{/if}
</div>
