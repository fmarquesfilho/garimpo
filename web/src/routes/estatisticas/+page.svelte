<script>
	import { onMount } from 'svelte';
	import { buscarEstatisticas, buscarEvolucaoLojas, listarPublicacoes, listarBuscasServidor } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';
	import { brl, pct, num, pctSinal, tempoAtras } from '$lib/formatters.js';
	import { PageHeader, Loading, EmptyState, Alert, StatCard } from '$lib/components/ui/index.js';

	let dias = $state(7);
	let carregando = $state(true);
	let erro = $state(null);

	// Dados
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
	let lojasMonitoradas = $derived((buscas ?? []).filter(b => b.shop_ids?.length > 0));
	let pubEnviadas = $derived((publicacoes ?? []).filter(p => p.status === 'enviada'));
	let pubErros = $derived((publicacoes ?? []).filter(p => p.status === 'erro'));
</script>

<PageHeader
	rotulo="análise"
	titulo="📊 Estatísticas"
	subtitulo="Evolução de preço das lojas monitoradas e resumo da operação."
/>

<label class="janela">
	período:
	<select bind:value={dias} onchange={carregar} class="dado">
		<option value={7}>7 dias</option>
		<option value={30}>30 dias</option>
		<option value={90}>90 dias</option>
	</select>
</label>

<!-- ── Resumo ──────────────────────────────────────────────────────────────── -->
{#if carregando}
	<Loading mensagem="Carregando…" />
{:else if erro}
	<Alert variant="error"><p>{erro}</p></Alert>
{:else}
	<section class="secao">
		<div class="resumo-cards">
			<StatCard label="Lojas monitoradas" valor={String(lojasMonitoradas.length)} />
			<StatCard label="Produtos coletados" valor={num(dados?.total_amostras ?? 0)} />
			<StatCard label="Publicações" valor={String(pubEnviadas.length)} variant="gold" />
		</div>
	</section>
{/if}

<!-- ── Evolução de preço (só aparece com dados reais) ──────────────────── -->
{#if $usuario && evolucao?.lojas?.length > 0}
<section class="secao">
	<h2>📈 Evolução de preço</h2>
	<p class="sub">Como os preços das lojas monitoradas mudaram no período.</p>

	{#if carregando}
		<Loading mensagem="Calculando…" />
	{:else}
		<!-- Resumo geral -->
		<div class="resumo-cards">
			<StatCard label="Lojas" valor={String(evolucao.resumo.total_lojas)} />
			<StatCard label="Produtos" valor={num(evolucao.resumo.total_produtos)} />
			<StatCard label="Preço médio" valor={brl(evolucao.resumo.preco_medio_global)} variant="gold" />
			<StatCard label="Variação média" valor={pctSinal(evolucao.resumo.variacao_media_global_pct)}
				variant={evolucao.resumo.variacao_media_global_pct > 0 ? 'positive' : evolucao.resumo.variacao_media_global_pct < 0 ? 'negative' : 'default'} />
			<StatCard label="↓ Quedas" valor={String(evolucao.resumo.total_quedas)} variant="negative" />
			<StatCard label="↑ Altas" valor={String(evolucao.resumo.total_altas)} variant="positive" />
		</div>

		<!-- Detalhes por loja -->
		{#each evolucao.lojas as loja (loja.busca_id)}
			<div class="loja-evo">
				<div class="loja-evo-header">
					<h3>🏪 {loja.busca_id}</h3>
					<div class="loja-evo-meta">
						<span>{loja.total_produtos} produtos</span>
						<span>{loja.coletas} coletas</span>
						<span class:positivo={loja.variacao_media_pct > 0} class:negativo={loja.variacao_media_pct < 0}>
							{pctSinal(loja.variacao_media_pct)}
						</span>
					</div>
				</div>

				<!-- Série temporal simplificada (barras) -->
				{#if loja.pontos?.length > 1}
					{@const maxPreco = Math.max(...loja.pontos.map(p => p.preco_medio))}
					{@const minPreco = Math.min(...loja.pontos.map(p => p.preco_medio))}
					{@const range = maxPreco - minPreco || 1}
					<div class="serie-temporal">
						<div class="serie-labels">
							<span class="serie-label-max">{brl(maxPreco)}</span>
							<span class="serie-label-min">{brl(minPreco)}</span>
						</div>
						<div class="serie-barras">
							{#each loja.pontos as ponto, i}
								{@const altura = ((ponto.preco_medio - minPreco) / range) * 100}
								<div class="barra-wrapper" title="{ponto.data}: {brl(ponto.preco_medio)} ({ponto.produtos} produtos)">
									<div class="barra" style="height: {Math.max(altura, 5)}%"
										class:barra-queda={i > 0 && ponto.preco_medio < loja.pontos[i-1].preco_medio}
										class:barra-alta={i > 0 && ponto.preco_medio > loja.pontos[i-1].preco_medio}
									></div>
								</div>
							{/each}
						</div>
						<div class="serie-datas">
							<span>{loja.pontos[0]?.data}</span>
							<span>{loja.pontos[loja.pontos.length - 1]?.data}</span>
						</div>
					</div>
				{/if}

				<!-- Top variações -->
				{#if loja.top_quedas?.length > 0 || loja.top_altas?.length > 0}
					<div class="variacoes-resumo">
						{#if loja.top_quedas?.length > 0}
							<div class="var-grupo">
								<h4 class="verde">↓ Maiores quedas</h4>
								{#each loja.top_quedas.slice(0, 3) as v}
									<div class="var-item">
										<span class="var-nome">{v.nome}</span>
										<span class="badge-var badge-baixou">↓ {Math.abs(v.variacao_pct * 100).toFixed(1)}%</span>
										<span class="var-precos">{brl(v.preco_anterior)} → {brl(v.preco_atual)}</span>
									</div>
								{/each}
							</div>
						{/if}
						{#if loja.top_altas?.length > 0}
							<div class="var-grupo">
								<h4 class="vermelho">↑ Maiores altas</h4>
								{#each loja.top_altas.slice(0, 3) as v}
									<div class="var-item">
										<span class="var-nome">{v.nome}</span>
										<span class="badge-var badge-subiu">↑ {Math.abs(v.variacao_pct * 100).toFixed(1)}%</span>
										<span class="var-precos">{brl(v.preco_anterior)} → {brl(v.preco_atual)}</span>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				{/if}
			</div>
		{/each}
	{/if}
</section>
{/if}

<style>
	h2 { font-size: 1.3rem; margin: 0 0 var(--r2); }

	h4 { font-size: 0.85rem; margin: 0 0 var(--r2); font-weight: 600; }
	.sub { color: var(--tinta-suave); margin: 0 0 var(--r4); font-size: 0.9rem; }
	.secao { margin-bottom: var(--r8); }
	.secao-lojas { margin-top: var(--r8); }
	.janela { font-size: 0.85rem; color: var(--tinta-suave); margin-bottom: var(--r6); display: block; }
	.janela select {
		font-family: var(--mono); padding: 6px 10px; border-radius: var(--raio-sm);
		border: 1px solid var(--linha); background: var(--porcelana); margin-left: 6px;
	}
	.meta { font-size: 0.8rem; color: var(--tinta-suave); margin: 0 0 var(--r4); }

	/* Resumo cards */
	.resumo-cards {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
		gap: var(--r3);
		margin-bottom: var(--r6);
	}

	/* Lojas e evolução */

	/* Loja evolução */
	.loja-evo {
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r4);
		margin-bottom: var(--r4);
		background: var(--branco);
	}
	.loja-evo-header {
		display: flex; justify-content: space-between; align-items: center;
		margin-bottom: var(--r4); flex-wrap: wrap; gap: var(--r2);
	}
	.loja-evo-meta { display: flex; gap: var(--r3); font-size: 0.8rem; color: var(--tinta-suave); }
	.positivo { color: var(--erro-texto); }
	.negativo { color: var(--sucesso-texto); }
	.verde { color: var(--sucesso-texto); }
	.vermelho { color: var(--erro-texto); }

	/* Série temporal */
	.serie-temporal { margin-bottom: var(--r4); display: flex; flex-direction: column; gap: 4px; }
	.serie-labels { display: flex; justify-content: space-between; font-size: 0.7rem; color: var(--tinta-suave); }
	.serie-barras { display: flex; align-items: flex-end; gap: 2px; height: 60px; padding: 4px 0; }
	.barra-wrapper { flex: 1; height: 100%; display: flex; align-items: flex-end; }
	.barra { width: 100%; background: var(--ouro); border-radius: 2px 2px 0 0; min-height: 3px; transition: height 0.3s ease; }
	.barra-queda { background: var(--sucesso-texto); }
	.barra-alta { background: var(--erro-texto); }
	.serie-datas { display: flex; justify-content: space-between; font-size: 0.68rem; color: var(--tinta-suave); }

	/* Variações resumo */
	.variacoes-resumo { display: grid; grid-template-columns: 1fr 1fr; gap: var(--r4); }
	.var-item { display: flex; align-items: center; gap: var(--r2); font-size: 0.8rem; padding: 4px 0; border-bottom: 1px solid var(--linha); }
	.var-item:last-child { border-bottom: none; }
	.var-nome { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 150px; }
	.badge-var { padding: 1px 6px; border-radius: var(--raio-full); font-size: 0.7rem; font-weight: 700; white-space: nowrap; }
	.badge-baixou { background: var(--sucesso-fundo); color: var(--sucesso-texto); }
	.badge-subiu { background: var(--erro-fundo); color: var(--erro-texto); }
	.var-precos { font-size: 0.72rem; color: var(--tinta-suave); white-space: nowrap; }

	@media (max-width: 720px) {
		.cab { display: none; }
		.linha { grid-template-columns: 1fr 1fr; gap: 4px; }
		.variacoes-resumo { grid-template-columns: 1fr; }
		.resumo-cards { grid-template-columns: repeat(3, 1fr); }
	}
</style>
