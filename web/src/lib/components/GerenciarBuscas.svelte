<script>
	/**
	 * Gerenciamento de buscas salvas — criar com todas as opções: keywords,
	 * lojas, categorias, fontes, dias_janela, e agendamento.
	 */
	import { buscasSalvas, slugificar } from '$lib/buscas.js';
	import TagInput from './TagInput.svelte';
	import AgendadorBusca from './AgendadorBusca.svelte';
	import BuscaCard from './BuscaCard.svelte';
	import { Button } from '$lib/components/ui';

	let buscasKw = $derived(($buscasSalvas ?? []).filter((b) => !b.shop_ids?.length));

	let mostrarForm = $state(false);
	let keywordsNovas = $state([]);
	let categoriasNovas = $state([]);
	let cronNova = $state('');
	let diasJanela = $state(7);
	let fontes = $state({ curadoria: true, quedas: false, novos: false });

	let fontesArray = $derived(
		Object.entries(fontes)
			.filter(([, v]) => v)
			.map(([k]) => k)
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

<div class="mb-6">
	<div class="mb-3 flex items-center justify-between">
		<h2 class="m-0 text-lg text-foreground">🔍 Buscas Agendadas</h2>
		<Button variant="secondary" size="sm" onclick={() => (mostrarForm = !mostrarForm)}>
			{mostrarForm ? '✕ cancelar' : '+ nova busca'}
		</Button>
	</div>

	{#if mostrarForm}
		<div class="mb-4 flex flex-col gap-4 rounded-md border border-border bg-card p-4">
			<TagInput bind:tags={keywordsNovas} label="Palavras-chave (opcional)" placeholder="ex.: sérum, skin1004…" />
			<TagInput bind:tags={categoriasNovas} label="Categorias (opcional)" placeholder="ex.: cosméticos, perfumaria…" />

			<!-- Fontes -->
			<div class="flex flex-col gap-1.5">
				<span class="text-sm font-semibold text-foreground">Fontes de dados:</span>
				<div class="flex flex-wrap gap-3">
					<label class="flex cursor-pointer items-center gap-1 text-sm"
						><input type="checkbox" class="accent-[var(--ouro)]" bind:checked={fontes.curadoria} /> 🔍 Curadoria</label
					>
					<label class="flex cursor-pointer items-center gap-1 text-sm"
						><input type="checkbox" class="accent-[var(--ouro)]" bind:checked={fontes.quedas} /> 📉 Quedas de preço</label
					>
					<label class="flex cursor-pointer items-center gap-1 text-sm"
						><input type="checkbox" class="accent-[var(--ouro)]" bind:checked={fontes.novos} /> 🆕 Produtos novos</label
					>
				</div>
			</div>

			<!-- Dias janela (para novos) -->
			{#if fontes.novos}
				<div class="flex flex-wrap items-center gap-2">
					<label for="dias-janela" class="text-sm font-semibold text-foreground"
						>Considerar "novo" se apareceu nos últimos:</label
					>
					<select
						id="dias-janela"
						class="rounded-lg border border-border px-2.5 py-1.5 text-sm"
						bind:value={diasJanela}
					>
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

			<div class="flex justify-end">
				<Button
					onclick={salvar}
					disabled={keywordsNovas.length === 0 && categoriasNovas.length === 0 && fontesArray.length === 0}
				>
					Salvar busca
				</Button>
			</div>
		</div>
	{/if}

	{#if buscasKw.length > 0}
		<div class="flex flex-col gap-3">
			{#each buscasKw as b (b.id)}
				<BuscaCard busca={b} onremover={(id) => buscasSalvas.remover(id)} />
			{/each}
		</div>
	{:else if !mostrarForm}
		<p class="text-sm italic text-tinta-suave">Nenhuma busca agendada. Clique em "+ nova busca" para criar.</p>
	{/if}
</div>
