<script>
	import { onMount } from 'svelte';
	import { buscarEstatisticas, buscarEvolucaoLojas } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';

	let dias = $state(30);
	let dados = $state(null);
	let carregando = $state(true);
	let erro = $state(null);

	// Evolução de lojas
	let evolucao = $state(null);
	let carregandoEvolucao = $state(false);
	let erroEvolucao = $state(null);

	const pct = (v) => `${(v * 100).toLocaleString('pt-BR', { maximumFractionDigits: 1 })}%`;
	const brl = (v) => (v ?? 0).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
	const num = (v) => (v ?? 0).toLocaleString('pt-BR', { maximumFractionDigits: 0 });
	const pctSinal = (v) => {
		const val = (v * 100).toFixed(1);
		return v >= 0 ? `+${val}%` : `${val}%`;
	};

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

<section class="intro">
	<p class="rotulo">análise de mercado</p>
	<h1>Estatísticas</h1>
	<p class="sub">
		Resumo dos dados coletados periodicamente. Mostra como cada categoria se comporta em
		comissão, preço e volume de vendas.
	</p>
	<label class="janela">
		janela:
		<select bind:value={dias} onchange={carregar} class="dado">
			<option value={7}>7 dias</option>
			<option value={30}>30 dias</option>
			<option value={90}>90 dias</option>
		</select>
	</label>
</section>

<!-- ── Estatísticas por Categoria ───────────────────────────────────────── -->
<section class="secao">
	<h2>Mercado por categoria</h2>
</section>

{#if carregando}
	<p class="aviso">Resumindo os dados…</p>
{:else if erro}
	<div class="erro">
		<p><strong>Não consegui carregar as estatísticas.</strong></p>
		<p>{erro}</p>
	</div>
{:else if !dados || dados.total_amostras === 0}
	<div class="vazio">
		<p>Ainda não há dados coletados nesta janela.</p>
		<p class="dica">
			{#if dados?.fonte === 'nop'}
				O servidor está sem o BigQuery ligado (modo local).
			{:else}
				Assim que a coleta periódica rodar, o resumo por categoria aparece aqui.
			{/if}
		</p>
	</div>
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
		<p class="aviso">Calculando evolução…</p>
	{:else if erroEvolucao}
		<div class="erro"><p>{erroEvolucao}</p></div>
	{:else if !evolucao || evolucao.lojas?.length === 0}
		<div class="vazio">
			<p>Nenhuma loja monitorada com dados suficientes para análise.</p>
			<p class="dica">Adicione lojas na <a href="/lojas">página de lojas</a> e aguarde pelo menos 2 coletas.</p>
		</div>
	{:else}
		<!-- Resumo geral -->
		<div class="resumo-cards">
			<div class="card-stat">
				<span class="stat-label">Lojas</span>
				<span class="stat-valor">{evolucao.resumo.total_lojas}</span>
			</div>
			<div class="card-stat">
				<span class="stat-label">Produtos</span>
				<span class="stat-valor">{num(evolucao.resumo.total_produtos)}</span>
			</div>
			<div class="card-stat">
				<span class="stat-label">Preço médio</span>
				<span class="stat-valor">{brl(evolucao.resumo.preco_medio_global)}</span>
			</div>
			<div class="card-stat">
				<span class="stat-label">Variação média</span>
				<span class="stat-valor" class:positivo={evolucao.resumo.variacao_media_global_pct > 0} class:negativo={evolucao.resumo.variacao_media_global_pct < 0}>
					{pctSinal(evolucao.resumo.variacao_media_global_pct)}
				</span>
			</div>
			<div class="card-stat">
				<span class="stat-label">↓ Quedas</span>
				<span class="stat-valor verde">{evolucao.resumo.total_quedas}</span>
			</div>
			<div class="card-stat">
				<span class="stat-label">↑ Altas</span>
				<span class="stat-valor vermelho">{evolucao.resumo.total_altas}</span>
			</div>
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
	.intro { max-width: 42rem; margin-bottom: var(--r8); }
	h1 { font-size: clamp(2rem, 6vw, 3rem); margin: var(--r2) 0 var(--r4); }
	h2 { font-size: 1.3rem; margin: 0 0 var(--r2); }
	h3 { font-size: 1.1rem; margin: 0; }
	h4 { font-size: 0.85rem; margin: 0 0 var(--r2); font-weight: 600; }
	.sub { color: var(--tinta-suave); margin: 0 0 var(--r4); font-size: 0.9rem; }
	.secao { margin-bottom: var(--r8); }
	.secao-lojas { margin-top: var(--r8); }
	.janela { font-size: 0.85rem; color: var(--tinta-suave); }
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
	.aviso { color: var(--tinta-suave); font-style: italic; }
	.vazio, .erro {
		background: var(--nevoa); border: 1px solid var(--linha);
		border-radius: var(--raio); padding: var(--r8); text-align: center;
	}
	.vazio a { color: var(--ouro); text-decoration: underline; }
	.dica { color: var(--tinta-suave); font-size: 0.85rem; max-width: 50ch; margin: var(--r2) auto 0; }

	/* Resumo cards */
	.resumo-cards {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
		gap: var(--r3);
		margin-bottom: var(--r6);
	}
	.card-stat {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: 10px;
		padding: var(--r3) var(--r4);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
	}
	.stat-label { font-size: 0.72rem; text-transform: uppercase; color: var(--tinta-suave); font-weight: 600; }
	.stat-valor { font-size: 1.2rem; font-weight: 700; }
	.positivo { color: #dc2626; }
	.negativo { color: #16a34a; }
	.verde { color: #16a34a; }
	.vermelho { color: #dc2626; }

	/* Loja evolução */
	.loja-evo {
		border: 1px solid var(--linha);
		border-radius: 12px;
		padding: var(--r4);
		margin-bottom: var(--r4);
		background: white;
	}
	.loja-evo-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--r4);
		flex-wrap: wrap;
		gap: var(--r2);
	}
	.loja-evo-meta {
		display: flex;
		gap: var(--r3);
		font-size: 0.8rem;
		color: var(--tinta-suave);
	}

	/* Série temporal (mini chart de barras) */
	.serie-temporal {
		margin-bottom: var(--r4);
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.serie-labels {
		display: flex;
		justify-content: space-between;
		font-size: 0.7rem;
		color: var(--tinta-suave);
	}
	.serie-barras {
		display: flex;
		align-items: flex-end;
		gap: 2px;
		height: 60px;
		padding: 4px 0;
	}
	.barra-wrapper {
		flex: 1;
		height: 100%;
		display: flex;
		align-items: flex-end;
	}
	.barra {
		width: 100%;
		background: var(--ouro);
		border-radius: 2px 2px 0 0;
		min-height: 3px;
		transition: height 0.3s ease;
	}
	.barra-queda { background: #16a34a; }
	.barra-alta { background: #dc2626; }
	.serie-datas {
		display: flex;
		justify-content: space-between;
		font-size: 0.68rem;
		color: var(--tinta-suave);
	}

	/* Variações resumo */
	.variacoes-resumo {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--r4);
	}
	.var-grupo { }
	.var-item {
		display: flex;
		align-items: center;
		gap: var(--r2);
		font-size: 0.8rem;
		padding: 4px 0;
		border-bottom: 1px solid var(--linha);
	}
	.var-item:last-child { border-bottom: none; }
	.var-nome {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		max-width: 150px;
	}
	.badge-var {
		padding: 1px 6px;
		border-radius: 999px;
		font-size: 0.7rem;
		font-weight: 700;
		white-space: nowrap;
	}
	.badge-baixou { background: #dcfce7; color: #16a34a; }
	.badge-subiu { background: #fef2f2; color: #dc2626; }
	.var-precos { font-size: 0.72rem; color: var(--tinta-suave); white-space: nowrap; }

	@media (max-width: 720px) {
		.cab { display: none; }
		.linha { grid-template-columns: 1fr 1fr; gap: 4px; }
		.variacoes-resumo { grid-template-columns: 1fr; }
		.resumo-cards { grid-template-columns: repeat(3, 1fr); }
	}
</style>
