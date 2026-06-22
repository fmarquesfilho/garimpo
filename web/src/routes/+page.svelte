<script>
	import { onMount } from 'svelte';
	import { buscarCandidatos, compararEstrategias, registrarSelecao, publicar } from '$lib/api.js';
	import { quadro } from '$lib/board.js';
	import { filtros as filtrosStore } from '$lib/filtros.js';
	import { buscasSalvas, slugificar } from '$lib/buscas.js';
	import { get } from 'svelte/store';
	import CandidateCard from '$lib/components/CandidateCard.svelte';
	import StrategyToggle from '$lib/components/StrategyToggle.svelte';
	import AgendadorBusca from '$lib/components/AgendadorBusca.svelte';

	let f = $state(get(filtrosStore));
	$effect(() => {
		filtrosStore.set({ ...f });
	});

	let carregando = $state(true);
	let erro = $state(null);
	let lista = $state([]);
	let pares = $state(null);
	let fonteAtiva = $state('');

	onMount(() => buscasSalvas.sincronizarDoServidor());

	// ── salvar busca ──────────────────────────────────────────────────────────
	let novaKeyword = $state('');     // keyword sendo adicionada a uma nova busca
	let cronNova = $state('');        // agendamento da nova busca
	let estrategiaNova = $state('nicho'); // estratégia da nova busca
	let mostrarFormBusca = $state(false);

	// keywords que estão sendo montadas antes de salvar
	let keywordsNovas = $state([]);

	function adicionarKeywordNova() {
		const kw = novaKeyword.trim();
		if (!kw || keywordsNovas.includes(kw)) return;
		keywordsNovas = [...keywordsNovas, kw];
		novaKeyword = '';
	}

	function removerKeywordNova(kw) {
		keywordsNovas = keywordsNovas.filter((k) => k !== kw);
	}

	function salvarBuscaNova() {
		// se não adicionou explicitamente, usa o termo atual da busca
		const kws = keywordsNovas.length > 0 ? keywordsNovas : (f.busca.trim() ? [f.busca.trim()] : []);
		if (kws.length === 0) return;
		buscasSalvas.salvar({
			id: slugificar(kws[0]),
			keywords: kws,
			categoria: f.categoria,
			estrategia: estrategiaNova === 'comparar' ? 'ambas' : estrategiaNova,
			comissao_min: f.comissaoMin,
			vendas_min: f.vendasMin,
			nota_min: f.notaMin,
			top: f.quantos,
			cron: cronNova
		});
		keywordsNovas = [];
		novaKeyword = '';
		cronNova = '';
		mostrarFormBusca = false;
	}

	// ── aplicar busca salva ───────────────────────────────────────────────────
	// Uma busca pode ter várias keywords; ao aplicar, usa a primeira como busca
	// ativa e muda a estratégia corretamente.
	function aplicarBusca(b) {
		const keyword = (b.keywords ?? [])[0] ?? '';
		// mapeia estrategia do servidor para o modo do frontend
		let modo = b.estrategia ?? 'nicho';
		if (modo === 'ambas') modo = 'comparar';
		if (modo !== 'nicho' && modo !== 'diversificada' && modo !== 'comparar') modo = 'nicho';
		f = {
			...f,
			busca: keyword,
			categoria: b.categoria ?? f.categoria,
			modo,
			comissaoMin: b.comissao_min ?? f.comissaoMin,
			vendasMin: b.vendas_min ?? f.vendasMin,
			notaMin: b.nota_min ?? f.notaMin,
			quantos: b.top ?? f.quantos
		};
	}

	// se a busca tem mais de uma keyword, percorre todas em sequência
	let keywordAtual = $state(0);
	function proximaKeyword(b) {
		if (!b.keywords || b.keywords.length <= 1) return;
		keywordAtual = (keywordAtual + 1) % b.keywords.length;
		f = { ...f, busca: b.keywords[keywordAtual] };
	}

	// ── carregar candidatos ───────────────────────────────────────────────────
	async function carregar() {
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

	function selecionar(c) {
		quadro.selecionar(c);
		registrarSelecao(c);
	}

	let aviso = $state(null);
	let publicando = $state(false);
	async function publicarOferta(c) {
		publicando = true;
		aviso = null;
		try {
			const r = await publicar(c);
			aviso = { ok: true, ...r };
		} catch (e) {
			aviso = { ok: false, erro: e.message };
		} finally {
			publicando = false;
		}
	}
</script>

<section class="intro">
	<p class="rotulo">a peneira do dia</p>
	<h1>O que vale a pena garimpar hoje</h1>
	<p class="sub">
		Produtos elegíveis (comissão no piso ou acima), ordenados pelo <strong>teor</strong> — o quanto
		cada um rende pelo esforço. Escolha a estratégia e mande os melhores pro quadro.
	</p>
</section>

<div class="controles">
	<StrategyToggle bind:valor={f.modo} />
</div>

<div class="filtros">
	<label class="campo busca">
		<span class="rotulo">buscar na shopee</span>
		<input type="search" bind:value={f.busca} placeholder="perfume, sérum, batom…" class="entrada" />
	</label>
	<label class="campo">
		<span class="rotulo">categoria</span>
		<select bind:value={f.categoria} class="dado">
			<option value="cosméticos">cosméticos</option>
			<option value="perfumaria">perfumaria</option>
			<option value="bem-estar">bem-estar</option>
			<option value="">sem rótulo</option>
		</select>
	</label>
	<label class="campo">
		<span class="rotulo">comissão mín.</span>
		<select bind:value={f.comissaoMin} class="dado">
			<option value={0.05}>5%</option>
			<option value={0.07}>7%</option>
			<option value={0.1}>10%</option>
			<option value={0.15}>15%</option>
		</select>
	</label>
	<label class="campo">
		<span class="rotulo">vendas mín.</span>
		<input type="number" min="0" step="1" bind:value={f.vendasMin} class="entrada num" />
	</label>
	<label class="campo">
		<span class="rotulo">nota mín.</span>
		<select bind:value={f.notaMin} class="dado">
			<option value={0}>todas</option>
			<option value={4}>4,0+</option>
			<option value={4.5}>4,5+</option>
		</select>
	</label>
	{#if f.modo !== 'comparar'}
		<label class="campo">
			<span class="rotulo">quantos</span>
			<select bind:value={f.quantos} class="dado">
				<option value={6}>6</option>
				<option value={9}>9</option>
				<option value={12}>12</option>
			</select>
		</label>
		<label class="campo-check" title="Reserva ~20% das vagas para produtos fora do topo, para descobrir o que converte">
			<input type="checkbox" bind:checked={f.explorar} />
			<span class="rotulo">explorar</span>
		</label>
	{/if}
</div>

<!-- ── Buscas Salvas ─────────────────────────────────────────────────────── -->
<div class="buscas">
	<div class="buscas-cabecalho">
		<span class="buscas-titulo">Buscas salvas</span>
		<button class="btn-nova" onclick={() => (mostrarFormBusca = !mostrarFormBusca)} type="button">
			{mostrarFormBusca ? '✕ cancelar' : '+ nova busca'}
		</button>
	</div>

	{#if mostrarFormBusca}
		<div class="form-nova-busca">
			<div class="form-linha">
				<label class="campo flex1">
					<span class="rotulo">palavras-chave</span>
					<div class="entrada-kw">
						<input
							class="entrada"
							bind:value={novaKeyword}
							placeholder="ex.: kenzo, shiseido, perfume oriental…"
							onkeydown={(e) => e.key === 'Enter' && adicionarKeywordNova()}
						/>
						<button type="button" class="btn-add-kw" onclick={adicionarKeywordNova}>+</button>
					</div>
				</label>
				<label class="campo">
					<span class="rotulo">estratégia</span>
					<select bind:value={estrategiaNova} class="dado">
						<option value="nicho">Nicho</option>
						<option value="diversificada">Diversificada</option>
						<option value="ambas">Comparar ambas</option>
					</select>
				</label>
			</div>
			{#if keywordsNovas.length > 0}
				<div class="kws-montadas">
					{#each keywordsNovas as kw}
						<span class="pilula-kw">
							{kw}
							<button type="button" class="x" onclick={() => removerKeywordNova(kw)}>✕</button>
						</span>
					{/each}
				</div>
			{:else if f.busca.trim()}
				<p class="dica-kw">A busca atual "<strong>{f.busca.trim()}</strong>" será usada se não adicionar keywords.</p>
			{/if}
			<AgendadorBusca bind:value={cronNova} />
			<div class="form-acoes">
				<button
					class="salvar"
					onclick={salvarBuscaNova}
					disabled={keywordsNovas.length === 0 && !f.busca.trim()}
					type="button"
				>Salvar busca</button>
			</div>
		</div>
	{/if}

	{#if $buscasSalvas.length > 0}
		<div class="buscas-lista">
			{#each $buscasSalvas as b (b.id)}
				<div class="cartao-busca">
					<div class="cartao-topo">
						<div class="cartao-kws">
							{#each b.keywords as kw, i}
								<button
									type="button"
									class="kw-btn"
									class:kw-ativa={f.busca === kw}
									onclick={() => aplicarBusca({ ...b, keywords: b.keywords.slice(i).concat(b.keywords.slice(0, i)) })}
									title="Aplicar filtros desta busca com a keyword '{kw}'"
								>{kw}</button>
							{/each}
						</div>
						<div class="cartao-acoes">
							{#if b.keywords.length > 1}
								<button type="button" class="btn-prox" onclick={() => proximaKeyword(b)} title="Próxima keyword">→</button>
							{/if}
							<button type="button" class="x" onclick={() => buscasSalvas.remover(b.id)} aria-label="remover busca">✕</button>
						</div>
					</div>
					<div class="cartao-meta">
						<span class="badge dado">{b.estrategia ?? 'nicho'}</span>
						{#if b.cron}
							<span class="badge cron dado" title="coleta periódica">⏱ {b.cron}</span>
						{/if}
						<span class="badge dado">{b.categoria}</span>
					</div>
				</div>
			{/each}
		</div>
	{:else if !mostrarFormBusca}
		<p class="buscas-vazia">Nenhuma busca salva ainda. Clique em "+ nova busca" para criar.</p>
	{/if}
</div>

<p class="nota-curadoria">
	Comissão alta com zero venda costuma ser produto-fantasma. O piso de vendas e a nota mínima
	deixam na peneira só o que já tem tração.{#if fonteAtiva}<span class="fonte dado"> · fonte: {fonteAtiva}</span>{/if}
</p>

{#if aviso}
	<div class="publicacao" class:falha={!aviso.ok} role="status">
		<button class="fechar" onclick={() => (aviso = null)} aria-label="fechar">✕</button>
		{#if aviso.ok}
			<p class="cab">Publicado no canal <strong>{aviso.canal}</strong> · <span class="dado">{aviso.detalhe}</span></p>
			<pre class="msg">{aviso.mensagem}</pre>
			{#if aviso.sub_id}
				<p class="subid dado">atribuição: {aviso.sub_id}</p>
			{/if}
		{:else}
			<p class="cab">Não consegui publicar: {aviso.erro}</p>
		{/if}
	</div>
{/if}

{#if carregando}
	<p class="aviso">Garimpando os melhores produtos…</p>
{:else if erro}
	<div class="erro">
		<p><strong>Não consegui falar com a API.</strong></p>
		<p>{erro}</p>
		<p class="dica dado">Confira se o servidor está rodando: <code>go run ./cmd/garimpo-api</code> (porta 8080).</p>
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
			<p>Comece por uma busca.</p>
			<p class="dica">Digite o que quer divulgar — perfume, sérum, batom — pra peneirar a Shopee.</p>
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
	.publicacao {
		position: relative;
		background: color-mix(in srgb, var(--rosa) 8%, var(--nevoa));
		border: 1px solid color-mix(in srgb, var(--rosa) 30%, var(--linha));
		border-left: 3px solid var(--rosa);
		border-radius: var(--raio);
		padding: var(--r4) var(--r6);
		margin-bottom: var(--r6);
	}
	.publicacao.falha {
		background: color-mix(in srgb, var(--alerta) 8%, var(--nevoa));
		border-color: color-mix(in srgb, var(--alerta) 30%, var(--linha));
		border-left-color: var(--alerta);
	}
	.publicacao .cab { margin: 0 var(--r6) 0 0; font-size: 0.92rem; }
	.publicacao .msg {
		margin: var(--r3) 0 0;
		padding: var(--r3);
		background: var(--porcelana);
		border-radius: 8px;
		font-family: var(--ui);
		font-size: 0.92rem;
		white-space: pre-wrap;
		word-break: break-word;
	}
	.fechar {
		position: absolute; top: var(--r3); right: var(--r3);
		border: none; background: transparent; font-size: 0.9rem;
		color: var(--tinta-suave); cursor: pointer;
	}
	.intro { max-width: 40rem; margin-bottom: var(--r8); }
	h1 { font-size: clamp(2rem, 6vw, 3.2rem); margin: var(--r2) 0 var(--r4); }
	.sub { color: var(--tinta-suave); font-size: 1.05rem; margin: 0; }
	.controles { display: flex; flex-wrap: wrap; align-items: flex-end; gap: var(--r4); margin-bottom: var(--r4); }
	.filtros {
		display: flex; flex-wrap: wrap; align-items: flex-end; gap: var(--r4);
		margin-bottom: var(--r4); padding: var(--r4);
		background: var(--nevoa); border: 1px solid var(--linha); border-radius: var(--raio);
	}
	.campo { display: flex; flex-direction: column; gap: 5px; }
	.campo-check {
		display: flex; align-items: center; gap: 6px;
		align-self: flex-end; padding-bottom: 9px; cursor: pointer;
	}
	.campo-check input { width: 16px; height: 16px; accent-color: var(--ouro); cursor: pointer; }
	.subid { margin: var(--r2) 0 0; font-size: 0.78rem; color: var(--tinta-suave); }
	.campo.busca { flex: 1 1 220px; }
	.entrada {
		font-family: var(--ui); font-size: 0.95rem; padding: 9px 12px;
		border-radius: 10px; border: 1px solid var(--linha);
		background: var(--porcelana); color: var(--tinta); width: 100%;
	}
	.entrada::placeholder { color: var(--tinta-suave); opacity: 0.7; }
	.entrada.num { font-family: var(--mono); width: 5.5rem; }
	.fonte { opacity: 0.8; }
	.nota-curadoria {
		margin: 0 0 var(--r8); font-size: 0.85rem; color: var(--tinta-suave);
		max-width: 60ch; border-left: 2px solid var(--ouro-claro); padding-left: var(--r3);
	}
	select {
		font-family: var(--mono); font-size: 0.9rem; padding: 9px 12px;
		border-radius: 10px; border: 1px solid var(--linha);
		background: var(--porcelana); color: var(--tinta);
	}
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
	.vazio, .erro {
		background: var(--nevoa); border: 1px solid var(--linha);
		border-radius: var(--raio); padding: var(--r8); text-align: center;
	}
	.erro { border-color: color-mix(in srgb, var(--alerta) 30%, var(--linha)); }
	.erro p { margin: var(--r2) 0; }
	.dica { color: var(--tinta-suave); font-size: 0.85rem; }
	code { background: var(--ouro-fundo); padding: 2px 6px; border-radius: 6px; }

	/* ── Buscas Salvas ─────────────────────────────────────────────────────── */
	.buscas { margin: 0 0 var(--r6); }
	.buscas-cabecalho {
		display: flex; align-items: center; justify-content: space-between;
		margin-bottom: var(--r3);
	}
	.buscas-titulo { font-weight: 700; font-size: 0.9rem; color: var(--tinta-suave); }
	.btn-nova {
		border: 1px solid var(--linha); background: var(--porcelana);
		color: var(--tinta); font-size: 0.82rem; font-weight: 600;
		padding: 6px 14px; border-radius: 999px; cursor: pointer;
	}
	.btn-nova:hover { border-color: var(--ouro); color: var(--ouro); }

	/* form nova busca */
	.form-nova-busca {
		background: var(--nevoa); border: 1px solid var(--linha);
		border-radius: var(--raio); padding: var(--r4);
		display: flex; flex-direction: column; gap: var(--r4);
		margin-bottom: var(--r4);
	}
	.form-linha { display: flex; flex-wrap: wrap; gap: var(--r4); }
	.flex1 { flex: 1 1 200px; }
	.entrada-kw { display: flex; gap: var(--r2); }
	.btn-add-kw {
		border: 1px solid var(--linha); background: var(--ouro-fundo);
		color: #7a5a1e; font-weight: 700; font-size: 1rem;
		padding: 0 14px; border-radius: 10px; cursor: pointer; flex-shrink: 0;
	}
	.kws-montadas { display: flex; flex-wrap: wrap; gap: var(--r2); }
	.pilula-kw {
		display: inline-flex; align-items: center; gap: 5px;
		background: var(--ouro-fundo); border: 1px solid color-mix(in srgb, var(--ouro) 40%, var(--linha));
		border-radius: 999px; padding: 3px 6px 3px 10px;
		font-size: 0.85rem; font-weight: 600; color: #7a5a1e;
	}
	.dica-kw { font-size: 0.82rem; color: var(--tinta-suave); margin: 0; }
	.form-acoes { display: flex; justify-content: flex-end; }
	.salvar {
		border: 1px solid var(--linha); background: var(--ouro-fundo);
		color: #7a5a1e; font-weight: 600; font-size: 0.85rem;
		padding: 9px 18px; border-radius: 10px; cursor: pointer;
	}
	.salvar:disabled { opacity: 0.5; cursor: not-allowed; }

	/* lista de buscas salvas */
	.buscas-lista { display: flex; flex-direction: column; gap: var(--r3); }
	.cartao-busca {
		background: var(--nevoa); border: 1px solid var(--linha);
		border-radius: var(--raio); padding: var(--r3) var(--r4);
		display: flex; flex-direction: column; gap: var(--r2);
	}
	.cartao-topo { display: flex; align-items: flex-start; justify-content: space-between; gap: var(--r3); }
	.cartao-kws { display: flex; flex-wrap: wrap; gap: var(--r2); flex: 1; }
	.kw-btn {
		border: 1px solid var(--linha); background: var(--porcelana);
		color: var(--tinta); font-weight: 600; font-size: 0.88rem;
		padding: 5px 12px; border-radius: 999px; cursor: pointer;
	}
	.kw-btn:hover, .kw-btn.kw-ativa { background: var(--ouro-fundo); border-color: var(--ouro); color: #7a5a1e; }
	.cartao-acoes { display: flex; align-items: center; gap: var(--r2); flex-shrink: 0; }
	.btn-prox {
		border: 1px solid var(--linha); background: transparent;
		color: var(--tinta-suave); font-size: 0.9rem; padding: 3px 8px;
		border-radius: 8px; cursor: pointer;
	}
	.btn-prox:hover { color: var(--tinta); border-color: var(--tinta-suave); }
	.cartao-meta { display: flex; flex-wrap: wrap; gap: var(--r2); }
	.badge {
		font-size: 0.72rem; padding: 2px 8px; border-radius: 999px;
		background: var(--porcelana); border: 1px solid var(--linha);
		color: var(--tinta-suave);
	}
	.badge.cron { color: color-mix(in srgb, var(--ouro) 70%, var(--tinta-suave)); }
	.x {
		border: none; background: transparent; color: var(--tinta-suave);
		font-size: 0.72rem; cursor: pointer; padding: 2px 4px;
	}
	.x:hover { color: var(--alerta); }
	.buscas-vazia { font-size: 0.85rem; color: var(--tinta-suave); font-style: italic; margin: var(--r3) 0; }
</style>
