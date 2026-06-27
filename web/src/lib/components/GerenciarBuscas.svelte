<script>
	/**
	 * Gerenciamento de buscas salvas — criar com todas as opções: keywords,
	 * lojas, categorias, fontes, dias_janela, e agendamento.
	 */
	import { buscasSalvas, slugificar } from '$lib/buscas.js';
	import TagInput from './TagInput.svelte';
	import AgendadorBusca from './AgendadorBusca.svelte';
	import BuscaCard from './BuscaCard.svelte';

	let buscasKw = $derived(($buscasSalvas ?? []).filter(b => !b.shop_ids?.length));

	let mostrarForm = $state(false);
	let keywordsNovas = $state([]);
	let categoriasNovas = $state([]);
	let cronNova = $state('');
	let diasJanela = $state(7);
	let fontes = $state({ curadoria: true, quedas: false, novos: false });

	let fontesArray = $derived(
		Object.entries(fontes).filter(([, v]) => v).map(([k]) => k)
	);

	function salvar() {
		if (keywordsNovas.length === 0 && categoriasNovas.length === 0 && fontesArray.length === 0) return;
		const id = slugificar(keywordsNovas[0] ?? categoriasNovas[0] ?? fontesArray[0]);
		buscasSalvas.salvar({
			id,
			keywords: keywordsNovas,
			categorias: categoriasNovas.length > 0 ? categoriasNovas : undefined,
			fontes: fontesArray,
			dias_janela: diasJanela,
			estrategia: 'nicho',
			cron: cronNova
		});
		keywordsNovas = [];
		categoriasNovas = [];
		cronNova = '';
		diasJanela = 7;
		fontes = { curadoria: true, quedas: false, novos: false };
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
			<TagInput bind:tags={keywordsNovas} label="Palavras-chave (opcional)" placeholder="ex.: sérum, skin1004…" />
			<TagInput bind:tags={categoriasNovas} label="Categorias (opcional)" placeholder="ex.: cosméticos, perfumaria…" />

			<!-- Fontes -->
			<div class="fontes-config">
				<span class="rotulo">Fontes de dados:</span>
				<div class="fontes-toggles">
					<label><input type="checkbox" bind:checked={fontes.curadoria} /> 🔍 Curadoria</label>
					<label><input type="checkbox" bind:checked={fontes.quedas} /> 📉 Quedas de preço</label>
					<label><input type="checkbox" bind:checked={fontes.novos} /> 🆕 Produtos novos</label>
				</div>
			</div>

			<!-- Dias janela (para novos) -->
			{#if fontes.novos}
				<div class="campo-dias">
					<label for="dias-janela">Considerar "novo" se apareceu nos últimos:</label>
					<select id="dias-janela" bind:value={diasJanela}>
						<option value={1}>1 dia</option>
						<option value={2}>2 dias</option>
						<option value={3}>3 dias</option>
						<option value={7}>7 dias</option>
						<option value={14}>14 dias</option>
						<option value={30}>30 dias</option>
					</select>
				</div>
			{/if}

			<AgendadorBusca bind:value={cronNova} />

			<div class="form-acoes">
				<button class="salvar" onclick={salvar} disabled={keywordsNovas.length === 0 && categoriasNovas.length === 0 && fontesArray.length === 0} type="button">
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
	.fontes-config { display: flex; flex-direction: column; gap: 6px; }
	.fontes-config .rotulo { font-size: 0.82rem; font-weight: 600; color: var(--tinta); }
	.fontes-toggles { display: flex; flex-wrap: wrap; gap: var(--r3); }
	.fontes-toggles label {
		font-size: 0.85rem; display: flex; align-items: center; gap: 4px; cursor: pointer;
	}
	.fontes-toggles input { accent-color: var(--ouro); }
	.campo-dias { display: flex; flex-wrap: wrap; align-items: center; gap: var(--r2); }
	.campo-dias label { font-size: 0.82rem; font-weight: 600; color: var(--tinta); }
	.campo-dias select { padding: 6px 10px; border: 1px solid var(--linha); border-radius: 8px; font-size: 0.85rem; }
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
