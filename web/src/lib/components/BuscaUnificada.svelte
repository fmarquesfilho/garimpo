<script>
	/**
	 * BuscaUnificada — componente integrado de busca, filtros, lojas e agendamento.
	 * Substitui FilterBar + FormAdicionarLoja + GerenciarBuscas.
	 * Lógica em BuscaUnificada.svelte.js (estado + handlers).
	 */
	import { onMount } from 'svelte';
	import { buscasSalvas } from '$lib/buscas.js';
	import { favoritos } from '$lib/favoritos.js';
	import {
		criarEstado,
		criarDerivados,
		criarHandlers,
		comissaoOpcoes,
		gerarLabelBusca,
		cronLabel
	} from './BuscaUnificada.svelte.js';
	import TagInput from './TagInput.svelte';
	import AgendadorBusca from './AgendadorBusca.svelte';
	import { Button, Input, Badge, Select, ToggleGroup } from '$lib/components/ui';

	let { onresultados = null, oncarregando = null, onerro = null } = $props();

	const s = criarEstado();
	const d = criarDerivados(
		s,
		() => $buscasSalvas,
		() => $favoritos
	);
	const h = criarHandlers(s, d, {
		onresultados: (r) => onresultados?.(r),
		oncarregando: (v) => oncarregando?.(v),
		onerro: (e) => onerro?.(e)
	});

	onMount(async () => {
		await h.inicializar();
		initialized = true;
	});

	// Debounce — re-executa busca quando qualquer parâmetro muda (após init)
	let initialized = $state(false);
	let timer;
	$effect(() => {
		void s.busca;
		void s.keywords;
		void s.shopIds;
		void s.comissaoMin;
		void s.vendasMin;
		void s.categorias;
		void s.fontes;
		void d.fontesAtivas;
		if (!initialized) return;
		clearTimeout(timer);
		timer = setTimeout(() => h.executar(), 400);
		return () => clearTimeout(timer);
	});
</script>

{#if s.colapsado}
	<button
		class="mb-4 flex w-full items-center justify-between rounded-md border border-border bg-card px-4 py-3 text-sm text-muted-foreground hover:border-primary hover:text-foreground"
		onclick={() => (s.colapsado = false)}
		type="button"
	>
		<span>🔍 {d.resumo}</span>
		<span class="text-xs">▼ abrir</span>
	</button>
{:else}
	<div class="mb-4 space-y-3 rounded-md border border-border bg-card p-4">
		<!-- Keywords + ações -->
		<div class="flex gap-2">
			<div class="relative flex-1">
				<input
					type="search"
					value={s.busca}
					oninput={(e) => (s.busca = /** @type {HTMLInputElement} */ (e.target).value)}
					placeholder="🔍 Buscar produto… (ex: sérum, perfume, batom)"
					class="w-full rounded-md border border-border bg-background px-4 py-3 text-base text-foreground placeholder:text-muted-foreground placeholder:opacity-60 focus:border-primary focus:outline-none focus:ring-2 focus:ring-ring/20"
					onkeydown={(e) => {
						if (e.key === 'Escape') {
							s.busca = '';
							/** @type {HTMLInputElement} */ (e.target).blur();
						} else if (e.key === 'Enter') /** @type {HTMLInputElement} */ (e.target).blur();
					}}
				/>
				{#if s.busca}
					<Button
						variant="ghost"
						size="icon"
						class="absolute right-2 top-1/2 h-6 w-6 -translate-y-1/2"
						onclick={() => (s.busca = '')}
						aria-label="Limpar">✕</Button
					>
				{/if}
			</div>
			<Button variant="secondary" size="sm" onclick={() => (s.filtrosAberto = !s.filtrosAberto)}>
				⚙️ Filtros {#if d.filtrosAtivos > 0 && !s.filtrosAberto}<Badge>{d.filtrosAtivos}</Badge>{/if}
			</Button>
			<Button variant="secondary" size="sm" onclick={() => (s.salvarAberto = !s.salvarAberto)}>💾</Button>
			<Button variant="ghost" size="sm" onclick={() => (s.colapsado = true)} aria-label="Colapsar">▲</Button>
		</div>

		<!-- Lojas selecionadas -->
		<div class="flex flex-wrap items-center gap-2">
			{#each s.shopIds as id (id)}
				<Badge variant="secondary">
					🏪 {s.shopNomes[id] || id}
					<button class="ml-1 text-xs" onclick={() => h.handleRemoverLoja(id)} type="button">✕</button>
				</Badge>
			{/each}
			<div class="flex items-center gap-1">
				<Input
					value={s.lojaInput}
					oninput={(e) => (s.lojaInput = /** @type {HTMLInputElement} */ (e.target).value)}
					placeholder={s.shopIds.length > 0 ? '+ outra loja' : '🏪 Adicionar loja (URL ou ID) — opcional'}
					size="sm"
					class="w-56"
					onkeydown={(e) => {
						if (e.key === 'Enter') h.handleAdicionarLoja();
					}}
				/>
				{#if s.lojaResolvendo}<span class="text-xs text-muted-foreground">resolvendo…</span>{/if}
				{#if s.lojaErro}<span class="text-xs text-destructive">{s.lojaErro}</span>{/if}
			</div>
		</div>

		<!-- Filtros colapsáveis -->
		{#if s.filtrosAberto}
			<div class="flex flex-wrap items-end gap-3 rounded-sm border border-border bg-muted p-3">
				<div class="flex flex-col gap-1">
					<span class="text-xs font-semibold text-muted-foreground">comissão mín.</span>
					<Select
						value={String(s.comissaoMin)}
						onchange={(v) => (s.comissaoMin = Number(v))}
						options={comissaoOpcoes}
						size="sm"
						class="w-24"
					/>
				</div>
				<div class="flex flex-col gap-1">
					<span class="text-xs font-semibold text-muted-foreground">vendas mín.</span>
					<Input
						type="number"
						min="0"
						value={String(s.vendasMin)}
						oninput={(e) => (s.vendasMin = Number(/** @type {HTMLInputElement} */ (e.target).value || 0))}
						size="sm"
						class="w-20"
					/>
				</div>
				<div class="flex-1">
					<TagInput bind:tags={s.categorias} label="Categorias" placeholder="ex: cosméticos, perfumaria" />
				</div>
			</div>
		{/if}

		<!-- Fontes -->
		<ToggleGroup type="multiple" value={d.fontesAtivas} options={d.fonteOpcoes} onchange={h.handleFontesChange} />

		<!-- Salvar busca -->
		{#if s.salvarAberto}
			<div class="rounded-sm border border-border bg-muted p-3">
				<p class="mb-2 text-sm font-semibold text-foreground">💾 Salvar configuração atual</p>
				<AgendadorBusca bind:value={s.cron} />
				<div class="mt-2 flex justify-end gap-2">
					<Button variant="ghost" size="sm" onclick={() => (s.salvarAberto = false)}>Cancelar</Button>
					<Button size="sm" onclick={h.handleSalvar}>Salvar{s.cron ? ' + agendar' : ''}</Button>
				</div>
			</div>
		{/if}

		<!-- Buscas salvas -->
		{#if s.buscasSalvasLista.length > 0}
			<div class="flex flex-wrap gap-2">
				{#each s.buscasSalvasLista as b (b.id)}
					<div class="flex items-center gap-0.5">
						{#if b.cron}<span class="text-xs text-primary" title={cronLabel(b.cron)}>⏱</span>{/if}
						<button
							class="rounded-full border border-border bg-porcelana px-3 py-1 text-xs font-semibold text-foreground hover:border-primary hover:text-primary"
							onclick={() => h.handleCarregarSalva(b)}
							type="button">{gerarLabelBusca(b)}</button
						>
						<button
							class="text-xs text-muted-foreground hover:text-destructive"
							onclick={() => h.handleRemoverSalva(b)}
							type="button">✕</button
						>
					</div>
				{/each}
			</div>
		{/if}
	</div>
{/if}
