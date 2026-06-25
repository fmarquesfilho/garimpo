<script>
	import { onMount } from 'svelte';
	import { buscarEstatisticas, buscarEvolucaoLojas } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';
	import { brl, pct, num, pctSinal } from '$lib/formatters.js';
	import { PageHeader, Loading, EmptyState, Alert, StatCard } from '$lib/components/ui/index.js';

	let dias = $state(30);
	let dados = $state(null);
	let carregando = $state(true);
	let erro = $state(null);

	// Evolução de lojas
	let evolucao = $state(null);
	let carregandoEvolucao = $state(false);
	let erroEvolucao = $state(null);

	async function carregar() {
		carregando = true;
		erro = null;
		try {
			dados = await buscarEstatisticas({ dias });
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}
		if ($usuario) {
			carregarEvolucao();
		}
	}

	async function carregarEvolucao() {
		carregandoEvolucao = true;
		erroEvolucao = null;
		try {
			evolucao = await buscarEvolucaoLojas({ dias });
		} catch (e) {
			erroEvolucao = e.message;
		} finally {
			carregandoEvolucao = false;
		}
	}

	onMount(carregar);
</script>

<PageHeader
	rotulo="análise de mercado"
	titulo="Estatísticas"
	subtitulo="Resumo dos dados coletados periodicamente. Mostra como cada categoria se comporta em comissão, preço e volume de vendas."
/>

<label class="janela">
	janela:
	<select bind:value={dias} onchange={carregar} class="dado">
		<option value={7}>7 dias</option>
		<option value={30}>30 dias</option>
		<option value={90}>90 dias</option>
	</select>
</label>

<!-- ── Estatísticas por Categoria ───────────────────────────────────────── -->
<section class="secao">
	<h2>Mercado por categoria</h2>
</section>

{#if carregando}
	<Loading mensagem="Resumindo os dados…" />
{:else if erro}
	<Alert variant="error"><p><strong>Não consegui carregar as estatísticas.</strong></p><p>{erro}</p></Alert>
{:else if !dados || dados.total_amostras === 0}
	<EmptyState
		mensagem="Ainda não há dados coletados nesta janela."
		dica={dados?.fonte === 'nop'
			? 'O servidor está sem o BigQuery ligado (modo local).'
			: 'Assim que a coleta periódica rodar, o resumo por categoria aparece aqui.'}
	/>
{:else}
	<p class="meta dado">
		fonte: {dados.fonte} · {num(dados.total_amostras)} amostras · janela {dados.dias_janela} dias
	</p>
	<div class="tabela">
		<div class="cab">
			<span>categoria</span>
			<span>amostras</span>
			<span>comissão méd.</span>
			<span>comissão med.</span>
			<span>preço méd.</span>
			<span>vendas méd.</span>
			<span>teor méd.</span>
		</div>
		{#each dados.por_categoria as c (c.categoria)}
			<div class="linha">
				<span class="cat">{c.categoria || '—'}</span>
				<span class="dado">{num(c.amostras)}</span>
				<span class="dado ouro">{pct(c.comissao_media)}</span>
				<span class="dado">{pct(c.comissao_mediana)}</span>
				<span class="dado">{brl(c.preco_medio)}</span>
				<span class="dado">{num(c.vendas_media)}</span>
				<span class="dado">{(c.teor_medio ?? 0).toFixed(3)}</span>
			</div>
		{/each}
	</div>
{/if}

<!-- ── Evolução de Lojas Monitoradas ───────────────────────────────────── -->
{#if $usuario}
<section class="secao secao-lojas">
	<h2>📈 Evolução de preço — Lojas monitoradas</h2>
	<p class="sub">Acompanhe como os preços médios das lojas monitoradas evoluem ao longo do tempo.</p>

	{#if carregandoEvolucao}
		<Loading mensagem="Calculando evolução…" />
	{:else if erroEvolucao}
		<Alert variant="error"><p>{erroEvolucao}</p></Alert>
	{:else if !evolucao || evolucao.lojas?.length === 0}
		<EmptyState
			mensagem="Nenhuma loja monitorada com dados suficientes para análise."
			dica='Adicione lojas na <a href="/lojas">página de lojas</a> e aguarde pelo menos 2 coletas.'
		/>
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
	h3 { font-size: 1.1rem; margin: 0; }
	h4 { font-size: 0.85rem; margin: 0 0 var(--r2); font-weight: 600; }
	.sub { color: var(--tinta-suave); margin: 0 0 var(--r4); font-size: 0.9rem; }
	.secao { margin-bottom: var(--r8); }
	.secao-lojas { margin-top: var(--r8); }
	.janela { font-size: 0.85rem; color: var(--tinta-suave); margin-bottom: var(--r6); display: block; }
	.janela select {
		font-family: var(--mono); padding: 6px 10px; border-radius: 8px;
		border: 1px solid var(--linha); background: var(--porcelana); margin-left: 6px;
	}
	.meta { font-size: 0.8rem; color: var(--tinta-suave); margin: 0 0 var(--r4); }

	/* Tabela de estatísticas */
	.tabela {
		border: 1px solid var(--linha); border-radius: var(--raio);
		overflow: hidden; background: var(--nevoa);
	}
	.cab, .linha {
		display: grid; grid-template-columns: 1.6fr repeat(6, 1fr);
		gap: var(--r2); padding: var(--r3) var(--r4); align-items: center;
	}
	.cab {
		background: color-mix(in srgb, var(--porcelana) 70%, white);
		font-size: 0.7rem; font-weight: 600; letter-spacing: 0.04em;
		text-transform: uppercase; color: var(--tinta-suave);
		border-bottom: 1px solid var(--linha);
	}
	.linha { border-top: 1px solid var(--linha); font-size: 0.9rem; }
	.linha:first-of-type { border-top: none; }
	.cat { font-weight: 600; color: var(--rosa); }
	.ouro { color: var(--ouro); font-weight: 700; }

	/* Resumo cards */
	.resumo-cards {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
		gap: var(--r3);
		margin-bottom: var(--r6);
	}

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
