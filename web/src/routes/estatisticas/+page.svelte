<script>
	import { onMount } from 'svelte';
	import { buscarEstatisticas, buscarEvolucaoLojas, listarPublicacoes, listarBuscasServidor } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';
	import { brl, num, pctSinal } from '$lib/formatters.js';
	import { Loading, Alert } from '$lib/components/ui/index.js';

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

	// Top produtos publicados (mais frequentes)
	let topProdutos = $derived(() => {
		const contagem = {};
		for (const p of pubEnviadas) {
			const nome = p.nome || '(sem título)';
			contagem[nome] = (contagem[nome] || 0) + 1;
		}
		return Object.entries(contagem)
			.sort((a, b) => b[1] - a[1])
			.slice(0, 5)
			.map(([nome, vezes]) => ({ nome, vezes }));
	});

	// Variações detectadas (do evolucao)
	let totalQuedas = $derived(evolucao?.resumo?.total_quedas ?? 0);
	let totalAltas = $derived(evolucao?.resumo?.total_altas ?? 0);
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
		<!-- Grid de métricas -->
		<div class="grid-metricas">
			<div class="metrica">
				<span class="metrica-valor">{lojas.length}</span>
				<span class="metrica-label">Lojas</span>
			</div>
			<div class="metrica">
				<span class="metrica-valor">{num(dados?.total_amostras ?? 0)}</span>
				<span class="metrica-label">Produtos coletados</span>
			</div>
			<div class="metrica ouro">
				<span class="metrica-valor">{pubEnviadas.length}</span>
				<span class="metrica-label">Publicações</span>
			</div>
			<div class="metrica">
				<span class="metrica-valor">{taxaSucesso}%</span>
				<span class="metrica-label">Taxa de sucesso</span>
			</div>
			<div class="metrica verde">
				<span class="metrica-valor">{totalQuedas}</span>
				<span class="metrica-label">↓ Quedas detectadas</span>
			</div>
			<div class="metrica vermelho">
				<span class="metrica-valor">{totalAltas}</span>
				<span class="metrica-label">↑ Altas detectadas</span>
			</div>
		</div>

		<!-- Painel inferior: 2 colunas -->
		<div class="grid-paineis">
			<!-- Coluna 1: Top produtos publicados -->
			<div class="painel">
				<h2>🏆 Mais publicados</h2>
				{#if topProdutos().length === 0}
					<p class="vazio-painel">Nenhuma publicação ainda.</p>
				{:else}
					<div class="lista-top">
						{#each topProdutos() as item, i}
							<div class="top-item">
								<span class="top-pos">{i + 1}</span>
								<span class="top-nome">{item.nome}</span>
								<span class="top-vezes">{item.vezes}×</span>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Coluna 2: Evolução de preço (mini) -->
			<div class="painel">
				<h2>📈 Preço médio</h2>
				{#if evolucao?.lojas?.length > 0}
					{#each evolucao.lojas.slice(0, 2) as loja (loja.busca_id)}
						<div class="evo-mini">
							<div class="evo-header">
								<span class="evo-nome">{loja.busca_id}</span>
								<span class="evo-var" class:verde={loja.variacao_media_pct < 0} class:vermelho={loja.variacao_media_pct > 0}>
									{pctSinal(loja.variacao_media_pct)}
								</span>
							</div>
							{#if loja.pontos?.length > 1}
								{@const maxP = Math.max(...loja.pontos.map(p => p.preco_medio))}
								{@const minP = Math.min(...loja.pontos.map(p => p.preco_medio))}
								{@const range = maxP - minP || 1}
								<div class="mini-barras">
									{#each loja.pontos as ponto, i}
										{@const h = ((ponto.preco_medio - minP) / range) * 100}
										<div class="mini-barra" style="height: {Math.max(h, 8)}%"
											class:queda={i > 0 && ponto.preco_medio < loja.pontos[i-1].preco_medio}
											class:alta={i > 0 && ponto.preco_medio > loja.pontos[i-1].preco_medio}
											title="{ponto.data}: {brl(ponto.preco_medio)}"
										></div>
									{/each}
								</div>
								<div class="evo-range">
									<span>{brl(minP)}</span>
									<span>{brl(maxP)}</span>
								</div>
							{:else}
								<p class="vazio-painel">Aguardando 2ª coleta…</p>
							{/if}
						</div>
					{/each}
				{:else}
					<p class="vazio-painel">Sem dados de evolução no período.</p>
				{/if}
			</div>
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

	/* Grid de métricas */
	.grid-metricas {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: var(--r3);
		margin-bottom: var(--r5);
	}
	.metrica {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		padding: var(--r3) var(--r4);
		text-align: center;
	}
	.metrica-valor {
		display: block;
		font-size: 1.5rem;
		font-weight: 700;
		font-family: var(--mono);
		line-height: 1.2;
	}
	.metrica-label {
		font-size: var(--text-xs);
		color: var(--tinta-suave);
		text-transform: uppercase;
		font-weight: 600;
	}
	.metrica.ouro .metrica-valor { color: var(--ouro); }
	.metrica.verde .metrica-valor { color: var(--sucesso-texto); }
	.metrica.vermelho .metrica-valor { color: var(--erro-texto); }

	/* Grid de painéis */
	.grid-paineis {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--r4);
	}
	.painel {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r4);
	}
	.painel h2 { font-size: 1rem; margin: 0 0 var(--r3); }
	.vazio-painel { font-size: var(--text-sm); color: var(--tinta-suave); font-style: italic; }

	/* Top produtos */
	.lista-top { display: flex; flex-direction: column; gap: 2px; }
	.top-item {
		display: flex;
		align-items: center;
		gap: var(--r2);
		padding: 6px 0;
		border-bottom: 1px solid var(--linha);
	}
	.top-item:last-child { border-bottom: none; }
	.top-pos { font-size: var(--text-xs); font-weight: 700; color: var(--tinta-suave); width: 18px; }
	.top-nome { flex: 1; font-size: var(--text-sm); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.top-vezes { font-size: var(--text-xs); font-weight: 700; color: var(--ouro); }

	/* Evolução mini */
	.evo-mini { margin-bottom: var(--r4); }
	.evo-mini:last-child { margin-bottom: 0; }
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

	.mini-barras {
		display: flex;
		align-items: flex-end;
		gap: 2px;
		height: 40px;
	}
	.mini-barra {
		flex: 1;
		background: var(--ouro-claro);
		border-radius: 2px 2px 0 0;
		min-height: 3px;
	}
	.mini-barra.queda { background: var(--sucesso-texto); }
	.mini-barra.alta { background: var(--erro-texto); }

	.evo-range {
		display: flex;
		justify-content: space-between;
		font-size: 0.65rem;
		color: var(--tinta-suave);
		margin-top: 2px;
	}

	@media (max-width: 600px) {
		.grid-metricas { grid-template-columns: repeat(2, 1fr); }
		.grid-paineis { grid-template-columns: 1fr; }
	}
</style>
