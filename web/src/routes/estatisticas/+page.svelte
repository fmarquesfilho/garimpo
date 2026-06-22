<script>
	import { onMount } from 'svelte';
	import { buscarEstatisticas, listarBuscasServidor } from '$lib/api.js';

	let dias = $state(30);
	let dados = $state(null);
	let buscas = $state([]);
	let carregando = $state(true);
	let erro = $state(null);

	const pct = (v) => `${(v * 100).toLocaleString('pt-BR', { maximumFractionDigits: 1 })}%`;
	const brl = (v) => (v ?? 0).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
	const num = (v) => (v ?? 0).toLocaleString('pt-BR', { maximumFractionDigits: 0 });
	const dataHora = (v) => {
		if (!v) return '—';
		const d = new Date(v);
		return d.toLocaleString('pt-BR', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' });
	};

	async function carregar() {
		carregando = true;
		erro = null;
		try {
			const [est, b] = await Promise.all([
				buscarEstatisticas({ dias }),
				listarBuscasServidor()
			]);
			dados = est;
			buscas = b?.buscas ?? [];
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}
	}
	onMount(carregar);
</script>

<section class="intro">
	<p class="rotulo">o que o mercado vem mostrando</p>
	<h1>Estatísticas</h1>
	<p class="sub">
		Resumo descritivo dos <strong>snapshots</strong> coletados e status das buscas agendadas.
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

<!-- ── Status das Buscas Agendadas ──────────────────────────────────────── -->
{#if buscas.length > 0}
	<section class="secao">
		<h2>Buscas agendadas</h2>
		<p class="sub-secao">Perfis com coleta periódica ativa. Cada um gera snapshots no BigQuery nos horários configurados.</p>
		<div class="cards-buscas">
			{#each buscas as b (b.id)}
				<div class="card-busca">
					<div class="card-busca-topo">
						<span class="card-id">{b.id}</span>
						{#if b.cron}
							<span class="badge-cron">⏱ {b.cron}</span>
						{:else}
							<span class="badge-manual">manual</span>
						{/if}
					</div>
					<div class="card-keywords">
						{#each b.keywords ?? [] as kw}
							<span class="kw-tag">{kw}</span>
						{/each}
					</div>
					<div class="card-meta">
						<span>{b.estrategia}</span>
						<span>{b.categoria}</span>
						<span>top {b.top}</span>
						<span>vendas≥{b.vendas_min}</span>
					</div>
					<p class="card-salvo">salvo em {dataHora(b.salvo_em)}</p>
				</div>
			{/each}
		</div>
	</section>
{/if}

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

<style>
	.intro { max-width: 42rem; margin-bottom: var(--r8); }
	h1 { font-size: clamp(2rem, 6vw, 3rem); margin: var(--r2) 0 var(--r4); }
	h2 { font-size: 1.3rem; margin: 0 0 var(--r2); }
	.sub { color: var(--tinta-suave); margin: 0 0 var(--r4); }
	.sub-secao { color: var(--tinta-suave); font-size: 0.85rem; margin: 0 0 var(--r4); }
	.secao { margin-bottom: var(--r8); }
	.janela { font-size: 0.85rem; color: var(--tinta-suave); }
	.janela select {
		font-family: var(--mono); padding: 6px 10px; border-radius: 8px;
		border: 1px solid var(--linha); background: var(--porcelana); margin-left: 6px;
	}
	.meta { font-size: 0.8rem; color: var(--tinta-suave); margin: 0 0 var(--r4); }

	/* Cards de buscas agendadas */
	.cards-buscas {
		display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
		gap: var(--r4);
	}
	.card-busca {
		background: var(--nevoa); border: 1px solid var(--linha);
		border-radius: var(--raio); padding: var(--r4);
		display: flex; flex-direction: column; gap: var(--r2);
	}
	.card-busca-topo {
		display: flex; justify-content: space-between; align-items: center;
	}
	.card-id { font-weight: 700; font-size: 0.95rem; }
	.badge-cron {
		font-size: 0.72rem; font-family: var(--mono);
		padding: 2px 8px; border-radius: 999px;
		background: var(--ouro-fundo); color: #7a5a1e;
	}
	.badge-manual {
		font-size: 0.72rem; padding: 2px 8px; border-radius: 999px;
		background: var(--porcelana); color: var(--tinta-suave);
		border: 1px solid var(--linha);
	}
	.card-keywords { display: flex; flex-wrap: wrap; gap: 4px; }
	.kw-tag {
		font-size: 0.8rem; font-weight: 600; padding: 2px 10px;
		border-radius: 999px; background: var(--porcelana);
		border: 1px solid var(--linha); color: var(--tinta);
	}
	.card-meta {
		display: flex; flex-wrap: wrap; gap: var(--r2);
		font-size: 0.72rem; color: var(--tinta-suave);
	}
	.card-salvo { font-size: 0.7rem; color: var(--tinta-suave); margin: 0; }

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
	.dica { color: var(--tinta-suave); font-size: 0.85rem; max-width: 50ch; margin: var(--r2) auto 0; }
	@media (max-width: 720px) {
		.cab { display: none; }
		.linha { grid-template-columns: 1fr 1fr; gap: 4px; }
	}
</style>
