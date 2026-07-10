<script>
	/**
	 * BuscasSalvasPanel — Painel colapsável de buscas salvas/agendadas.
	 *
	 * Conceitualmente separado das raias de Filtros e Lojas: buscas salvas
	 * são atalhos para configurações, não filtros em si. Fica inline no
	 * console superior, colapsável via botão.
	 *
	 * @prop engine — instância da BuscaEngine (lê ctx, despacha events)
	 * @prop open — bindable, estado aberto/fechado do painel
	 */
	import { MODOS } from '$lib/busca-engine-state.js';
	import { gerarLabelBusca, cronLabel } from '$lib/busca-engine.svelte.js';
	import AgendadorBusca from './AgendadorBusca.svelte';
	import BuscaCard from './BuscaCard.svelte';
	import { Button } from '$lib/components/ui';

	let { engine, open = $bindable(false) } = $props();
</script>

{#if open}
	<div class="mt-2 rounded-md border border-border bg-card shadow-sm">
		<!-- Cabeçalho do painel -->
		<div class="flex items-center gap-2 border-b border-border bg-muted px-3.5 py-2">
			<span class="font-[var(--display)] text-[0.95rem] font-bold text-foreground">
				💾 Buscas salvas
			</span>
			<span class="font-[var(--mono)] text-xs text-muted-foreground">
				{engine.contadorBuscas} {engine.contadorBuscas === 1 ? 'salva' : 'salvas'}
			</span>
			<span class="flex-1"></span>
			<button
				type="button"
				class="rounded px-2 py-1 text-xs text-muted-foreground hover:bg-accent hover:text-primary"
				onclick={() => (engine.salvarAberto = !engine.salvarAberto)}
			>＋ salvar busca atual</button>
			<button
				type="button"
				class="rounded px-2 py-1 text-xs text-muted-foreground hover:bg-accent hover:text-foreground"
				onclick={() => (open = false)}
			>✕</button>
		</div>

		<div class="p-3.5">
			<!-- Formulário de salvar/editar -->
			{#if engine.salvarAberto}
				<div class="mb-3 rounded-sm border border-border bg-background p-3">
					<p class="mb-2 text-sm font-semibold text-foreground">
						💾 {engine.ctx.editandoId ? 'Editar busca' : 'Salvar configuração atual'}
					</p>
					<AgendadorBusca bind:value={engine.ctx.cron} />

					{#if engine.ctx.erroDuplicata}
						<div class="mt-2 rounded-sm border border-[var(--aviso-borda)] bg-[var(--aviso-fundo)] px-3 py-2 text-sm text-[var(--aviso-texto)]">
							⚠️ {engine.ctx.erroDuplicata}
						</div>
					{/if}

					<div class="mt-2 flex justify-end gap-2">
						{#if engine.ctx.editandoId}
							<Button variant="ghost" size="sm" onclick={() => engine.send({ type: 'CANCELAR_EDICAO' })}>Cancelar edição</Button>
						{:else}
							<Button variant="ghost" size="sm" onclick={() => (engine.salvarAberto = false)}>Cancelar</Button>
						{/if}
						<Button size="sm" onclick={() => engine.send({ type: 'SALVAR' })}>
							{engine.ctx.editandoId ? 'Salvar alterações' : 'Salvar'}{engine.ctx.cron ? ' + agendar' : ''}
						</Button>
					</div>
				</div>
			{/if}

			<!-- Feedback reativo de duplicata -->
			{#if engine.buscaDuplicada && !engine.salvarAberto}
				<div class="mb-3 rounded-sm border border-[var(--aviso-borda)] bg-[var(--aviso-fundo)] px-3 py-2 text-sm text-[var(--aviso-texto)]">
					💡 Esta configuração já existe como busca salva: <strong>"{gerarLabelBusca(engine.buscaDuplicada)}"</strong>
				</div>
			{/if}

			<!-- Indicador de modo -->
			{#if engine.modo === 'vinculada' && engine.ctx.buscaSelecionadaId}
				<div class="mb-3 flex items-center gap-2 text-xs text-muted-foreground">
					<span class="rounded-full border border-primary bg-[var(--ouro-fundo)] px-2 py-0.5 font-semibold text-[var(--ouro-escuro)]">
						↻ rodando busca salva
					</span>
				</div>
			{:else if engine.modo === 'editando'}
				<div class="mb-3 flex items-center gap-2 text-xs">
					<span class="rounded-full border border-primary bg-[var(--ouro-fundo)] px-2 py-0.5 font-bold text-[var(--ouro-escuro)]">
						✎ editando busca salva
					</span>
				</div>
			{/if}

			<!-- Lista de buscas salvas -->
			{#if engine.ctx.buscasSalvas.length}
				<div class="flex flex-wrap gap-2.5">
					{#each engine.ctx.buscasSalvas as b (b.id)}
						<BuscaCard
							busca={b}
							editando={engine.ctx.editandoId === b.id}
							selecionada={engine.ctx.buscaSelecionadaId === b.id}
							onrodar={(c) => engine.send({ type: 'CARREGAR_SALVA', config: c })}
							oneditar={(c) => engine.send({ type: 'EDITAR_SALVA', config: c })}
							onremover={(c) => engine.send({ type: 'REMOVER_SALVA', config: c })}
						/>
					{/each}
				</div>
			{:else}
				<p class="text-sm italic text-muted-foreground">
					Nenhuma busca salva ainda. Configure filtros/lojas e clique em "＋ salvar busca atual".
				</p>
			{/if}
		</div>
	</div>
{/if}
