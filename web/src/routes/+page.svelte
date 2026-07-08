<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { buscasSalvas } from '$lib/buscas.js';
	import { favoritos } from '$lib/favoritos.js';
	import { carregarFontes } from '$lib/descobrir.js';
	import { montarResultados, buildFonteOpcoes } from '$lib/descobrir-logic.js';
	import { prepararPublicacao } from '$lib/publicar-store.js';
	import ProductCard from '$lib/components/ProductCard.svelte';
	import FilterBar from '$lib/components/FilterBar.svelte';
	import FormAdicionarLoja from '$lib/components/FormAdicionarLoja.svelte';
	import GerenciarBuscas from '$lib/components/GerenciarBuscas.svelte';
	import PainelAlertas from '$lib/components/PainelAlertas.svelte';
	import { Loading, EmptyState, Button, ToggleGroup, Collapsible } from '$lib/components/ui/index.js';
	import { usuario } from '$lib/firebase.js';

	// ── Filtros ───────────────────────────────────────────────────────────────
	let busca = $state('');
	let categoria = $state('');
	let comissaoMin = $state(0.07);
	let vendasMin = $state(0);
	let categorias = $state([]);
	let fontes = $state({ curadoria: true, quedas: true, novos: true, favoritos: false, lojas: false });
	let categoriasEfetivas = $derived(categoria.trim() ? [categoria.trim(), ...categorias] : categorias);

	// ── Estado ────────────────────────────────────────────────────────────────
	let carregando = $state(false);
	let erro = $state(null);
	let resultados = $state([]);
	let dadosCuradoria = $state([]);
	let dadosQuedas = $state([]);
	let dadosNovos = $state([]);
	let dadosLojas = $state([]);
	let lojaFiltro = $state(null);
	let mostrarConfig = $state(false);

	let buscasComLojas = $derived(($buscasSalvas ?? []).filter((b) => b.shop_ids?.length > 0));
	let nomesLojas = $derived(Object.fromEntries(buscasComLojas.map((b) => [b.id, b.nome || b.id])));
	let buscasSalvasKw = $derived(($buscasSalvas ?? []).filter((b) => !b.shop_ids?.length));
	let dadosLojasFiltrados = $derived(lojaFiltro ? dadosLojas.filter((p) => p._loja_id === lojaFiltro) : dadosLojas);

	// Contagens e opções de UI
	let contagemCuradoria = $derived(resultados.filter((r) => r._fonte === 'curadoria').length);
	let contagemQuedas = $derived(resultados.filter((r) => r._fonte === 'queda').length);
	let contagemNovos = $derived(resultados.filter((r) => r._fonte === 'novo').length);
	let contagemLojas = $derived(resultados.filter((r) => r._fonte === 'loja').length);
	let fontesAtivas = $derived(
		Object.entries(fontes)
			.filter(([, v]) => v)
			.map(([k]) => k)
	);
	let fonteOpcoes = $derived(
		buildFonteOpcoes({
			contagemCuradoria,
			contagemQuedas,
			contagemNovos,
			contagemLojas,
			totalFavoritos: $favoritos.length
		})
	);
	let lojaOpcoes = $derived([
		{ value: '', label: 'Todas' },
		...buscasComLojas.map((b) => ({ value: b.id, label: b.nome || b.id }))
	]);
	let nenhumaFonteAtiva = $derived(
		!fontes.curadoria && !fontes.quedas && !fontes.novos && !fontes.favoritos && !fontes.lojas
	);

	function handleFontesChange(v) {
		fontes = {
			curadoria: v.includes('curadoria'),
			quedas: v.includes('quedas'),
			novos: v.includes('novos'),
			favoritos: v.includes('favoritos'),
			lojas: v.includes('lojas')
		};
	}

	// ── Carregamento ──────────────────────────────────────────────────────────
	onMount(async () => {
		await buscasSalvas.sincronizarDoServidor();
		favoritos.sincronizar();
		carregar();
	});

	async function carregar() {
		carregando = true;
		erro = null;
		try {
			const r = await carregarFontes({
				fontes,
				busca,
				comissaoMin,
				categorias: categoriasEfetivas,
				buscasComLojas,
				nomesLojas
			});
			dadosCuradoria = r.curadoria;
			dadosQuedas = r.quedas;
			dadosNovos = r.novos;
			dadosLojas = r.lojas;
		} catch (e) {
			erro = e;
		} finally {
			carregando = false;
			resultados = montarResultados({
				fontes,
				dadosCuradoria,
				dadosQuedas,
				dadosNovos,
				dadosLojas: dadosLojasFiltrados,
				favoritos: $favoritos,
				busca,
				categorias: categoriasEfetivas,
				comissaoMin,
				vendasMin
			});
		}
	}

	// Debounce — re-carrega quando qualquer filtro muda
	let timer;
	$effect(() => {
		// Track reactive dependencies (reads trigger re-run)
		void busca;
		void categoria;
		void categorias;
		void comissaoMin;
		void vendasMin;
		void fontes.curadoria;
		void fontes.quedas;
		void fontes.novos;
		void fontes.favoritos;
		void fontes.lojas;
		void lojaFiltro;
		clearTimeout(timer);
		timer = setTimeout(carregar, 400);
		return () => clearTimeout(timer);
	});

	// ── Ações ─────────────────────────────────────────────────────────────────
	function publicar(c) {
		goto(prepararPublicacao(c));
	}
	function handleLojaAdicionada() {
		buscasSalvas.sincronizarDoServidor();
	}

	function aplicarBuscaSalva(b) {
		const kw = (b.keywords ?? [])[0] ?? '';
		busca = kw;
		categorias = b.categorias ?? [];
		if (b.fontes?.length) {
			fontes = {
				curadoria: b.fontes.includes('curadoria'),
				quedas: b.fontes.includes('quedas'),
				novos: b.fontes.includes('novos'),
				favoritos: b.fontes.includes('favoritos'),
				lojas: fontes.lojas
			};
		} else {
			fontes = { ...fontes, curadoria: kw.length > 0 };
		}
	}
</script>

<svelte:head>
	<title>Garimpar — Garimpei</title>
</svelte:head>

<section class="max-w-[900px] space-y-8">
	<div>
		<h1 class="text-[clamp(1.8rem,5vw,2.5rem)] mb-2">O que publicar hoje?</h1>
		<p class="text-tinta-suave text-[0.95rem]">Busque produtos, monitore lojas e publique com um clique.</p>
	</div>

	<FilterBar bind:busca bind:categoria bind:comissaoMin bind:vendasMin mostrarBusca={true} />

	<!-- Fontes -->
	<ToggleGroup type="multiple" value={fontesAtivas} options={fonteOpcoes} onchange={handleFontesChange} />

	<!-- Seletor de loja (visível quando fonte Lojas ativa) -->
	{#if fontes.lojas && buscasComLojas.length > 0}
		<ToggleGroup
			type="single"
			value={lojaFiltro ?? ''}
			options={lojaOpcoes}
			nullable={false}
			onchange={(v) => (lojaFiltro = v || null)}
			size="sm"
		/>
	{/if}

	<!-- Hints -->
	{#if nenhumaFonteAtiva}
		<p class="hint-fontes text-[0.82rem] text-tinta-suave italic">Ative ao menos uma fonte para ver resultados.</p>
	{:else if fontes.curadoria && !busca.trim() && categoriasEfetivas.length === 0 && !fontes.quedas && !fontes.novos && !fontes.favoritos && !fontes.lojas}
		<p class="text-[0.82rem] text-tinta-suave italic">Digite um termo acima para buscar produtos.</p>
	{:else if fontes.lojas && buscasComLojas.length === 0 && !fontes.curadoria && !fontes.quedas && !fontes.novos && !fontes.favoritos}
		<p class="text-[0.82rem] text-tinta-suave italic">
			Nenhuma loja monitorada. Adicione uma na seção ⚙️ Configuração abaixo.
		</p>
	{/if}

	<!-- Atalhos -->
	{#if buscasSalvasKw.length > 0}
		<div class="flex flex-wrap gap-2">
			{#each buscasSalvasKw as b (b.id)}
				<div class="flex items-center gap-1">
					{#if b.cron}<span class="text-xs text-ouro" title="Busca agendada">⏱</span>{/if}
					{#if b.fontes?.includes('quedas')}<span class="text-xs text-ouro" title="Monitora quedas">📉</span>{/if}
					{#if b.fontes?.includes('novos')}<span class="text-xs text-ouro" title="Monitora novos">🆕</span>{/if}
					{#each b.keywords ?? [] as kw}
						<button
							class="py-[5px] px-3 border border-border rounded-full bg-porcelana text-foreground text-[0.82rem] font-semibold cursor-pointer hover:border-ouro hover:text-ouro-escuro {busca ===
							kw
								? 'bg-ouro-fundo border-ouro-claro text-ouro-escuro'
								: ''}"
							onclick={() => aplicarBuscaSalva(b)}
							type="button">{kw}</button
						>
					{/each}
					{#if (b.keywords ?? []).length === 0 && b.categorias?.length}
						{#each b.categorias as cat}
							<button
								class="py-[5px] px-3 border rounded-full bg-porcelana text-rosa border-[color-mix(in_srgb,var(--rosa)_30%,var(--linha))] text-[0.82rem] font-semibold cursor-pointer hover:border-ouro hover:text-ouro-escuro"
								onclick={() => aplicarBuscaSalva(b)}
								type="button">{cat}</button
							>
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
		<div
			class="msg-erro bg-card border border-[color-mix(in_srgb,var(--erro-texto)_30%,var(--linha))] rounded-md p-5 text-center"
		>
			<p class="my-2"><strong>😕 {erro.message ?? erro}</strong></p>
			<Button size="sm" onclick={carregar}>🔄 Tentar novamente</Button>
		</div>
	{:else if resultados.length === 0 && !nenhumaFonteAtiva}
		<EmptyState
			icone="🔍"
			mensagem={busca.trim()
				? `Nenhum resultado para "${busca}".`
				: buscasComLojas.length === 0 && (fontes.quedas || fontes.novos || fontes.lojas)
					? 'Você ainda não monitora nenhuma loja.'
					: 'Nenhum resultado com os filtros atuais.'}
			dica={busca.trim()
				? 'Tente outro termo ou ative mais fontes.'
				: buscasComLojas.length === 0 && (fontes.quedas || fontes.novos || fontes.lojas)
					? 'Adicione lojas na seção ⚙️ Configuração abaixo para ver produtos.'
					: fontes.curadoria
						? 'Digite um termo acima para buscar por palavra-chave.'
						: 'Ative "Busca" e digite um termo, ou monitore lojas para ver oportunidades.'}
		/>
	{:else if resultados.length > 0}
		<p class="contagem text-[0.82rem] text-tinta-suave">
			{resultados.length}
			{resultados.length === 1 ? 'produto' : 'produtos'}
		</p>
		<div class="grade grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
			{#each resultados as produto, i (produto.id || produto.produto_id || i)}
				<ProductCard
					{produto}
					posicao={fontes.curadoria && produto._fonte === 'curadoria' ? i + 1 : null}
					variacao={produto._fonte === 'queda'
						? {
								tipo: 'queda',
								pct: produto.variacao_pct,
								preco_anterior: produto.preco_anterior,
								preco_atual: produto.preco,
								detectado_em: produto.detectado_em
							}
						: produto._fonte === 'novo'
							? { tipo: 'novo', detectado_em: produto.detectado_em }
							: null}
					onpublicar={publicar}
					onfavoritar={(p) => favoritos.toggle(p)}
				/>
			{/each}
		</div>
	{/if}

	<!-- Configuração (colapsável) -->
	{#if $usuario}
		<Collapsible bind:open={mostrarConfig} title="⚙️ Configuração">
			<FormAdicionarLoja onadicionada={handleLojaAdicionada} />
			<GerenciarBuscas />
			<PainelAlertas buscaSelecionada={null} />
		</Collapsible>
	{/if}
</section>
