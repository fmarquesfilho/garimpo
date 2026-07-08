<script>
	import { onMount } from 'svelte';
	import { buscarEstatisticas, buscarEvolucaoLojas, listarPublicacoes, listarBuscasServidor } from '$lib/api.js';
	import { brl, num, pctSinal } from '$lib/formatters.js';
	import {
		Loading,
		Alert,
		MetricCard,
		DashPanel,
		MiniChart,
		RankList,
		Select,
		Badge
	} from '$lib/components/ui/index.js';

	let dias = $state(7);
	const diasOpcoes = [7, 30, 90].map((d) => ({ value: String(d), label: `${d} dias` }));
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

	// Segmentação por fonte
	let porFonte = $derived(dados?.por_fonte ?? { lojas: {}, keywords: {} });
	let kwColetas = $derived(porFonte.keywords?.total_coletas ?? 0);
	let kwProdutos = $derived(porFonte.keywords?.total_produtos ?? 0);

	// Evolução — lojas e keywords separados
	let evolucaoLojas = $derived(evolucao?.lojas ?? []);
	let evolucaoKeywords = $derived(evolucao?.keywords ?? []);
	let resumoGlobal = $derived(evolucao?.resumo ?? {});
	let resumoKeywords = $derived(evolucao?.resumo_keywords ?? {});

	let totalQuedas = $derived(resumoGlobal.total_quedas ?? 0);
	let totalAltas = $derived(resumoGlobal.total_altas ?? 0);
	let quedasLojas = $derived(totalQuedas - (resumoKeywords.total_quedas ?? 0));
	let quedasKw = $derived(resumoKeywords.total_quedas ?? 0);
	let altasLojas = $derived(totalAltas - (resumoKeywords.total_altas ?? 0));
	let altasKw = $derived(resumoKeywords.total_altas ?? 0);
	let temBreakdown = $derived(quedasKw > 0 || altasKw > 0);

	// Top keywords por variação absoluta (para o painel de evolução)
	let topKeywords = $derived(
		[...evolucaoKeywords].sort((a, b) => Math.abs(b.variacao_media_pct) - Math.abs(a.variacao_media_pct)).slice(0, 3)
	);

	// Top produtos publicados
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
		<Select
			value={String(dias)}
			onchange={(v) => {
				dias = Number(v);
				carregar();
			}}
			options={diasOpcoes}
			size="sm"
			class="w-28"
		/>
	</header>

	{#if carregando}
		<Loading mensagem="Carregando…" />
	{:else if erro}
		<Alert variant="error"><p>{erro}</p></Alert>
	{:else}
		<!-- Row 1: métricas gerais -->
		<div class="mb-3 grid grid-cols-2 gap-3 sm:grid-cols-4">
			<MetricCard valor={String(lojas.length)} label="Lojas" />
			<MetricCard valor={num(dados?.total_amostras ?? 0)} label="Produtos coletados" />
			<MetricCard valor={String(pubEnviadas.length)} label="Publicações" variant="gold" />
			<MetricCard valor="{taxaSucesso}%" label="Taxa de sucesso" />
		</div>

		<!-- Row 2: métricas keywords + quedas/altas -->
		<div class="mb-5 grid grid-cols-2 gap-3 sm:grid-cols-4">
			<MetricCard valor={String(kwColetas)} label="Buscas keyword" />
			<MetricCard valor={num(kwProdutos)} label="Produtos (kw)" />
			<div>
				<MetricCard valor={String(totalQuedas)} label="↓ Quedas" variant="green" />
				{#if temBreakdown}
					<p class="mt-1 text-center text-xs text-muted-foreground">{quedasLojas} lojas · {quedasKw} kw</p>
				{/if}
			</div>
			<div>
				<MetricCard valor={String(totalAltas)} label="↑ Altas" variant="red" />
				{#if temBreakdown}
					<p class="mt-1 text-center text-xs text-muted-foreground">{altasLojas} lojas · {altasKw} kw</p>
				{/if}
			</div>
		</div>

		<!-- Painéis -->
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
			<DashPanel titulo="🏆 Mais publicados">
				<RankList items={topProdutos()} vazio="Nenhuma publicação ainda." />
			</DashPanel>

			<DashPanel titulo="📈 Preço médio (lojas)">
				{#if evolucaoLojas.length > 0}
					{#each evolucaoLojas.slice(0, 2) as loja (loja.busca_id)}
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
					<p class="m-0 text-sm italic text-muted-foreground">Sem dados de lojas no período.</p>
				{/if}
			</DashPanel>
		</div>

		<!-- Painel evolução keywords -->
		<div class="mt-4">
			<DashPanel titulo="📈 Preço médio (keywords)">
				{#if topKeywords.length > 0}
					{#each topKeywords as kw (kw.busca_id)}
						<div class="mb-4 last:mb-0">
							<div class="mb-2 flex items-center justify-between">
								<span class="text-sm font-semibold">{kw.busca_id}</span>
								<span
									class="text-sm font-bold"
									class:text-[var(--sucesso-texto)]={kw.variacao_media_pct < 0}
									class:text-[var(--erro-texto)]={kw.variacao_media_pct > 0}
								>
									{pctSinal(kw.variacao_media_pct)}
								</span>
							</div>
							{#if kw.pontos?.length >= 2}
								<MiniChart pontos={kw.pontos.map((p) => ({ data: p.data, valor: p.preco_medio }))} formatValor={brl} />
							{:else if kw.pontos?.length === 1}
								<Badge>{brl(kw.pontos[0].preco_medio)}</Badge>
							{/if}
						</div>
					{/each}
				{:else}
					<p class="m-0 text-sm italic text-muted-foreground">Sem dados de buscas por keyword no período.</p>
				{/if}
			</DashPanel>
		</div>
	{/if}
</div>
