<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { buscasSalvas } from '$lib/buscas.js';
	import { favoritos } from '$lib/favoritos.js';
	import { carregarCuradoria, carregarOportunidades } from '$lib/descobrir.js';
	import { montarResultados } from '$lib/descobrir-logic.js';
	import ProductCard from '$lib/components/ProductCard.svelte';
	import FilterBar from '$lib/components/FilterBar.svelte';
	import { Loading, EmptyState } from '$lib/components/ui/index.js';

	// ── Filtros ───────────────────────────────────────────────────────────────
	let busca = $state('');
	let categoria = $state('');
	let comissaoMin = $state(0.07);
	let vendasMin = $state(0);
	let notaMin = $state(0);
	let categorias = $state([]);
	let fontes = $state({ curadoria: true, quedas: true, novos: true, favoritos: false });

	let categoriasEfetivas = $derived(
		categoria.trim() ? [categoria.trim(), ...categorias] : categorias
	);

	// ── Estado ────────────────────────────────────────────────────────────────
	let carregando = $state(false);
	let erro = $state(null);
	let resultados = $state([]);
	let dadosCuradoria = $state([]);
	let dadosQuedas = $state([]);
	let dadosNovos = $state([]);

	let buscasComLojas = $derived(($buscasSalvas ?? []).filter(b => b.shop_ids?.length > 0));
	let nomesLojas = $derived(Object.fromEntries(buscasComLojas.map(b => [b.id, b.nome || b.id])));
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

		if (fontes.curadoria && (busca.trim() || categoriasEfetivas.length > 0)) {
			promises.push(carregarCuradoria({ busca, comissaoMin, vendasMin, notaMin, categorias: categoriasEfetivas, buscasComLojas }).then(r => { dadosCuradoria = r; }));
		} else {
			dadosCuradoria = [];
		}

		if ((fontes.quedas || fontes.novos) && buscasComLojas.length > 0) {
			promises.push(carregarOportunidades(buscasComLojas, nomesLojas).then(r => { dadosQuedas = r.quedas; dadosNovos = r.novos; }));
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
			resultados = montarResultados({ fontes, dadosCuradoria, dadosQuedas, dadosNovos, favoritos: $favoritos, busca, categorias: categoriasEfetivas, comissaoMin, vendasMin, notaMin });
		}
	}

	// Debounce
	let timer;
	$effect(() => {
		busca; categoria; categorias; comissaoMin; vendasMin; notaMin;
		fontes.curadoria; fontes.quedas; fontes.novos; fontes.favoritos;
		clearTimeout(timer);
		timer = setTimeout(carregar, 400);
		return () => clearTimeout(timer);
	});

	// ── Ações ─────────────────────────────────────────────────────────────────
	function publicar(c) {
		goto(`/publicar?dados=${encodeURIComponent(JSON.stringify(c))}`);
	}

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

	<FilterBar bind:busca={busca} bind:categoria={categoria} bind:comissaoMin={comissaoMin} bind:vendasMin={vendasMin} bind:notaMin={notaMin} mostrarBusca={true} />

	<!-- Fontes -->
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
	{:else if fontes.curadoria && !busca.trim() && categoriasEfetivas.length === 0 && !fontes.quedas && !fontes.novos && !fontes.favoritos}
		<p class="hint-fontes">Digite um termo acima para buscar produtos.</p>
	{/if}

	<!-- Atalhos -->
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
			mensagem={busca.trim() ? `Nenhum resultado para "${busca}".` : buscasComLojas.length === 0 && (fontes.quedas || fontes.novos) ? 'Você ainda não monitora nenhuma loja.' : 'Nenhum resultado com os filtros atuais.'}
			dica={busca.trim() ? 'Tente outro termo ou ative mais fontes.' : buscasComLojas.length === 0 && (fontes.quedas || fontes.novos) ? 'Adicione lojas em <a href="/lojas">Lojas</a> para ver quedas e produtos novos.' : fontes.curadoria ? 'Digite um termo acima para buscar por palavra-chave.' : 'Ative "Busca" e digite um termo, ou monitore lojas para ver oportunidades.'}
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
	.fontes { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: var(--r6); }
	.fonte-btn { padding: 7px 14px; border: 1px solid var(--linha); border-radius: var(--raio-full); background: var(--porcelana); color: var(--tinta-suave); font-size: 0.82rem; font-weight: 600; cursor: pointer; display: flex; align-items: center; gap: 4px; transition: border-color 0.15s, background 0.15s; }
	.fonte-btn:hover { border-color: var(--ouro); color: var(--tinta); }
	.fonte-btn.ativa { background: var(--ouro-fundo); border-color: var(--ouro-claro); color: var(--ouro-escuro); }
	.fonte-badge { font-size: 0.65rem; background: var(--ouro); color: white; width: 16px; height: 16px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 700; }
	.fonte-badge.queda { background: var(--sucesso-texto); }
	.fonte-badge.novo { background: var(--rosa); }
	.hint-fontes { font-size: 0.82rem; color: var(--tinta-suave); font-style: italic; margin: 0 0 var(--r4); }
	.buscas-atalhos { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: var(--r5); }
	.atalho-busca { display: flex; align-items: center; gap: 4px; }
	.atalho-icone { font-size: 0.75rem; color: var(--ouro); }
	.kw-pill { padding: 5px 12px; border: 1px solid var(--linha); border-radius: var(--raio-full); background: var(--porcelana); color: var(--tinta); font-size: 0.82rem; font-weight: 600; cursor: pointer; }
	.kw-pill:hover { border-color: var(--ouro); color: var(--ouro-escuro); }
	.kw-pill.ativa { background: var(--ouro-fundo); border-color: var(--ouro-claro); color: var(--ouro-escuro); }
	.cat-pill { color: var(--rosa); border-color: color-mix(in srgb, var(--rosa) 30%, var(--linha)); }
	.contagem { font-size: 0.82rem; color: var(--tinta-suave); margin-bottom: var(--r4); }
	.grade { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: var(--r5); }
	.msg-erro { background: var(--nevoa); border: 1px solid color-mix(in srgb, var(--erro-texto) 30%, var(--linha)); border-radius: var(--raio); padding: var(--r5); text-align: center; }
	.msg-erro p { margin: var(--r2) 0; }
	.btn-retry { margin-top: var(--r3); padding: 8px 16px; background: var(--ouro); color: white; border: none; border-radius: 8px; font-weight: 600; font-size: 0.85rem; cursor: pointer; }
	.btn-retry:hover { opacity: 0.9; }
</style>
