<script>
	import { onMount } from 'svelte';
	import { buscarEstatisticas, buscarEvolucaoLojas, listarPublicacoes, listarBuscasServidor } from '$lib/api.js';
	import { brl, num, pctSinal } from '$lib/formatters.js';
	import { Loading, Alert, MetricCard, DashPanel, MiniChart, RankList } from '$lib/components/ui/index.js';

	let dias = $state(7);
	let carregando = $state(true);
	let erro = $state(null);

	let dados = $state(null);
	let evolucao = $state(null);
	let publicacoes = $state([]);
	let buscas = $state([]);

	async function carregar() {
		carregando = true;
		erro = null;
		try {
			const [est, evo, pub, bus] = await Promise.all([
				buscarEstatisticas({ dias }).catch(() => null),
				buscarEvolucaoLojas({ dias }).catch(() => null),
				listarPublicacoes({ status: '' }).catch(() => ({ publicacoes: [] })),
				listarBuscasServidor().catch(() => ({ buscas: [] }))
			]);
			dados = est;
			evolucao = evo;
			publicacoes = pub?.publicacoes ?? [];
			buscas = bus?.buscas ?? [];
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}
	}

	onMount(carregar);

	// Derivados — reactive computed values for dashboard statistics
	let lojas = $derived((buscas ?? []).filter((b) => b.shop_ids?.length > 0));
	let pubEnviadas = $derived((publicacoes ?? []).filter((p) => p.status === 'enviada'));
	let pubErros = $derived((publicacoes ?? []).filter((p) => p.status === 'erro'));
	let taxaSucesso = $derived(
		pubEnviadas.length + pubErros.length > 0
			? Math.round((pubEnviadas.length / (pubEnviadas.length + pubErros.length)) * 100)
			: 100
	);
	let totalQuedas = $derived(evolucao?.resumo?.total_quedas ?? 0);
	let totalAltas = $derived(evolucao?.resumo?.total_altas ?? 0);

	// Top produtos
	let topProdutos = $derived(() => {
		const contagem = {};
		for (const p of pubEnviadas) {
			const nome = p.nome || '(sem título)';
			contagem[nome] = (contagem[nome] || 0) + 1;
		}
		return Object.entries(contagem)
			.sort((a, b) => b[1] - a[1])
			.slice(0, 5)
			.map(([nome, vezes]) => ({ nome, valor: `${vezes}×` }));
	});
</script>

<svelte:head>
	<title>Estatísticas — Garimpei</title>
</svelte:head>

<div class="max-w-[900px]">
	<header class="mb-6 flex items-center justify-between">
		<h1 class="m-0 text-2xl">📊 Dashboard</h1>
		<select
			bind:value={dias}
			onchange={carregar}
			class="rounded-sm border border-border bg-porcelana px-3 py-1.5 font-mono text-sm"
		>
			<option value={7}>7 dias</option>
			<option value={30}>30 dias</option>
			<option value={90}>90 dias</option>
		</select>
	</header>

	{#if carregando}
		<Loading mensagem="Carregando…" />
	{:else if erro}
		<Alert variant="error"><p>{erro}</p></Alert>
	{:else}
		<div class="mb-5 grid grid-cols-2 gap-3 sm:grid-cols-3">
			<MetricCard valor={String(lojas.length)} label="Lojas" />
			<MetricCard valor={num(dados?.total_amostras ?? 0)} label="Produtos coletados" />
			<MetricCard valor={String(pubEnviadas.length)} label="Publicações" variant="gold" />
			<MetricCard valor="{taxaSucesso}%" label="Taxa de sucesso" />
			<MetricCard valor={String(totalQuedas)} label="↓ Quedas" variant="green" />
			<MetricCard valor={String(totalAltas)} label="↑ Altas" variant="red" />
		</div>

		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
			<DashPanel titulo="🏆 Mais publicados">
				<RankList items={topProdutos()} vazio="Nenhuma publicação ainda." />
			</DashPanel>

			<DashPanel titulo="📈 Preço médio">
				{#if evolucao?.lojas?.length > 0}
					{#each evolucao.lojas.slice(0, 2) as loja (loja.busca_id)}
						<div class="mb-4 last:mb-0">
							<div class="mb-2 flex items-center justify-between">
								<span class="text-sm font-semibold">{loja.busca_id}</span>
								<span
									class="text-sm font-bold"
									class:text-[var(--sucesso-texto)]={loja.variacao_media_pct < 0}
									class:text-[var(--erro-texto)]={loja.variacao_media_pct > 0}
								>
									{pctSinal(loja.variacao_media_pct)}
								</span>
							</div>
							<MiniChart
								pontos={loja.pontos?.map((p) => ({ data: p.data, valor: p.preco_medio })) ?? []}
								formatValor={brl}
							/>
						</div>
					{/each}
				{:else}
					<p class="m-0 text-sm italic text-tinta-suave">Sem dados de evolução no período.</p>
				{/if}
			</DashPanel>
		</div>
	{/if}
</div>
