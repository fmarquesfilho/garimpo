<script>
	import { onMount } from 'svelte';
	import { buscarColetas, listarBuscasServidor } from '$lib/api.js';

	let dias = $state(30);
	let coletas = $state([]);
	let buscas = $state([]);
	let carregando = $state(true);
	let erro = $state(null);

	const dataHora = (v) => {
		if (!v) return '—';
		const d = new Date(v);
		return d.toLocaleString('pt-BR', {
			day: '2-digit', month: '2-digit', year: '2-digit',
			hour: '2-digit', minute: '2-digit'
		});
	};

	const tempoAtras = (v) => {
		if (!v) return '';
		const diff = Date.now() - new Date(v).getTime();
		const min = Math.floor(diff / 60000);
		if (min < 60) return `${min}min atrás`;
		const h = Math.floor(min / 60);
		if (h < 24) return `${h}h atrás`;
		const d = Math.floor(h / 24);
		return `${d}d atrás`;
	};

	async function carregar() {
		carregando = true;
		erro = null;
		try {
			const [c, b] = await Promise.all([
				buscarColetas({ dias }),
				listarBuscasServidor()
			]);
			coletas = c?.coletas ?? [];
			buscas = b?.buscas ?? [];
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}
	}

	onMount(carregar);

	// Agrupa coletas por keyword para mostrar resumo
	let resumoPorKeyword = $derived(() => {
		const mapa = new Map();
		for (const c of coletas) {
			const key = c.keyword || '(sem keyword)';
			if (!mapa.has(key)) {
				mapa.set(key, { keyword: key, categoria: c.categoria, total: 0, ultimaColeta: c.coletado_em, produtosTotal: 0 });
			}
			const r = mapa.get(key);
			r.total++;
			r.produtosTotal += c.produtos;
			if (new Date(c.coletado_em) > new Date(r.ultimaColeta)) {
				r.ultimaColeta = c.coletado_em;
			}
		}
		return [...mapa.values()].sort((a, b) => new Date(b.ultimaColeta) - new Date(a.ultimaColeta));
	});
</script>

<section class="intro">
	<p class="rotulo">monitoramento</p>
	<h1>Coletas</h1>
	<p class="sub">
		Histórico de coletas executadas e status das buscas agendadas.
		Cada coleta grava um snapshot dos top produtos do momento.
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

{#if carregando}
	<p class="aviso">Carregando histórico…</p>
{:else if erro}
	<div class="erro">
		<p><strong>Não consegui carregar as coletas.</strong></p>
		<p>{erro}</p>
	</div>
{:else}
	<!-- Buscas agendadas -->
	{#if buscas.length > 0}
		<section class="secao">
			<h2>Buscas agendadas</h2>
			<div class="cards-agendadas">
				{#each buscas as b (b.id)}
					<div class="card-agendada">
						<div class="card-topo">
							<span class="card-id">{b.id}</span>
							{#if b.cron}
								<span class="badge-cron">⏱ {b.cron}</span>
							{:else}
								<span class="badge-manual">manual</span>
							{/if}
						</div>
						<div class="card-kws">
							{#each b.keywords ?? [] as kw}
								<span class="kw-tag">{kw}</span>
							{/each}
						</div>
						<div class="card-info">
							<span>{b.estrategia}</span> · <span>{b.categoria}</span> · <span>top {b.top}</span>
						</div>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	<!-- Resumo por keyword -->
	{#if resumoPorKeyword().length > 0}
		<section class="secao">
			<h2>Resumo por keyword</h2>
			<p class="sub-secao">Quantas coletas cada keyword teve na janela de {dias} dias.</p>
			<div class="tabela">
				<div class="cab">
					<span>keyword</span>
					<span>categoria</span>
					<span>coletas</span>
					<span>produtos</span>
					<span>última coleta</span>
				</div>
				{#each resumoPorKeyword() as r (r.keyword)}
					<div class="linha">
						<span class="kw-cell">{r.keyword}</span>
						<span class="dado">{r.categoria || '—'}</span>
						<span class="dado">{r.total}</span>
						<span class="dado">{r.produtosTotal}</span>
						<span class="dado tempo">{tempoAtras(r.ultimaColeta)}</span>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	<!-- Histórico detalhado -->
	<section class="secao">
		<h2>Histórico de execuções</h2>
		{#if coletas.length === 0}
			<div class="vazio">
				<p>Nenhuma coleta encontrada nesta janela.</p>
				<p class="dica">As coletas acontecem nos horários agendados (cron) ou quando disparadas manualmente.</p>
			</div>
		{:else}
			<p class="sub-secao">{coletas.length} execuções nos últimos {dias} dias.</p>
			<div class="tabela">
				<div class="cab">
					<span>quando</span>
					<span>keyword</span>
					<span>categoria</span>
					<span>estratégia</span>
					<span>produtos</span>
				</div>
				{#each coletas as c, i (i)}
					<div class="linha">
						<span class="dado tempo">{dataHora(c.coletado_em)}</span>
						<span class="kw-cell">{c.keyword || '—'}</span>
						<span class="dado">{c.categoria || '—'}</span>
						<span class="dado">{c.estrategia}</span>
						<span class="dado produtos">{c.produtos}</span>
					</div>
				{/each}
			</div>
		{/if}
	</section>
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

	/* Cards agendadas */
	.cards-agendadas {
		display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
		gap: var(--r4); margin-top: var(--r3);
	}
	.card-agendada {
		background: var(--nevoa); border: 1px solid var(--linha);
		border-radius: var(--raio); padding: var(--r3) var(--r4);
		display: flex; flex-direction: column; gap: var(--r2);
	}
	.card-topo { display: flex; justify-content: space-between; align-items: center; }
	.card-id { font-weight: 700; font-size: 0.92rem; }
	.badge-cron {
		font-size: 0.7rem; font-family: var(--mono);
		padding: 2px 8px; border-radius: 999px;
		background: var(--ouro-fundo); color: #7a5a1e;
	}
	.badge-manual {
		font-size: 0.7rem; padding: 2px 8px; border-radius: 999px;
		background: var(--porcelana); color: var(--tinta-suave); border: 1px solid var(--linha);
	}
	.card-kws { display: flex; flex-wrap: wrap; gap: 4px; }
	.kw-tag {
		font-size: 0.78rem; font-weight: 600; padding: 2px 8px;
		border-radius: 999px; background: var(--porcelana);
		border: 1px solid var(--linha); color: var(--tinta);
	}
	.card-info { font-size: 0.75rem; color: var(--tinta-suave); }

	/* Tabela */
	.tabela {
		border: 1px solid var(--linha); border-radius: var(--raio);
		overflow: hidden; background: var(--nevoa);
	}
	.cab, .linha {
		display: grid; grid-template-columns: 1.4fr 1.2fr 1fr 1fr 0.8fr;
		gap: var(--r2); padding: var(--r3) var(--r4); align-items: center;
	}
	.cab {
		background: color-mix(in srgb, var(--porcelana) 70%, white);
		font-size: 0.7rem; font-weight: 600; letter-spacing: 0.04em;
		text-transform: uppercase; color: var(--tinta-suave);
		border-bottom: 1px solid var(--linha);
	}
	.linha { border-top: 1px solid var(--linha); font-size: 0.88rem; }
	.linha:first-of-type { border-top: none; }
	.kw-cell { font-weight: 600; color: var(--rosa); }
	.tempo { font-family: var(--mono); font-size: 0.8rem; }
	.produtos { font-weight: 700; color: var(--ouro); }
	.aviso { color: var(--tinta-suave); font-style: italic; }
	.vazio, .erro {
		background: var(--nevoa); border: 1px solid var(--linha);
		border-radius: var(--raio); padding: var(--r8); text-align: center;
	}
	.dica { color: var(--tinta-suave); font-size: 0.85rem; }
	@media (max-width: 720px) {
		.cab { display: none; }
		.linha { grid-template-columns: 1fr 1fr; gap: 4px; }
	}
</style>
