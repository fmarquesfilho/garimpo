<script>
	/**
	 * BuscaUnificada — View da página Descobrir em raias (v3). Renderiza estado da
	 * BuscaEngine e despacha events. Zero lógica de negócio — tudo na engine.
	 *
	 * Layout v3 (separação conceitual):
	 *   0. Console superior: keyword + botões de grupo (Filtros/Lojas) + buscas salvas inline
	 *   1. Filtros  (fontes + quantitativos + categorias + marketplaces)
	 *   2. Lojas    (autocomplete de monitoradas + resolver nova via link/ID)
	 *
	 * Buscas salvas NÃO são uma raia irmã — são atalhos para configurações,
	 * exibidos num painel colapsável dentro do console (BuscasSalvasPanel).
	 */
	import { onMount } from 'svelte';
	import { buscasSalvas } from '$lib/buscas.js';
	import { favoritos } from '$lib/favoritos.js';
	import { BuscaEngine } from '$lib/busca-engine.svelte.js';
	import { criarEffects } from '$lib/busca-engine-effects.js';
	import { DEFAULTS } from '$lib/busca-config.js';
	import Lane from './Lane.svelte';
	import CategoriaCard from './CategoriaCard.svelte';
	import LojaCard from './LojaCard.svelte';
	import MarketplaceFilter from './MarketplaceFilter.svelte';
	import BuscasSalvasPanel from './BuscasSalvasPanel.svelte';
	import { Button, Input, Select, Combobox } from '$lib/components/ui';

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

	// Estado de UI das raias (aberto/fechado) — puramente visual, mora na view
	let lanes = $state({ filtros: false, lojas: false });
	function toggleLane(nome) {
		lanes[nome] = !lanes[nome];
	}
	function colapsarTudo() {
		const algumAberto = lanes.filtros || lanes.lojas || engine.buscasPainelAberto;
		const v = !algumAberto;
		lanes = { filtros: v, lojas: v };
		engine.buscasPainelAberto = v;
	}

	// ── Fontes: os toggles Novos/Quedas/Favoritos controlam essas 3 fontes.
	// Curadoria fica implícita; a fonte "lojas" deriva de haver lojas no escopo.
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

	// Combobox: itens ainda não selecionados
	let categoriasDisponiveis = $derived(
		engine.ctx.categoriasDisponiveis.filter((c) => !engine.ctx.categorias.includes(c.nome ?? c))
	);
	let lojasDisponiveis = $derived(engine.ctx.lojasDisponiveis.filter((l) => !engine.ctx.shopIds.includes(l.id)));
	// "parece URL/ID" → oferece resolver loja nova
	const pareceLoja = (t) => /[./]/.test(t) || /^\d{3,}$/.test(t.trim()) || t.trim().length > 2;

	function limparFiltros() {
		for (const nome of [...engine.ctx.categorias]) engine.send({ type: 'REMOVER_CATEGORIA', nome });
		engine.send({ type: 'MUDAR_FILTRO', comissaoMin: DEFAULTS.comissaoMin, vendasMin: 0 });
		engine.send({ type: 'MUDAR_FONTES', fontes: { ...DEFAULTS.fontes } });
		engine.send({ type: 'MUDAR_MARKETPLACES', marketplaces: [] });
	}
	function limparLojas() {
		for (const id of [...engine.ctx.shopIds]) engine.send({ type: 'REMOVER_LOJA', shopId: id });
	}
</script>

<div class="mb-4 space-y-3">
	<!-- ══ Console superior ══ -->
	<div class="space-y-2 rounded-md border border-border bg-card p-3.5 shadow-sm">
		<div class="flex flex-wrap items-center gap-2">
			<div class="relative min-w-[240px] flex-1">
				<span class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 opacity-50">🔍</span>
				<input
					type="search"
					value={engine.ctx.keyword}
					oninput={(e) => engine.send({ type: 'DIGITAR', value: e.currentTarget.value })}
					placeholder="Buscar produto… (ex: sérum, perfume, batom)"
					class="w-full rounded-sm border border-input bg-background py-2.5 pl-9 pr-4 text-base text-foreground placeholder:text-muted-foreground focus-visible:border-ring focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/20"
					onkeydown={(e) => {
						if (e.key === 'Escape') {
							engine.send({ type: 'DIGITAR', value: '' });
							e.currentTarget.blur();
						} else if (e.key === 'Enter') e.currentTarget.blur();
					}}
				/>
			</div>

			<Button variant={lanes.filtros ? 'primary' : 'secondary'} size="sm" onclick={() => toggleLane('filtros')}>
				⚙️ Filtros{#if engine.contadorFiltros > 0}<span
						class="ml-1.5 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-1 font-[var(--mono)] text-[0.7rem] font-bold text-primary-foreground"
						>{engine.contadorFiltros}</span
					>{/if}
			</Button>
			<Button variant={lanes.lojas ? 'primary' : 'secondary'} size="sm" onclick={() => toggleLane('lojas')}>
				🏪 Lojas{#if engine.contadorLojas > 0}<span
						class="ml-1.5 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-1 font-[var(--mono)] text-[0.7rem] font-bold text-primary-foreground"
						>{engine.contadorLojas}</span
					>{/if}
			</Button>

			<!-- Buscas salvas: botão separado (não é um filtro) -->
			<span class="mx-0.5 hidden self-stretch border-l border-border sm:block"></span>
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

			<span class="flex-1"></span>

			<!-- Indicador de modo (sutil) -->
			{#if engine.modo === 'vinculada'}
				<span
					class="rounded-full border border-primary/40 bg-[var(--ouro-fundo)] px-2 py-0.5 text-[0.68rem] font-semibold text-[var(--ouro-escuro)]"
				>
					↻ busca ativa
				</span>
			{:else if engine.modo === 'editando'}
				<span
					class="rounded-full border border-primary bg-[var(--ouro-fundo)] px-2 py-0.5 text-[0.68rem] font-bold text-[var(--ouro-escuro)]"
				>
					✎ editando
				</span>
			{/if}

			<button
				type="button"
				class="rounded-md px-2 py-1.5 text-xs text-muted-foreground hover:bg-accent hover:text-foreground"
				onclick={colapsarTudo}>⇅ colapsar tudo</button
			>
			<button
				type="button"
				class="rounded-md px-2 py-1.5 text-xs text-muted-foreground hover:bg-[var(--rosa-fundo)] hover:text-destructive"
				onclick={() => engine.send({ type: 'LIMPAR' })}>✕ limpar tudo</button
			>
		</div>

		<!-- Painel de buscas salvas (inline, colapsável) -->
		<BuscasSalvasPanel {engine} bind:open={engine.buscasPainelAberto} />
	</div>

	<!-- ══ Raia 1 — Filtros ══ -->
	<Lane
		title="Filtros"
		tag="fontes · qtd · categorias · marketplaces"
		count={engine.contadorFiltros ? `${engine.contadorFiltros} aplicados` : 'vazio'}
		bind:open={lanes.filtros}
	>
		{#snippet actions()}
			<button
				type="button"
				class="rounded px-2 py-1 text-xs text-muted-foreground hover:bg-[var(--rosa-fundo)] hover:text-destructive"
				onclick={limparFiltros}>limpar raia</button
			>
		{/snippet}

		<!-- sub-raia: fontes + quantitativos -->
		<div class="flex flex-wrap items-end gap-3 rounded-sm border border-border bg-background p-3">
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
		</div>

		<!-- sub-raia: marketplaces -->
		<div class="mt-2.5 rounded-sm border border-border bg-background p-3">
			<MarketplaceFilter
				marketplaces={engine.ctx.marketplacesFiltro}
				onchange={(mkts) => engine.send({ type: 'MUDAR_MARKETPLACES', marketplaces: mkts })}
			/>
		</div>

		<!-- sub-raia: categorias -->
		<div class="mt-2.5 rounded-sm border border-border bg-background p-3">
			<span class="mb-1.5 block font-[var(--mono)] text-[0.6rem] uppercase tracking-wider text-muted-foreground"
				>Categorias (1º nível, por marketplace)</span
			>
			<Combobox
				items={categoriasDisponiveis}
				placeholder="Adicionar categoria… (ex: beleza, casa, eletrônicos)"
				onselect={(c) => engine.send({ type: 'ADICIONAR_CATEGORIA', nome: c.nome, categoria: c })}
			>
				{#snippet option(c)}
					<span class="font-semibold">{c.nome}</span>
					<span class="flex flex-wrap gap-1">
						{#each c.marketplaces ?? [] as m (m)}<span
								class="rounded border border-border bg-muted px-1.5 py-px font-[var(--mono)] text-[0.68rem] text-muted-foreground"
								>{m}</span
							>{/each}
					</span>
				{/snippet}
			</Combobox>
			{#if engine.categoriaCards.length}
				<div class="mt-3 flex flex-wrap gap-2.5">
					{#each engine.categoriaCards as c (c.nome)}
						<CategoriaCard
							nome={c.nome}
							marketplaces={c.marketplaces}
							onremover={() => engine.send({ type: 'REMOVER_CATEGORIA', nome: c.nome })}
						/>
					{/each}
				</div>
			{/if}
		</div>
	</Lane>

	<!-- ══ Raia 2 — Lojas ══ -->
	<Lane
		title="Lojas"
		tag="escopo da busca"
		count={engine.contadorLojas ? `${engine.contadorLojas} no escopo` : 'todas as lojas'}
		bind:open={lanes.lojas}
	>
		{#snippet actions()}
			<button
				type="button"
				class="rounded px-2 py-1 text-xs text-muted-foreground hover:bg-[var(--rosa-fundo)] hover:text-destructive"
				onclick={limparLojas}>limpar raia</button
			>
		{/snippet}

		<Combobox
			items={lojasDisponiveis}
			placeholder="Buscar loja monitorada, ou colar link/ID para adicionar uma nova…"
			allowFree={true}
			isFree={pareceLoja}
			freeLabel={(t) => `↳ resolver e adicionar loja "${t}"`}
			onselect={(l) => engine.send({ type: 'ADICIONAR_LOJA', loja: l })}
			onfree={(t) => engine.send({ type: 'ADICIONAR_LOJA', value: t })}
		>
			{#snippet option(l)}
				<span class="font-semibold">{l.origem ? l.origem + ' ' : ''}{l.nome}</span>
				<span
					class="rounded border border-border bg-muted px-1.5 py-px font-[var(--mono)] text-[0.68rem] text-muted-foreground"
					>{l.marketplace}</span
				>
			{/snippet}
		</Combobox>
		{#if engine.ctx.lojaResolvendo}<span class="mt-1 block text-xs text-muted-foreground">resolvendo…</span>{/if}
		{#if engine.ctx.lojaErro}<span class="mt-1 block text-xs text-destructive">{engine.ctx.lojaErro}</span>{/if}

		{#if engine.lojaCards.length}
			<div class="mt-3 flex flex-wrap gap-2.5">
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
			</div>
		{:else}
			<p class="mt-2 text-sm italic text-muted-foreground">
				Nenhuma loja no escopo — a busca roda em todas as lojas monitoradas.
			</p>
		{/if}
	</Lane>
</div>
