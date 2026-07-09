<script>
	/**
	 * BuscaUnificada — View da página Descobrir em raias. Renderiza estado da
	 * BuscaEngine e despacha events. Zero lógica de negócio — tudo na engine.
	 *
	 * Layout em raias (piscina):
	 *   0. Console superior: keyword + botões de grupo (contadores) + colapsar/limpar tudo
	 *   1. Filtros  (fontes + quantitativos em cima, categorias embaixo)
	 *   2. Lojas    (autocomplete de monitoradas + resolver nova via link/ID)
	 *   3. Buscas   (salvas & agendadas, com edit mode)
	 */
	import { onMount } from 'svelte';
	import { buscasSalvas } from '$lib/buscas.js';
	import { favoritos } from '$lib/favoritos.js';
	import { BuscaEngine } from '$lib/busca-engine.svelte.js';
	import { criarEffects } from '$lib/busca-engine-effects.js';
	import { DEFAULTS } from '$lib/busca-config.js';
	import AgendadorBusca from './AgendadorBusca.svelte';
	import Lane from './Lane.svelte';
	import CategoriaCard from './CategoriaCard.svelte';
	import LojaCard from './LojaCard.svelte';
	import BuscaCard from './BuscaCard.svelte';
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
	let lanes = $state({ filtros: false, lojas: false, buscas: false });
	function toggleLane(nome) {
		lanes[nome] = !lanes[nome];
	}
	function colapsarTudo() {
		const algumAberto = lanes.filtros || lanes.lojas || lanes.buscas;
		const v = !algumAberto;
		lanes = { filtros: v, lojas: v, buscas: v };
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
			<Button variant={lanes.buscas ? 'primary' : 'secondary'} size="sm" onclick={() => toggleLane('buscas')}>
				💾 Buscas{#if engine.contadorBuscas > 0}<span
						class="ml-1.5 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-1 font-[var(--mono)] text-[0.7rem] font-bold text-primary-foreground"
						>{engine.contadorBuscas}</span
					>{/if}
			</Button>

			<span class="flex-1"></span>
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
	</div>

	<!-- ══ Raia 1 — Filtros ══ -->
	<Lane
		title="Filtros"
		tag="fontes · qtd · categorias"
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
			freeLabel={(t) => `↳ resolver e adicionar loja “${t}”`}
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

	<!-- ══ Raia 3 — Buscas ══ -->
	<Lane title="Buscas" tag="salvas & agendadas" count={`${engine.contadorBuscas} salvas`} bind:open={lanes.buscas}>
		{#snippet actions()}
			<button
				type="button"
				class="rounded px-2 py-1 text-xs text-muted-foreground hover:bg-accent hover:text-primary"
				onclick={() => (engine.salvarAberto = !engine.salvarAberto)}>＋ salvar busca atual</button
			>
		{/snippet}

		{#if engine.salvarAberto}
			<div class="mb-3 rounded-sm border border-border bg-background p-3">
				<p class="mb-2 text-sm font-semibold text-foreground">
					💾 {engine.ctx.editandoId ? 'Editar busca' : 'Salvar configuração atual'}
				</p>
				<AgendadorBusca bind:value={engine.ctx.cron} />
				<div class="mt-2 flex justify-end gap-2">
					<Button variant="ghost" size="sm" onclick={() => (engine.salvarAberto = false)}>Cancelar</Button>
					<Button size="sm" onclick={() => engine.send({ type: 'SALVAR' })}
						>{engine.ctx.editandoId ? 'Salvar alterações' : 'Salvar'}{engine.ctx.cron ? ' + agendar' : ''}</Button
					>
				</div>
			</div>
		{/if}

		{#if engine.ctx.buscasSalvas.length}
			<div class="flex flex-wrap gap-2.5">
				{#each engine.ctx.buscasSalvas as b (b.id)}
					<BuscaCard
						busca={b}
						editando={engine.ctx.editandoId === b.id}
						onrodar={(c) => engine.send({ type: 'CARREGAR_SALVA', config: c })}
						oneditar={(c) => engine.send({ type: 'EDITAR_SALVA', config: c })}
						onremover={(c) => engine.send({ type: 'REMOVER_SALVA', config: c })}
					/>
				{/each}
			</div>
		{:else}
			<p class="text-sm italic text-muted-foreground">
				Nenhuma busca salva ainda. Configure filtros/lojas e clique em “＋ salvar busca atual”.
			</p>
		{/if}
	</Lane>
</div>
