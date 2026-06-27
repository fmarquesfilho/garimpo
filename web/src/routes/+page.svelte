<script>
	import { onMount } from 'svelte';
	import { buscarCandidatos, buscarNovidades, registrarSelecao } from '$lib/api.js';
	import { goto } from '$app/navigation';
	import { buscasSalvas } from '$lib/buscas.js';
	import { favoritos } from '$lib/favoritos.js';
	import ProductCard from '$lib/components/ProductCard.svelte';
	import FilterBar from '$lib/components/FilterBar.svelte';
	import { Loading, EmptyState } from '$lib/components/ui/index.js';

	// ── Filtros ───────────────────────────────────────────────────────────────
	let busca = $state('');
	let categorias = $state([]); // categorias ativas (filtro)
	let fontes = $state({ curadoria: true, quedas: true, novos: true, favoritos: false });

	// ── Estado dos dados ──────────────────────────────────────────────────────
	let carregando = $state(false);
	let erro = $state(null);
	let resultados = $state([]);

	// Dados brutos de cada fonte (cache)
	let dadosCuradoria = $state([]);
	let dadosQuedas = $state([]);
	let dadosNovos = $state([]);

	let buscasComLojas = $derived(($buscasSalvas ?? []).filter(b => b.shop_ids?.length > 0));
	let nomesLojas = $derived(Object.fromEntries(buscasComLojas.map(b => [b.id, b.nome || b.id])));

	// Buscas salvas por keyword (sem lojas — atalhos na UI)
	let buscasSalvasKw = $derived(($buscasSalvas ?? []).filter(b => !b.shop_ids?.length));

	// ── Carregamento ──────────────────────────────────────────────────────────
	onMount(async () => {
		await buscasSalvas.sincronizarDoServidor();
		favoritos.sincronizar();
		carregar();
	});

	async function carregar() {
		carregando = true;
		erro = null;

		const promises = [];

		// Curadoria: busca por keyword ou categoria (se tem termo/categoria E fonte ativa)
		if (fontes.curadoria && (busca.trim() || categorias.length > 0)) {
			promises.push(carregarCuradoria());
		} else {
			dadosCuradoria = [];
		}

		// Oportunidades: quedas e novos das lojas monitoradas
		if ((fontes.quedas || fontes.novos) && buscasComLojas.length > 0) {
			promises.push(carregarOportunidades());
		} else {
			dadosQuedas = [];
			dadosNovos = [];
		}

		let timeoutId;
		const timeout = new Promise((_, reject) => {
			timeoutId = setTimeout(() => reject(new Error('A busca demorou demais. Tente novamente.')), 25000);
		});

		try {
			await Promise.race([Promise.all(promises), timeout]);
		} catch (e) {
			erro = e;
		} finally {
			clearTimeout(timeoutId);
			carregando = false;
			montarResultados();
		}
	}

	async function carregarCuradoria() {
		try {
			const params = { estrategia: 'nicho', top: 20, keyword: busca.trim() };
			if (categorias.length > 0) params.categoria = categorias[0];
			const r = await buscarCandidatos(params);
			dadosCuradoria = (r.candidatos ?? []).map(c => ({ ...c, _fonte: 'curadoria' }));
		} catch { dadosCuradoria = []; }
	}

	// Cache de oportunidades (não re-busca se < 2 min)
	let cacheOportunidades = { em: 0, quedas: [], novos: [] };

	async function carregarOportunidades() {
		// Usa cache se dados foram buscados há menos de 2 minutos
		if (Date.now() - cacheOportunidades.em < 120000 && (cacheOportunidades.quedas.length > 0 || cacheOportunidades.novos.length > 0)) {
			dadosQuedas = cacheOportunidades.quedas;
			dadosNovos = cacheOportunidades.novos;
			return;
		}

		try {
			const promises = buscasComLojas.map(b =>
				buscarNovidades({ buscaId: b.id, dias: 7 }).then(r => ({ ...r, loja: b.id })).catch(() => null)
			);
			const resultados = await Promise.all(promises);
			const quedas = [], novos = [];

			for (const r of resultados) {
				if (!r) continue;
				for (const v of (r.variacoes ?? [])) {
					if (v.variacao_pct < 0) {
						quedas.push({
							id: v.produto_id, produto_id: v.produto_id, nome: v.nome,
							preco: v.preco_atual, preco_anterior: v.preco_anterior,
							variacao_pct: v.variacao_pct, detectado_em: v.detectado_em,
							loja: v.loja || (nomesLojas[r.loja] ?? r.loja), _loja_id: r.loja,
							imagem: v.imagem, link: v.link, comissao: v.comissao ?? 0,
							vendas: v.vendas ?? 0, _fonte: 'queda'
						});
					}
				}
				for (const p of (r.produtos_novos ?? [])) {
					novos.push({
						id: p.produto_id, produto_id: p.produto_id, nome: p.nome,
						preco: p.preco, comissao: p.comissao ?? 0, vendas: p.vendas ?? 0,
						detectado_em: p.detectado_em, loja: p.loja || (nomesLojas[r.loja] ?? r.loja),
						_loja_id: r.loja, imagem: p.imagem, link: p.link, _fonte: 'novo'
					});
				}
			}
			quedas.sort((a, b) => a.variacao_pct - b.variacao_pct);
			novos.sort((a, b) => (b.detectado_em ?? '').localeCompare(a.detectado_em ?? ''));
			dadosQuedas = quedas;
			dadosNovos = novos;
			cacheOportunidades = { em: Date.now(), quedas, novos };
		} catch {
			dadosQuedas = [];
			dadosNovos = [];
		}
	}

	function montarResultados() {
		let todos = [];
		if (fontes.curadoria) todos.push(...dadosCuradoria);
		if (fontes.quedas) todos.push(...dadosQuedas);
		if (fontes.novos) todos.push(...dadosNovos);
		if (fontes.favoritos) {
			const favs = ($favoritos ?? []).map(f => ({ ...f, id: f.produto_id, _fonte: 'favorito' }));
			todos.push(...favs);
		}

		// Filtra por busca (keyword/loja)
		const termo = busca.trim().toLowerCase();
		if (termo) {
			todos = todos.filter(r =>
				(r.nome ?? '').toLowerCase().includes(termo) ||
				(r.loja ?? '').toLowerCase().includes(termo)
			);
		}

		// Filtra por categorias (se alguma está selecionada)
		if (categorias.length > 0) {
			const cats = categorias.map(c => c.toLowerCase());
			todos = todos.filter(r =>
				!r.categoria || cats.some(c => (r.categoria ?? '').toLowerCase().includes(c))
			);
		}

		resultados = todos;
	}

	// Debounce: recarrega quando busca ou fontes mudam
	let timer;
	$effect(() => {
		busca; categorias; fontes.curadoria; fontes.quedas; fontes.novos; fontes.favoritos;
		clearTimeout(timer);
		timer = setTimeout(carregar, 400);
		return () => clearTimeout(timer);
	});

	// ── Ações ─────────────────────────────────────────────────────────────────
	function publicar(c) {
		registrarSelecao(c);
		goto(`/publicar?dados=${encodeURIComponent(JSON.stringify(c))}`);
	}

	function aplicarBusca(kw) {
		busca = kw;
		fontes.curadoria = true;
	}

	/** Aplica uma busca salva completa — seta keyword + ativa fontes correspondentes. */
	function aplicarBuscaSalva(b) {
		const kw = (b.keywords ?? [])[0] ?? '';
		busca = kw;
		categorias = b.categorias ?? [];
		if (b.fontes?.length) {
			fontes.curadoria = b.fontes.includes('curadoria');
			fontes.quedas = b.fontes.includes('quedas');
			fontes.novos = b.fontes.includes('novos');
			fontes.favoritos = b.fontes.includes('favoritos');
		} else {
			fontes.curadoria = kw.length > 0;
		}
	}
</script>

<svelte:head>
	<title>Descobrir — Garimpei</title>
</svelte:head>

<section class="page">
	<h1>O que publicar hoje?</h1>
	<p class="sub">Encontre produtos para divulgar — por busca, oportunidades ou favoritos.</p>

	<!-- Busca universal -->
	<FilterBar bind:busca={busca} mostrarBusca={true} />

	<!-- Filtros de fonte -->
	<div class="fontes">
		<button class="fonte-btn" class:ativa={fontes.curadoria} onclick={() => { fontes.curadoria = !fontes.curadoria; }} type="button" title="Busca por palavra-chave na API de afiliados Shopee">
			🔍 Busca {#if fontes.curadoria && dadosCuradoria.length > 0}<span class="fonte-badge">{dadosCuradoria.length}</span>{/if}
		</button>
		<button class="fonte-btn" class:ativa={fontes.quedas} onclick={() => { fontes.quedas = !fontes.quedas; }} type="button" title="Produtos que caíram de preço nas lojas monitoradas">
			📉 Quedas {#if dadosQuedas.length > 0}<span class="fonte-badge queda">{dadosQuedas.length}</span>{/if}
		</button>
		<button class="fonte-btn" class:ativa={fontes.novos} onclick={() => { fontes.novos = !fontes.novos; }} type="button" title="Produtos novos detectados nas lojas monitoradas">
			🆕 Novos {#if dadosNovos.length > 0}<span class="fonte-badge novo">{dadosNovos.length}</span>{/if}
		</button>
		<button class="fonte-btn" class:ativa={fontes.favoritos} onclick={() => { fontes.favoritos = !fontes.favoritos; }} type="button" title="Produtos que você salvou como favorito">
			⭐ Favoritos {#if $favoritos.length > 0}<span class="fonte-badge">{$favoritos.length}</span>{/if}
		</button>
	</div>
	{#if !fontes.curadoria && !fontes.quedas && !fontes.novos && !fontes.favoritos}
		<p class="hint-fontes">Ative ao menos uma fonte para ver resultados.</p>
	{:else if fontes.curadoria && !busca.trim() && !fontes.quedas && !fontes.novos && !fontes.favoritos}
		<p class="hint-fontes">Digite um termo acima para buscar produtos.</p>
	{/if}

	<!-- Buscas salvas (atalhos) -->
	{#if buscasSalvasKw.length > 0}
		<div class="buscas-atalhos">
			{#each buscasSalvasKw as b (b.id)}
				<div class="atalho-busca">
					{#if b.cron}<span class="atalho-icone" title="Busca agendada">⏱</span>{/if}
					{#if b.fontes?.includes('quedas')}<span class="atalho-icone" title="Monitora quedas">📉</span>{/if}
					{#if b.fontes?.includes('novos')}<span class="atalho-icone" title="Monitora novos">🆕</span>{/if}
					{#each b.keywords ?? [] as kw}
						<button class="kw-pill" class:ativa={busca === kw} onclick={() => aplicarBuscaSalva(b)} type="button">{kw}</button>
					{/each}
					{#if (b.keywords ?? []).length === 0 && b.categorias?.length}
						{#each b.categorias as cat}
							<button class="kw-pill cat-pill" onclick={() => aplicarBuscaSalva(b)} type="button">{cat}</button>
						{/each}
					{/if}
				</div>
			{/each}
		</div>
	{/if}

	<!-- Resultados -->
	{#if carregando}
		<Loading mensagem="Buscando produtos…" />
	{:else if erro}
		<div class="msg-erro">
			<p><strong>😕 {erro.message ?? erro}</strong></p>
			<button class="btn-retry" onclick={carregar}>🔄 Tentar novamente</button>
		</div>
	{:else if resultados.length === 0}
		<EmptyState
			icone="🔍"
			mensagem={busca.trim()
				? `Nenhum resultado para "${busca}".`
				: buscasComLojas.length === 0 && (fontes.quedas || fontes.novos)
					? 'Você ainda não monitora nenhuma loja.'
					: 'Nenhum resultado com os filtros atuais.'}
			dica={busca.trim()
				? 'Tente outro termo ou ative mais fontes.'
				: buscasComLojas.length === 0 && (fontes.quedas || fontes.novos)
					? 'Adicione lojas em <a href="/lojas">Lojas</a> para ver quedas e produtos novos.'
					: fontes.curadoria
						? 'Digite um termo acima para buscar por palavra-chave.'
						: 'Ative "Busca" e digite um termo, ou monitore lojas para ver oportunidades.'}
		/>
	{:else}
		<p class="contagem">{resultados.length} {resultados.length === 1 ? 'produto' : 'produtos'}</p>
		<div class="grade">
			{#each resultados as produto, i (produto.id || produto.produto_id || i)}
				<ProductCard
					{produto}
					posicao={fontes.curadoria && produto._fonte === 'curadoria' ? i + 1 : null}
					variacao={produto._fonte === 'queda' ? { tipo: 'queda', pct: produto.variacao_pct, preco_anterior: produto.preco_anterior, preco_atual: produto.preco, detectado_em: produto.detectado_em } : produto._fonte === 'novo' ? { tipo: 'novo', detectado_em: produto.detectado_em } : null}
					onpublicar={publicar}
					onfavoritar={(p) => favoritos.toggle(p)}
				/>
			{/each}
		</div>
	{/if}
</section>

<style>
	.page { max-width: 900px; }
	h1 { font-size: clamp(1.8rem, 5vw, 2.5rem); margin: 0 0 var(--r2); }
	.sub { color: var(--tinta-suave); font-size: 0.95rem; margin: 0 0 var(--r5); }

	/* Fontes (toggle buttons) */
	.fontes { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: var(--r6); }
	.fonte-btn {
		padding: 7px 14px; border: 1px solid var(--linha); border-radius: var(--raio-full);
		background: var(--porcelana); color: var(--tinta-suave);
		font-size: 0.82rem; font-weight: 600; cursor: pointer;
		display: flex; align-items: center; gap: 4px;
		transition: border-color 0.15s, background 0.15s;
	}
	.fonte-btn:hover { border-color: var(--ouro); color: var(--tinta); }
	.fonte-btn.ativa { background: var(--ouro-fundo); border-color: var(--ouro-claro); color: var(--ouro-escuro); }
	.fonte-badge { font-size: 0.65rem; background: var(--ouro); color: white; width: 16px; height: 16px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 700; }
	.fonte-badge.queda { background: var(--sucesso-texto); }
	.fonte-badge.novo { background: var(--rosa); }
	.hint-fontes { font-size: 0.82rem; color: var(--tinta-suave); font-style: italic; margin: 0 0 var(--r4); }

	/* Buscas salvas (atalhos) */
	.buscas-atalhos { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: var(--r5); }
	.atalho-busca { display: flex; align-items: center; gap: 4px; }
	.atalho-icone { font-size: 0.75rem; color: var(--ouro); }
	.kw-pill {
		padding: 5px 12px; border: 1px solid var(--linha); border-radius: var(--raio-full);
		background: var(--porcelana); color: var(--tinta); font-size: 0.82rem;
		font-weight: 600; cursor: pointer;
	}
	.kw-pill:hover { border-color: var(--ouro); color: var(--ouro-escuro); }
	.kw-pill.ativa { background: var(--ouro-fundo); border-color: var(--ouro-claro); color: var(--ouro-escuro); }
	.cat-pill { color: var(--rosa); border-color: color-mix(in srgb, var(--rosa) 30%, var(--linha)); }

	/* Resultados */
	.contagem { font-size: 0.82rem; color: var(--tinta-suave); margin-bottom: var(--r4); }
	.grade {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: var(--r5);
	}
	.msg-erro {
		background: var(--nevoa); border: 1px solid color-mix(in srgb, var(--erro-texto) 30%, var(--linha));
		border-radius: var(--raio); padding: var(--r5); text-align: center;
	}
	.msg-erro p { margin: var(--r2) 0; }
	.btn-retry { margin-top: var(--r3); padding: 8px 16px; background: var(--ouro); color: white; border: none; border-radius: 8px; font-weight: 600; font-size: 0.85rem; cursor: pointer; }
	.btn-retry:hover { opacity: 0.9; }
</style>
