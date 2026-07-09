<script>
	/**
	 * BuscaUnificada — View pura. Renderiza estado da BuscaEngine e despacha events.
	 * Zero lógica de negócio aqui — tudo na engine.
	 */
	import { onMount } from 'svelte';
	import { buscasSalvas } from '$lib/buscas.js';
	import { favoritos } from '$lib/favoritos.js';
	import { BuscaEngine, gerarLabelBusca, cronLabel } from '$lib/busca-engine.svelte.js';
	import { criarEffects } from '$lib/busca-engine-effects.js';
	import { buildFonteOpcoes } from '$lib/descobrir-logic.js';
	import TagInput from './TagInput.svelte';
	import AgendadorBusca from './AgendadorBusca.svelte';
	import { Button, Input, Badge, Select, ToggleGroup } from '$lib/components/ui';

	let { onresultados = null, oncarregando = null, onerro = null } = $props();

	const effects = criarEffects({ getBuscasSalvas: () => $buscasSalvas, getFavoritos: () => $favoritos });
	const engine = new BuscaEngine(effects);

	onMount(() => engine.send({ type: 'INICIALIZAR' }));

	// Propaga estado para o parent via callbacks
	$effect(() => {
		onresultados?.(engine.ctx.resultados);
	});
	$effect(() => {
		oncarregando?.(engine.loading);
	});
	$effect(() => {
		onerro?.(engine.ctx.error ? new Error(engine.ctx.error) : null);
	});

	// Derivados UI
	let fonteOpcoes = $derived(
		buildFonteOpcoes({
			contagemCuradoria: engine.ctx.contagens.curadoria,
			contagemQuedas: engine.ctx.contagens.quedas,
			contagemNovos: engine.ctx.contagens.novos,
			contagemLojas: engine.ctx.contagens.lojas,
			totalFavoritos: $favoritos.length
		})
	);

	const comissaoOpcoes = [0.05, 0.07, 0.1, 0.15].map((c) => ({ value: String(c), label: `${Math.round(c * 100)}%` }));
</script>

{#if engine.colapsado}
	<button
		class="mb-4 flex w-full items-center justify-between rounded-md border border-border bg-card px-4 py-3 text-sm text-muted-foreground hover:border-primary hover:text-foreground"
		onclick={() => (engine.colapsado = false)}
		type="button"
	>
		<span>🔍 {engine.resumo}</span>
		<span class="text-xs">▼ abrir</span>
	</button>
{:else}
	<div class="mb-4 space-y-3 rounded-md border border-border bg-card p-4">
		<!-- Keywords + ações -->
		<div class="flex gap-2">
			<div class="relative flex-1">
				<input
					type="search"
					value={engine.ctx.keyword}
					oninput={(e) => engine.send({ type: 'DIGITAR', value: /** @type {HTMLInputElement} */ (e.target).value })}
					placeholder="🔍 Buscar produto… (ex: sérum, perfume, batom)"
					class="w-full rounded-md border border-border bg-background px-4 py-3 text-base text-foreground placeholder:text-muted-foreground placeholder:opacity-60 focus:border-primary focus:outline-none focus:ring-2 focus:ring-ring/20"
					onkeydown={(e) => {
						if (e.key === 'Escape') {
							engine.send({ type: 'LIMPAR' });
							/** @type {HTMLInputElement} */ (e.target).blur();
						} else if (e.key === 'Enter') /** @type {HTMLInputElement} */ (e.target).blur();
					}}
				/>
				{#if engine.ctx.keyword}
					<Button
						variant="ghost"
						size="icon"
						class="absolute right-2 top-1/2 h-6 w-6 -translate-y-1/2"
						onclick={() => engine.send({ type: 'DIGITAR', value: '' })}
						aria-label="Limpar">✕</Button
					>
				{/if}
			</div>
			<Button variant="secondary" size="sm" onclick={() => (engine.filtrosAberto = !engine.filtrosAberto)}>
				⚙️ Filtros {#if engine.filtrosAtivos > 0 && !engine.filtrosAberto}<Badge>{engine.filtrosAtivos}</Badge>{/if}
			</Button>
			<Button variant="secondary" size="sm" onclick={() => (engine.salvarAberto = !engine.salvarAberto)}>💾</Button>
			<Button variant="ghost" size="sm" onclick={() => (engine.colapsado = true)} aria-label="Colapsar">▲</Button>
		</div>

		<!-- Lojas selecionadas -->
		<div class="flex flex-wrap items-center gap-2">
			{#each engine.ctx.shopIds as id (id)}
				<Badge variant="secondary">
					🏪 {engine.ctx.shopNomes[id] || id}
					<button class="ml-1 text-xs" onclick={() => engine.send({ type: 'REMOVER_LOJA', shopId: id })} type="button"
						>✕</button
					>
				</Badge>
			{/each}
			<div class="flex items-center gap-1">
				<Input
					placeholder={engine.ctx.shopIds.length > 0 ? '+ outra loja' : '🏪 Adicionar loja (URL ou ID) — opcional'}
					size="sm"
					class="w-56"
					onkeydown={(e) => {
						if (e.key === 'Enter') {
							engine.send({ type: 'ADICIONAR_LOJA', value: /** @type {HTMLInputElement} */ (e.target).value });
							/** @type {HTMLInputElement} */ (e.target).value = '';
						}
					}}
				/>
				{#if engine.ctx.lojaResolvendo}<span class="text-xs text-muted-foreground">resolvendo…</span>{/if}
				{#if engine.ctx.lojaErro}<span class="text-xs text-destructive">{engine.ctx.lojaErro}</span>{/if}
			</div>
		</div>

		<!-- Filtros colapsáveis -->
		{#if engine.filtrosAberto}
			<div class="flex flex-wrap items-end gap-3 rounded-sm border border-border bg-muted p-3">
				<div class="flex flex-col gap-1">
					<span class="text-xs font-semibold text-muted-foreground">comissão mín.</span>
					<Select
						value={String(engine.ctx.comissaoMin)}
						onchange={(v) => engine.send({ type: 'MUDAR_FILTRO', comissaoMin: Number(v) })}
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
						value={String(engine.ctx.vendasMin)}
						oninput={(e) =>
							engine.send({
								type: 'MUDAR_FILTRO',
								vendasMin: Number(/** @type {HTMLInputElement} */ (e.target).value || 0)
							})}
						size="sm"
						class="w-20"
					/>
				</div>
				<div class="flex-1">
					<TagInput bind:tags={engine.ctx.categorias} label="Categorias" placeholder="ex: cosméticos, perfumaria" />
				</div>
			</div>
		{/if}

		<!-- Fontes -->
		<ToggleGroup
			type="multiple"
			value={engine.fontesAtivas}
			options={fonteOpcoes}
			onchange={(v) =>
				engine.send({
					type: 'MUDAR_FONTES',
					fontes: {
						curadoria: v.includes('curadoria'),
						quedas: v.includes('quedas'),
						novos: v.includes('novos'),
						lojas: v.includes('lojas'),
						favoritos: v.includes('favoritos')
					}
				})}
		/>

		<!-- Salvar busca -->
		{#if engine.salvarAberto}
			<div class="rounded-sm border border-border bg-muted p-3">
				<p class="mb-2 text-sm font-semibold text-foreground">💾 Salvar configuração atual</p>
				<AgendadorBusca bind:value={engine.ctx.cron} />
				<div class="mt-2 flex justify-end gap-2">
					<Button variant="ghost" size="sm" onclick={() => (engine.salvarAberto = false)}>Cancelar</Button>
					<Button size="sm" onclick={() => engine.send({ type: 'SALVAR' })}
						>Salvar{engine.ctx.cron ? ' + agendar' : ''}</Button
					>
				</div>
			</div>
		{/if}

		<!-- Buscas salvas -->
		{#if engine.ctx.buscasSalvas.length > 0}
			<div class="flex flex-wrap gap-2">
				{#each engine.ctx.buscasSalvas as b (b.id)}
					<div class="flex items-center gap-0.5">
						{#if b.cron}<span class="text-xs text-primary" title={cronLabel(b.cron)}>⏱</span>{/if}
						<button
							class="rounded-full border border-border bg-porcelana px-3 py-1 text-xs font-semibold text-foreground hover:border-primary hover:text-primary"
							onclick={() => engine.send({ type: 'CARREGAR_SALVA', config: b })}
							type="button">{gerarLabelBusca(b)}</button
						>
						<button
							class="text-xs text-muted-foreground hover:text-destructive"
							onclick={() => engine.send({ type: 'REMOVER_SALVA', config: b })}
							type="button">✕</button
						>
					</div>
				{/each}
			</div>
		{/if}
	</div>
{/if}
