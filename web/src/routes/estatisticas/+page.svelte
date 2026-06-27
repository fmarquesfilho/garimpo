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

	// Derivados
	let lojas = $derived((buscas ?? []).filter(b => b.shop_ids?.length > 0));
	let pubEnviadas = $derived((publicacoes ?? []).filter(p => p.status === 'enviada'));
	let pubErros = $derived((publicacoes ?? []).filter(p => p.status === 'erro'));
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

<div class="dashboard">
	<header class="dash-header">
		<h1>📊 Dashboard</h1>
		<select bind:value={dias} onchange={carregar} class="periodo">
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
		<div class="grid-metricas">
			<MetricCard valor={String(lojas.length)} label="Lojas" />
			<MetricCard valor={num(dados?.total_amostras ?? 0)} label="Produtos coletados" />
			<MetricCard valor={String(pubEnviadas.length)} label="Publicações" variant="gold" />
			<MetricCard valor="{taxaSucesso}%" label="Taxa de sucesso" />
			<MetricCard valor={String(totalQuedas)} label="↓ Quedas" variant="green" />
			<MetricCard valor={String(totalAltas)} label="↑ Altas" variant="red" />
		</div>

		<div class="grid-paineis">
			<DashPanel titulo="🏆 Mais publicados">
				<RankList items={topProdutos()} vazio="Nenhuma publicação ainda." />
			</DashPanel>

			<DashPanel titulo="📈 Preço médio">
				{#if evolucao?.lojas?.length > 0}
					{#each evolucao.lojas.slice(0, 2) as loja (loja.busca_id)}
						<div class="evo-bloco">
							<div class="evo-header">
								<span class="evo-nome">{loja.busca_id}</span>
								<span class="evo-var" class:verde={loja.variacao_media_pct < 0} class:vermelho={loja.variacao_media_pct > 0}>
									{pctSinal(loja.variacao_media_pct)}
								</span>
							</div>
							<MiniChart
								pontos={loja.pontos?.map(p => ({ data: p.data, valor: p.preco_medio })) ?? []}
								formatValor={brl}
							/>
						</div>
					{/each}
				{:else}
					<p class="vazio">Sem dados de evolução no período.</p>
				{/if}
			</DashPanel>
		</div>
	{/if}
</div>

<style>
	.dashboard { max-width: 900px; }

	.dash-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--r6);
	}
	.dash-header h1 { font-size: 1.5rem; margin: 0; }
	.periodo {
		font-family: var(--mono);
		padding: 6px 12px;
		border-radius: var(--raio-sm);
		border: 1px solid var(--linha);
		background: var(--porcelana);
		font-size: var(--text-sm);
	}

	.grid-metricas {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: var(--r3);
		margin-bottom: var(--r5);
	}

	.grid-paineis {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--r4);
	}

	.evo-bloco { margin-bottom: var(--r4); }
	.evo-bloco:last-child { margin-bottom: 0; }
	.evo-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--r2);
	}
	.evo-nome { font-size: var(--text-sm); font-weight: 600; }
	.evo-var { font-size: var(--text-sm); font-weight: 700; }
	.verde { color: var(--sucesso-texto); }
	.vermelho { color: var(--erro-texto); }
	.vazio { font-size: var(--text-sm); color: var(--tinta-suave); font-style: italic; margin: 0; }

	@media (max-width: 600px) {
		.grid-metricas { grid-template-columns: repeat(2, 1fr); }
		.grid-paineis { grid-template-columns: 1fr; }
	}
</style>
