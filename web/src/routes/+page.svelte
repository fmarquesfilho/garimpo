<script>
	import { onMount } from 'svelte';
	import { buscarCandidatos, compararEstrategias, registrarSelecao } from '$lib/api.js';
	import { goto } from '$app/navigation';
	import { quadro } from '$lib/board.js';
	import { filtros as filtrosStore } from '$lib/filtros.js';
	import { buscasSalvas, slugificar } from '$lib/buscas.js';
	import { get } from 'svelte/store';
	import CandidateCard from '$lib/components/CandidateCard.svelte';
	import StrategyToggle from '$lib/components/StrategyToggle.svelte';
	import FilterBar from '$lib/components/FilterBar.svelte';
	import TagInput from '$lib/components/TagInput.svelte';
	import BuscaCard from '$lib/components/BuscaCard.svelte';
	import AgendadorBusca from '$lib/components/AgendadorBusca.svelte';

	// ── Estado dos filtros ────────────────────────────────────────────────────
	let f = $state(get(filtrosStore));
	$effect(() => { filtrosStore.set({ ...f }); });

	let carregando = $state(true);
	let erro = $state(null);
	let lista = $state([]);
	let pares = $state(null);
	let fonteAtiva = $state('');

	onMount(async () => {
		await buscasSalvas.sincronizarDoServidor();
	});

	// ── Nova busca ────────────────────────────────────────────────────────────
	let mostrarFormBusca = $state(false);
	let buscasColapsadas = $state(false);
	let keywordsNovas = $state([]);
	let shopIdsNovas = $state([]);
	let cronNova = $state('');
	let estrategiaNova = $state('nicho');

	function parseShopId(raw) {
		const match = raw.match(/(\d{5,})/);
		const id = match ? match[1] : raw;
		return /^\d+$/.test(id) ? id : null;
	}

	function salvarBuscaNova() {
		const kws = keywordsNovas.length > 0 ? keywordsNovas : (f.busca.trim() ? [f.busca.trim()] : []);
		const shops = shopIdsNovas.map(Number).filter(Boolean);
		if (kws.length === 0 && shops.length === 0) return;

		buscasSalvas.salvar({
			id: slugificar(kws[0] ?? `loja-${shops[0]}`),
			keywords: kws,
			shop_ids: shops.length > 0 ? shops : undefined,
			categoria: f.categoria,
			estrategia: estrategiaNova === 'comparar' ? 'ambas' : estrategiaNova,
			comissao_min: f.comissaoMin,
			vendas_min: f.vendasMin,
			nota_min: f.notaMin,
			top: f.quantos,
			cron: cronNova
		});
		keywordsNovas = [];
		shopIdsNovas = [];
		cronNova = '';
		mostrarFormBusca = false;
	}

	// ── Aplicar busca salva ───────────────────────────────────────────────────
	function aplicarBusca(b) {
		const keyword = (b.keywords ?? [])[0] ?? '';
		let modo = b.estrategia ?? 'nicho';
		if (modo === 'ambas') modo = 'comparar';
		if (!['nicho', 'diversificada', 'comparar'].includes(modo)) modo = 'nicho';
		f = {
			...f, busca: keyword, categoria: b.categoria ?? f.categoria, modo,
			comissaoMin: b.comissao_min ?? f.comissaoMin,
			vendasMin: b.vendas_min ?? f.vendasMin,
			notaMin: b.nota_min ?? f.notaMin,
			quantos: b.top ?? f.quantos
		};
	}

	function proximaKeyword(b) {
		if (!b.keywords || b.keywords.length <= 1) return;
		const idx = (b.keywords.indexOf(f.busca) + 1) % b.keywords.length;
		f = { ...f, busca: b.keywords[idx] };
	}

	// ── Carregar candidatos ───────────────────────────────────────────────────
	async function carregar() {
		// Não busca sem keyword (mostra estado vazio)
		if (!f.busca.trim()) {
			lista = [];
			pares = null;
			carregando = false;
			return;
		}
		carregando = true;
		erro = null;
		pares = null;
		const filtrosReq = {
			keyword: f.busca.trim(),
			categoria: f.categoria,
			comissaoMin: f.comissaoMin,
			vendasMin: f.vendasMin,
			notaMin: f.notaMin,
			exploracao: f.explorar ? 0.2 : 0
		};
		try {
			if (f.modo === 'comparar') {
				const r = await compararEstrategias({ top: 6, ...filtrosReq });
				pares = r;
				fonteAtiva = r.fonte ?? '';
			} else {
				const r = await buscarCandidatos({ estrategia: f.modo, top: f.quantos, ...filtrosReq });
				lista = (r.candidatos ?? []).map((c) => ({ ...c, estrategia: f.modo }));
				fonteAtiva = r.fonte ?? '';
			}
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}
	}

	let timer;
	$effect(() => {
		f.modo; f.busca; f.categoria; f.comissaoMin; f.quantos; f.vendasMin; f.notaMin; f.explorar;
		clearTimeout(timer);
		timer = setTimeout(carregar, 350);
		return () => clearTimeout(timer);
	});

	// ── Ações ─────────────────────────────────────────────────────────────────
	function selecionar(c) {
		quadro.selecionar(c);
		registrarSelecao(c);
	}

	function publicarOferta(c) {
		goto(`/publicar?dados=${encodeURIComponent(JSON.stringify(c))}`);
	}
</script>

<!-- ── Intro ───────────────────────────────────────────────────────────────── -->
<section class="intro">
	<p class="rotulo">a peneira do dia</p>
	<h1>O que vale a pena garimpar hoje</h1>
	<p class="sub">
		Produtos elegíveis ordenados pelo <strong>teor</strong> — o quanto cada um rende pelo esforço.
	</p>
</section>

<!-- ── Controles ───────────────────────────────────────────────────────────── -->
<div class="controles">
	<StrategyToggle bind:valor={f.modo} />
</div>

<FilterBar
	bind:busca={f.busca}
	bind:categoria={f.categoria}
	bind:comissaoMin={f.comissaoMin}
	bind:vendasMin={f.vendasMin}
	bind:notaMin={f.notaMin}
	bind:quantos={f.quantos}
	bind:explorar={f.explorar}
	modo={f.modo}
/>

<!-- ── Buscas Salvas ───────────────────────────────────────────────────────── -->
<section class="buscas">
	<div class="buscas-cabecalho">
		<button class="buscas-titulo-btn" onclick={() => (buscasColapsadas = !buscasColapsadas)} type="button">
			<span class="seta" class:girada={!buscasColapsadas}>▸</span>
			Buscas salvas
			{#if $buscasSalvas.length > 0}<span class="badge-contagem">{$buscasSalvas.length}</span>{/if}
		</button>
		{#if !buscasColapsadas}
			<button class="btn-nova" onclick={() => (mostrarFormBusca = !mostrarFormBusca)} type="button">
				{mostrarFormBusca ? '✕ cancelar' : '+ nova busca'}
			</button>
		{/if}
	</div>

	{#if mostrarFormBusca && !buscasColapsadas}
		<div class="form-nova-busca">
			<div class="form-linha">
				<div class="flex1">
					<TagInput bind:tags={keywordsNovas} label="palavras-chave" placeholder="ex.: kenzo, shiseido…" />
				</div>
				<label class="campo-estrategia">
					<span class="rotulo">estratégia</span>
					<select bind:value={estrategiaNova} class="dado">
						<option value="nicho">Nicho</option>
						<option value="diversificada">Diversificada</option>
						<option value="ambas">Comparar ambas</option>
					</select>
				</label>
			</div>

			{#if keywordsNovas.length === 0 && f.busca.trim() && shopIdsNovas.length === 0}
				<p class="dica-kw">A busca atual "<strong>{f.busca.trim()}</strong>" será usada se não adicionar keywords.</p>
			{/if}

			<TagInput bind:tags={shopIdsNovas} label="🏪 lojas shopee (ID ou URL)" placeholder="ex.: 12345678 ou shopee.com.br/shop/12345678" variant="shop" parse={parseShopId} />

			<AgendadorBusca bind:value={cronNova} />

			<div class="form-acoes">
				<button class="salvar" onclick={salvarBuscaNova} disabled={keywordsNovas.length === 0 && !f.busca.trim() && shopIdsNovas.length === 0} type="button">
					Salvar busca
				</button>
			</div>
		</div>
	{/if}

	{#if $buscasSalvas.length > 0 && !buscasColapsadas}
		<div class="buscas-lista">
			{#each $buscasSalvas as b (b.id)}
				<BuscaCard busca={b} buscaAtiva={f.busca} onaplicar={aplicarBusca} onproximakw={proximaKeyword} onremover={(id) => buscasSalvas.remover(id)} />
			{/each}
		</div>
	{:else if !mostrarFormBusca && !buscasColapsadas}
		<p class="buscas-vazia">Nenhuma busca salva. Clique em "+ nova busca" para criar.</p>
	{/if}
</section>

<!-- ── Resultados ──────────────────────────────────────────────────────────── -->
{#if carregando}
	<p class="aviso">Garimpando os melhores produtos…</p>
{:else if erro}
	<div class="msg-erro">
		<p><strong>Não consegui falar com a API.</strong></p>
		<p>{erro}</p>
		<p class="dica">Confira se o servidor está rodando: <code>go run ./cmd/garimpei-api</code></p>
	</div>
{:else if f.modo === 'comparar' && pares}
	<div class="comparacao">
		<div class="coluna">
			<h2 class="tit-col rosa">Nicho</h2>
			<div class="empilhado">
				{#each pares.nicho as c, i (c.id)}
					<CandidateCard candidato={{ ...c, estrategia: 'nicho' }} posicao={i + 1} onselecionar={selecionar} onpublicar={publicarOferta} />
				{/each}
			</div>
		</div>
		<div class="coluna">
			<h2 class="tit-col ardosia">Diversificada</h2>
			<div class="empilhado">
				{#each pares.diversificada as c, i (c.id)}
					<CandidateCard candidato={{ ...c, estrategia: 'diversificada' }} posicao={i + 1} onselecionar={selecionar} onpublicar={publicarOferta} />
				{/each}
			</div>
		</div>
	</div>
{:else if lista.length === 0}
	<div class="vazio">
		{#if f.busca.trim() === ''}
			<p>🔍 Digite um termo de busca para encontrar produtos.</p>
			<p class="dica">Ex: "skincare", "perfume", "maquiagem" — ou use uma busca salva abaixo.</p>
		{:else}
			<p>Nada na peneira para "{f.busca}".</p>
			<p class="dica">Tente outro termo, ou afrouxe os pisos de comissão, vendas e nota.</p>
		{/if}
	</div>
{:else}
	<div class="grade">
		{#each lista as c, i (c.id)}
			<CandidateCard candidato={c} posicao={i + 1} destaque={i === 0} onselecionar={selecionar} onpublicar={publicarOferta} />
		{/each}
	</div>
{/if}

<style>
	/* ── Layout ─────────────────────────────────────────────────────────── */
	.intro { max-width: 40rem; margin-bottom: var(--r8); }
	h1 { font-size: clamp(2rem, 6vw, 3.2rem); margin: var(--r2) 0 var(--r4); }
	.sub { color: var(--tinta-suave); font-size: 1.05rem; margin: 0; }
	.controles { display: flex; flex-wrap: wrap; align-items: flex-end; gap: var(--r4); margin-bottom: var(--r4); }

	/* ── Resultados ─────────────────────────────────────────────────────── */
	.grade {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: var(--r6);
	}
	.grade :global(.destaque) { grid-column: 1 / -1; }
	@media (min-width: 720px) { .grade :global(.destaque) { grid-column: span 2; } }

	.comparacao { display: grid; grid-template-columns: 1fr; gap: var(--r8); }
	@media (min-width: 800px) { .comparacao { grid-template-columns: 1fr 1fr; } }
	.empilhado { display: flex; flex-direction: column; gap: var(--r4); }
	.tit-col { font-size: 1.3rem; margin-bottom: var(--r4); padding-bottom: var(--r2); border-bottom: 2px solid; }
	.tit-col.rosa { color: var(--rosa); }
	.tit-col.ardosia { color: var(--ardosia); }

	.aviso { color: var(--tinta-suave); font-style: italic; }
	.vazio, .msg-erro {
		background: var(--nevoa); border: 1px solid var(--linha);
		border-radius: var(--raio); padding: var(--r8); text-align: center;
	}
	.msg-erro { border-color: color-mix(in srgb, var(--alerta) 30%, var(--linha)); }
	.msg-erro p { margin: var(--r2) 0; }
	.dica { color: var(--tinta-suave); font-size: 0.85rem; }
	code { background: var(--ouro-fundo); padding: 2px 6px; border-radius: 6px; }

	/* ── Buscas Salvas ──────────────────────────────────────────────────── */
	.buscas { margin: 0 0 var(--r6); }
	.buscas-cabecalho {
		display: flex; align-items: center; justify-content: space-between;
		margin-bottom: var(--r3);
	}
	.buscas-titulo-btn {
		border: none; background: transparent; cursor: pointer;
		font-weight: 700; font-size: 0.9rem; color: var(--tinta-suave);
		display: flex; align-items: center; gap: 6px; padding: 0;
	}
	.buscas-titulo-btn:hover { color: var(--tinta); }
	.seta { display: inline-block; transition: transform 0.15s ease; font-size: 0.8rem; }
	.seta.girada { transform: rotate(90deg); }
	.badge-contagem {
		font-size: 0.7rem; background: var(--ouro-fundo); color: #7a5a1e;
		padding: 1px 6px; border-radius: var(--raio-full); font-weight: 700;
	}
	.btn-nova {
		border: 1px solid var(--linha); background: var(--porcelana);
		color: var(--tinta); font-size: 0.82rem; font-weight: 600;
		padding: 6px 14px; border-radius: var(--raio-full); cursor: pointer;
	}
	.btn-nova:hover { border-color: var(--ouro); color: var(--ouro); }

	/* ── Form nova busca ────────────────────────────────────────────────── */
	.form-nova-busca {
		background: var(--nevoa); border: 1px solid var(--linha);
		border-radius: var(--raio); padding: var(--r4);
		display: flex; flex-direction: column; gap: var(--r4);
		margin-bottom: var(--r4);
	}
	.form-linha { display: flex; flex-wrap: wrap; gap: var(--r4); align-items: flex-end; }
	.flex1 { flex: 1 1 240px; }
	.campo-estrategia { display: flex; flex-direction: column; gap: 5px; }
	.campo-estrategia select {
		font-family: var(--mono); font-size: 0.9rem; padding: 9px 12px;
		border-radius: var(--raio-sm); border: 1px solid var(--linha);
		background: var(--porcelana); color: var(--tinta);
	}
	.dica-kw { font-size: 0.82rem; color: var(--tinta-suave); margin: 0; }
	.form-acoes { display: flex; justify-content: flex-end; }
	.salvar {
		border: 1px solid var(--linha); background: var(--ouro-fundo);
		color: #7a5a1e; font-weight: 600; font-size: 0.85rem;
		padding: 9px 18px; border-radius: var(--raio-sm); cursor: pointer;
	}
	.salvar:disabled { opacity: 0.5; cursor: not-allowed; }

	.buscas-lista { display: flex; flex-direction: column; gap: var(--r3); }
	.buscas-vazia { font-size: 0.85rem; color: var(--tinta-suave); font-style: italic; margin: var(--r3) 0; }
</style>
