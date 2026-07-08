<script>
	/**
	 * Gerenciamento de buscas salvas — criar com todas as opções: keywords,
	 * lojas, categorias, fontes, dias_janela, e agendamento.
	 */
	import { buscasSalvas, slugificar } from '$lib/buscas.js';
	import TagInput from './TagInput.svelte';
	import AgendadorBusca from './AgendadorBusca.svelte';
	import BuscaCard from './BuscaCard.svelte';
	import PainelNovidades from './PainelNovidades.svelte';
	import { Button, Checkbox, Select } from '$lib/components/ui';

	let buscasKw = $derived(($buscasSalvas ?? []).filter((b) => !b.shop_ids?.length));

	let mostrarForm = $state(false);
	let buscaSelecionada = $state(null);
	let keywordsNovas = $state([]);
	let categoriasNovas = $state([]);
	let cronNova = $state('');
	let diasJanela = $state('7');
	let fontes = $state({ curadoria: true, quedas: false, novos: false });

	const diasOpcoes = [1, 2, 3, 7, 14, 30].map((d) => ({ value: String(d), label: `${d} ${d === 1 ? 'dia' : 'dias'}` }));

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
			dias_janela: Number(diasJanela),
			estrategia: 'nicho',
			cron: cronNova
		});
		keywordsNovas = [];
		categoriasNovas = [];
		cronNova = '';
		diasJanela = '7';
		fontes = { curadoria: true, quedas: false, novos: false };
		mostrarForm = false;
		buscaSelecionada = null;
	}

	function selecionarBusca(busca) {
		buscaSelecionada = buscaSelecionada?.id === busca.id ? null : busca;
	}

	function removerBusca(id) {
		if (buscaSelecionada?.id === id) buscaSelecionada = null;
		buscasSalvas.remover(id);
	}
</script>

<div class="mb-6">
	<div class="mb-1 flex items-center justify-between">
		<h2 class="m-0 text-lg text-foreground">🔍 Buscas por palavra-chave</h2>
		<Button variant="secondary" size="sm" onclick={() => { mostrarForm = !mostrarForm; buscaSelecionada = null; }}>
			{mostrarForm ? '✕ cancelar' : '+ nova busca'}
		</Button>
	</div>
	<p class="mb-3 text-sm text-muted-foreground">
		Agende buscas por palavra-chave na Shopee inteira, com ou sem coleta automática. Para monitorar palavras-chave
		dentro de uma loja específica, use o formulário “Adicionar loja” acima.
	</p>

	{#if mostrarForm}
		<div class="mb-4 flex flex-col gap-4 rounded-md border border-border bg-card p-4">
			<TagInput bind:tags={keywordsNovas} label="Palavras-chave (opcional)" placeholder="ex.: sérum, skin1004…" />
			<TagInput bind:tags={categoriasNovas} label="Categorias (opcional)" placeholder="ex.: cosméticos, perfumaria…" />

			<!-- Fontes -->
			<div class="flex flex-col gap-1.5">
				<span class="text-sm font-semibold text-foreground">Fontes de dados:</span>
				<div class="flex flex-wrap gap-4">
					<Checkbox bind:checked={fontes.curadoria} label="🔍 Curadoria" />
					<Checkbox bind:checked={fontes.quedas} label="📉 Quedas de preço" />
					<Checkbox bind:checked={fontes.novos} label="🆕 Produtos novos" />
				</div>
			</div>

			<!-- Dias janela (para novos) -->
			{#if fontes.novos}
				<div class="flex flex-wrap items-center gap-2">
					<span class="text-sm font-semibold text-foreground">Considerar "novo" se apareceu nos últimos:</span>
					<Select bind:value={diasJanela} options={diasOpcoes} size="sm" class="w-32" />
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
				<BuscaCard busca={b} selecionado={buscaSelecionada?.id === b.id} onremover={removerBusca} onselecionar={selecionarBusca} />
			{/each}
		</div>

		{#if buscaSelecionada}
			<PainelNovidades buscaId={buscaSelecionada.id} keywords={buscaSelecionada.keywords ?? []} />
		{/if}
	{:else if !mostrarForm}
		<p class="text-sm italic text-muted-foreground">Nenhuma busca agendada. Clique em "+ nova busca" para criar.</p>
	{/if}
</div>
