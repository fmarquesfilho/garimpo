<script>
	/**
	 * Gerenciamento de buscas salvas — criar, visualizar, remover.
	 * Inclui agendamento (cron) e múltiplas keywords.
	 */
	import { buscasSalvas, slugificar } from '$lib/buscas.js';
	import TagInput from './TagInput.svelte';
	import AgendadorBusca from './AgendadorBusca.svelte';
	import BuscaCard from './BuscaCard.svelte';

	let buscasKw = $derived(($buscasSalvas ?? []).filter(b => !b.shop_ids?.length));

	let mostrarForm = $state(false);
	let keywordsNovas = $state([]);
	let cronNova = $state('');

	function salvar() {
		if (keywordsNovas.length === 0) return;
		buscasSalvas.salvar({
			id: slugificar(keywordsNovas[0]),
			keywords: keywordsNovas,
			estrategia: 'nicho',
			cron: cronNova
		});
		keywordsNovas = [];
		cronNova = '';
		mostrarForm = false;
	}
</script>

<div class="gerenciar-buscas">
	<div class="cabecalho">
		<h2>🔍 Buscas Agendadas</h2>
		<button class="btn-nova" onclick={() => (mostrarForm = !mostrarForm)} type="button">
			{mostrarForm ? '✕ cancelar' : '+ nova busca'}
		</button>
	</div>

	{#if mostrarForm}
		<div class="form-nova">
			<TagInput bind:tags={keywordsNovas} label="palavras-chave" placeholder="ex.: kenzo, shiseido…" />
			<AgendadorBusca bind:value={cronNova} />
			<div class="form-acoes">
				<button class="salvar" onclick={salvar} disabled={keywordsNovas.length === 0} type="button">
					Salvar busca
				</button>
			</div>
		</div>
	{/if}

	{#if buscasKw.length > 0}
		<div class="buscas-lista">
			{#each buscasKw as b (b.id)}
				<BuscaCard busca={b} onremover={(id) => buscasSalvas.remover(id)} />
			{/each}
		</div>
	{:else if !mostrarForm}
		<p class="vazio">Nenhuma busca agendada. Clique em "+ nova busca" para criar.</p>
	{/if}
</div>

<style>
	.gerenciar-buscas { margin-bottom: var(--r6); }
	.cabecalho { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--r3); }
	h2 { font-size: 1.1rem; margin: 0; color: var(--tinta); }
	.btn-nova {
		border: 1px solid var(--linha); background: var(--porcelana);
		color: var(--tinta); font-size: 0.82rem; font-weight: 600;
		padding: 6px 14px; border-radius: var(--raio-full); cursor: pointer;
	}
	.btn-nova:hover { border-color: var(--ouro); color: var(--ouro); }
	.form-nova {
		background: var(--nevoa); border: 1px solid var(--linha);
		border-radius: var(--raio); padding: var(--r4);
		display: flex; flex-direction: column; gap: var(--r4); margin-bottom: var(--r4);
	}
	.form-acoes { display: flex; justify-content: flex-end; }
	.salvar {
		border: 1px solid var(--linha); background: var(--ouro-fundo);
		color: var(--ouro-escuro); font-weight: 600; font-size: 0.85rem;
		padding: 9px 18px; border-radius: var(--raio-sm); cursor: pointer;
	}
	.salvar:disabled { opacity: 0.5; cursor: not-allowed; }
	.buscas-lista { display: flex; flex-direction: column; gap: var(--r3); }
	.vazio { font-size: 0.85rem; color: var(--tinta-suave); font-style: italic; }
</style>
