<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { buscasSalvas } from '$lib/buscas.js';
	import { favoritos } from '$lib/favoritos.js';
	import { carregarCuradoria, carregarOportunidades } from '$lib/descobrir.js';
	import { montarResultados } from '$lib/descobrir-logic.js';
	import { prepararPublicacao } from '$lib/publicar-store.js';
	import ProductCard from '$lib/components/ProductCard.svelte';
	import FilterBar from '$lib/components/FilterBar.svelte';
	import { Loading, EmptyState, Button } from '$lib/components/ui/index.js';

	// ── Filtros ───────────────────────────────────────────────────────────────
	let busca = $state('');
	let categoria = $state('');
	let comissaoMin = $state(0.07);
	let vendasMin = $state(0);
	let categorias = $state([]);
	let fontes = $state({ curadoria: true, quedas: true, novos: true, favoritos: false });

	let categoriasEfetivas = $derived(categoria.trim() ? [categoria.trim(), ...categorias] : categorias);

	// ── Estado ────────────────────────────────────────────────────────────────
	let carregando = $state(false);
	let erro = $state(null);
	let resultados = $state([]);
	let dadosCuradoria = $state([]);
	let dadosQuedas = $state([]);
	let dadosNovos = $state([]);

	let buscasComLojas = $derived(($buscasSalvas ?? []).filter((b) => b.shop_ids?.length > 0));
	let nomesLojas = $derived(Object.fromEntries(buscasComLojas.map((b) => [b.id, b.nome || b.id])));
	let buscasSalvasKw = $derived(($buscasSalvas ?? []).filter((b) => !b.shop_ids?.length));

	// Contagens por fonte nos resultados filtrados (para badges)
	let contagemCuradoria = $derived(resultados.filter((r) => r._fonte === 'curadoria').length);
	let contagemQuedas = $derived(resultados.filter((r) => r._fonte === 'queda').length);
	let contagemNovos = $derived(resultados.filter((r) => r._fonte === 'novo').length);

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
			promises.push(
				carregarCuradoria({ busca, comissaoMin, categorias: categoriasEfetivas, buscasComLojas }).then((r) => {
					dadosCuradoria = r;
				})
			);
		} else {
			dadosCuradoria = [];
		}

		if ((fontes.quedas || fontes.novos) && buscasComLojas.length > 0) {
			promises.push(
				carregarOportunidades(buscasComLojas, nomesLojas).then((r) => {
					dadosQuedas = r.quedas;
					dadosNovos = r.novos;
				})
			);
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
			resultados = montarResultados({
				fontes,
				dadosCuradoria,
				dadosQuedas,
				dadosNovos,
				favoritos: $favoritos,
				busca,
				categorias: categoriasEfetivas,
				comissaoMin,
				vendasMin
			});
		}
	}

	// Debounce
	let timer;
	$effect(() => {
		busca;
		categoria;
		categorias;
		comissaoMin;
		vendasMin;
		fontes.curadoria;
		fontes.quedas;
		fontes.novos;
		fontes.favoritos;
		clearTimeout(timer);
		timer = setTimeout(carregar, 400);
		return () => clearTimeout(timer);
	});

	// ── Ações ─────────────────────────────────────────────────────────────────
	function publicar(c) {
		goto(prepararPublicacao(c));
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

<section class="max-w-[900px] space-y-8">
	<div>
		<h1 class="text-[clamp(1.8rem,5vw,2.5rem)] mb-2">O que publicar hoje?</h1>
		<p class="text-tinta-suave text-[0.95rem]">
			Encontre produtos para divulgar — por busca, oportunidades ou favoritos.
		</p>
	</div>

	<FilterBar bind:busca bind:categoria bind:comissaoMin bind:vendasMin mostrarBusca={true} />

	<!-- Fontes -->
	<div class="flex flex-wrap gap-1.5">
		<button
			class="py-[7px] px-3.5 border border-border rounded-full bg-porcelana text-tinta-suave text-[0.82rem] font-semibold cursor-pointer flex items-center gap-1 transition-[border-color,background] duration-150 hover:border-ouro hover:text-foreground {fontes.curadoria
				? 'bg-ouro-fundo border-ouro-claro text-ouro-escuro'
				: ''}"
			onclick={() => {
				fontes.curadoria = !fontes.curadoria;
			}}
			type="button"
			title="Busca por palavra-chave na API de afiliados Shopee"
		>
			🔍 Busca {#if fontes.curadoria && contagemCuradoria > 0}<span
					class="text-[0.65rem] bg-ouro text-white w-4 h-4 rounded-full flex items-center justify-center font-bold"
					>{contagemCuradoria}</span
				>{/if}
		</button>
		<button
			class="py-[7px] px-3.5 border border-border rounded-full bg-porcelana text-tinta-suave text-[0.82rem] font-semibold cursor-pointer flex items-center gap-1 transition-[border-color,background] duration-150 hover:border-ouro hover:text-foreground {fontes.quedas
				? 'bg-ouro-fundo border-ouro-claro text-ouro-escuro'
				: ''}"
			onclick={() => {
				fontes.quedas = !fontes.quedas;
			}}
			type="button"
			title="Produtos que caíram de preço nas lojas monitoradas"
		>
			📉 Quedas {#if contagemQuedas > 0}<span
					class="text-[0.65rem] bg-sucesso text-white w-4 h-4 rounded-full flex items-center justify-center font-bold"
					>{contagemQuedas}</span
				>{/if}
		</button>
		<button
			class="py-[7px] px-3.5 border border-border rounded-full bg-porcelana text-tinta-suave text-[0.82rem] font-semibold cursor-pointer flex items-center gap-1 transition-[border-color,background] duration-150 hover:border-ouro hover:text-foreground {fontes.novos
				? 'bg-ouro-fundo border-ouro-claro text-ouro-escuro'
				: ''}"
			onclick={() => {
				fontes.novos = !fontes.novos;
			}}
			type="button"
			title="Produtos novos detectados nas lojas monitoradas"
		>
			🆕 Novos {#if contagemNovos > 0}<span
					class="text-[0.65rem] bg-rosa text-white w-4 h-4 rounded-full flex items-center justify-center font-bold"
					>{contagemNovos}</span
				>{/if}
		</button>
		<button
			class="py-[7px] px-3.5 border border-border rounded-full bg-porcelana text-tinta-suave text-[0.82rem] font-semibold cursor-pointer flex items-center gap-1 transition-[border-color,background] duration-150 hover:border-ouro hover:text-foreground {fontes.favoritos
				? 'bg-ouro-fundo border-ouro-claro text-ouro-escuro'
				: ''}"
			onclick={() => {
				fontes.favoritos = !fontes.favoritos;
			}}
			type="button"
			title="Produtos que você salvou como favorito"
		>
			⭐ Favoritos {#if $favoritos.length > 0}<span
					class="text-[0.65rem] bg-ouro text-white w-4 h-4 rounded-full flex items-center justify-center font-bold"
					>{$favoritos.length}</span
				>{/if}
		</button>
	</div>
	{#if !fontes.curadoria && !fontes.quedas && !fontes.novos && !fontes.favoritos}
		<p class="text-[0.82rem] text-tinta-suave italic">Ative ao menos uma fonte para ver resultados.</p>
	{:else if fontes.curadoria && !busca.trim() && categoriasEfetivas.length === 0 && !fontes.quedas && !fontes.novos && !fontes.favoritos}
		<p class="text-[0.82rem] text-tinta-suave italic">Digite um termo acima para buscar produtos.</p>
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
			class="bg-card border border-[color-mix(in_srgb,var(--erro-texto)_30%,var(--linha))] rounded-md p-5 text-center"
		>
			<p class="my-2"><strong>😕 {erro.message ?? erro}</strong></p>
			<Button size="sm" onclick={carregar}>🔄 Tentar novamente</Button>
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
		<p class="text-[0.82rem] text-tinta-suave">
			{resultados.length}
			{resultados.length === 1 ? 'produto' : 'produtos'}
		</p>
		<div class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
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
</section>
