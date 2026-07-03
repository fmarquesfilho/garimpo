<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { buscarCandidatos, buscarNovidades, removerLoja } from '$lib/api.js';
	import { buscasSalvas } from '$lib/buscas.js';
	import { prepararPublicacao } from '$lib/publicar-store.js';
	import { usuario } from '$lib/firebase.js';
	import { brl, pct } from '$lib/formatters.js';
	import { Tabs, Loading, Alert } from '$lib/components/ui/index.js';
	import FormAdicionarLoja from '$lib/components/FormAdicionarLoja.svelte';
	import PainelAlertas from '$lib/components/PainelAlertas.svelte';
	import ListaProdutosLoja from '$lib/components/ListaProdutosLoja.svelte';
	import GerenciarBuscas from '$lib/components/GerenciarBuscas.svelte';

	let buscasComLojas = $derived(($buscasSalvas ?? []).filter((b) => b.shop_ids?.length > 0));
	let buscaSelecionada = $state(null);
	let aba = $state('produtos');

	// Produtos da loja
	let produtos = $state([]);
	let carregandoProdutos = $state(false);
	let erroProdutos = $state(null);

	// Novidades
	let novidades = $state(null);
	let carregandoNovidades = $state(false);
	let erroNovidades = $state(null);

	let tabs = $derived([
		{ id: 'produtos', label: 'Produtos', badge: produtos.length > 0 ? String(produtos.length) : '' },
		{
			id: 'novidades',
			label: '🆕 Novidades',
			badge: novidades?.produtos_novos?.length ? String(novidades.produtos_novos.length) : '',
			badgeVariant: novidades?.produtos_novos?.length ? 'alert' : ''
		},
		{ id: 'precos', label: '📉 Preços', badge: novidades?.variacoes?.length ? String(novidades.variacoes.length) : '' }
	]);

	onMount(() => {
		buscasSalvas.sincronizarDoServidor();
	});

	function handleLojaAdicionada(r) {
		setTimeout(() => {
			const nova = buscasComLojas.find((b) => b.id === r.id);
			if (nova) selecionarBusca(nova);
		}, 100);
	}

	async function handleRemoverLoja(b) {
		if (!confirm(`Remover monitoramento da loja "${b.id}"?`)) return;
		try {
			await removerLoja(b.id);
			await buscasSalvas.sincronizarDoServidor();
			if (buscaSelecionada?.id === b.id) buscaSelecionada = null;
		} catch (e) {
			alert('Erro ao remover: ' + e.message);
		}
	}

	async function selecionarBusca(b) {
		buscaSelecionada = b;
		aba = 'produtos';
		await carregarProdutos();
		carregarNovidades();
	}

	async function carregarProdutos() {
		if (!buscaSelecionada) return;
		carregandoProdutos = true;
		erroProdutos = null;

		let timeoutId;
		const timeout = new Promise((_, reject) => {
			timeoutId = setTimeout(() => reject(new Error('A busca de produtos demorou demais. Tente novamente.')), 25000);
		});

		try {
			const r = await Promise.race([
				buscarCandidatos({
					fonte: 'shopee-shop',
					shopIds: buscaSelecionada.shop_ids,
					keyword: buscaSelecionada.keywords?.[0] ?? '',
					categoria: buscaSelecionada.categoria,
					estrategia: buscaSelecionada.estrategia ?? 'nicho',
					top: 50,
					semFiltro: true
				}),
				timeout
			]);
			produtos = r?.candidatos ?? [];
		} catch (e) {
			erroProdutos = e.message;
		} finally {
			clearTimeout(timeoutId);
			carregandoProdutos = false;
		}
	}

	async function carregarNovidades() {
		if (!buscaSelecionada) return;
		carregandoNovidades = true;
		erroNovidades = null;

		let timeoutId;
		const timeout = new Promise((_, reject) => {
			timeoutId = setTimeout(() => reject(new Error('A análise de novidades demorou demais.')), 25000);
		});

		try {
			novidades = await Promise.race([buscarNovidades({ buscaId: buscaSelecionada.id, dias: 7 }), timeout]);
		} catch (e) {
			erroNovidades = e.message;
			novidades = null;
		} finally {
			clearTimeout(timeoutId);
			carregandoNovidades = false;
		}
	}

	function irParaPublicar(c) {
		goto(prepararPublicacao(c));
	}
</script>

<svelte:head>
	<title>Lojas — Garimpei</title>
</svelte:head>

<section class="max-w-[900px]">
	<h1 class="text-2xl mb-1">🏪 Lojas Monitoradas</h1>
	<p class="text-tinta-suave text-sm mb-6">
		Acompanhe os produtos das lojas que você monitora. Veja novidades, variações de preço e publique ofertas
		diretamente.
	</p>

	{#if !$usuario}
		<div class="bg-porcelana p-4 rounded-sm text-tinta-suave">Faça login para ver as lojas monitoradas.</div>
	{:else}
		<FormAdicionarLoja onadicionada={handleLojaAdicionada} />
		<GerenciarBuscas />
		<PainelAlertas {buscaSelecionada} />

		{#if buscasComLojas.length === 0}
			<div class="bg-porcelana p-4 rounded-sm text-tinta-suave">
				<p>Nenhuma loja monitorada ainda.</p>
				<p class="text-sm mt-2 text-tinta-suave">Use o formulário acima para adicionar uma loja Shopee.</p>
			</div>
		{:else}
			<!-- Lista de buscas com lojas -->
			<div class="flex flex-wrap gap-3 mb-5">
				{#each buscasComLojas as b (b.id)}
					<div class="relative">
						<button
							class="border border-border bg-card rounded-sm px-4 py-3 pr-8 cursor-pointer text-left flex flex-col gap-0.5 hover:border-ouro {buscaSelecionada?.id === b.id ? 'border-ouro bg-ouro-fundo' : ''}"
							onclick={() => selecionarBusca(b)}
						>
							<strong class="text-sm">{b.nome || b.id}</strong>
							{#if b.cron}
								<span class="text-xs text-tinta-suave">⏱ coleta automática</span>
							{/if}
						</button>
						<button
							class="absolute top-1 right-1 w-[22px] h-[22px] rounded-full border-none bg-transparent text-tinta-suave text-xs cursor-pointer flex items-center justify-center hover:bg-erro-fundo hover:text-erro"
							onclick={() => handleRemoverLoja(b)}
							title="Remover monitoramento"
						>✕</button>
					</div>
				{/each}
			</div>

			{#if buscaSelecionada}
				<Tabs {tabs} bind:active={aba}>
					{#if aba === 'produtos'}
						<ListaProdutosLoja
							{produtos}
							carregando={carregandoProdutos}
							erro={erroProdutos}
							onpublicar={irParaPublicar}
						/>
					{:else if aba === 'novidades'}
						{#if carregandoNovidades}
							<Loading mensagem="Analisando novidades…" />
						{:else if erroNovidades}
							<Alert variant="error">{erroNovidades}</Alert>
						{:else if !novidades || novidades.produtos_novos?.length === 0}
							<p class="text-sm text-tinta-suave py-4">Nenhum produto novo detectado nos últimos {novidades?.dias_janela ?? 7} dias.</p>
						{:else}
							<p class="text-sm mb-4">
								<strong>{novidades.produtos_novos.length}</strong> produto(s) novo(s) detectado(s) nos últimos {novidades.dias_janela}
								dias.
							</p>
							<div class="flex flex-col gap-3">
								{#each novidades.produtos_novos as p (p.produto_id)}
									<div class="flex gap-3 px-4 py-3 border border-sucesso-borda border-l-[3px] border-l-sucesso rounded-sm bg-sucesso-fundo">
										<div class="text-xl">🆕</div>
										<div class="flex-1">
											<strong class="text-sm">{p.nome}</strong>
											<div class="flex gap-3 text-xs text-tinta-suave mt-0.5">
												<span>{brl(p.preco)}</span>
												<span>{pct(p.comissao)} comissão</span>
												<span>{p.vendas} vendas</span>
											</div>
											<span class="text-[0.72rem] text-tinta-suave">Detectado: {p.detectado_em?.split('T')[0]}</span>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					{:else if aba === 'precos'}
						{#if carregandoNovidades}
							<Loading mensagem="Analisando variações…" />
						{:else if !novidades || novidades.variacoes?.length === 0}
							<p class="text-sm text-tinta-suave py-4">
								Nenhuma variação de preço detectada nos últimos {novidades?.dias_janela ?? 7} dias.
							</p>
						{:else}
							<p class="text-sm mb-4">
								<strong>{novidades.variacoes.length}</strong> variação(ões) de preço detectada(s).
							</p>
							<div class="overflow-x-auto">
								<table class="w-full border-collapse text-[0.85rem]">
									<thead>
										<tr>
											<th class="text-left font-semibold px-2.5 py-2 border-b-2 border-border text-xs uppercase text-tinta-suave">Produto</th>
											<th class="text-left font-semibold px-2.5 py-2 border-b-2 border-border text-xs uppercase text-tinta-suave">Antes</th>
											<th class="text-left font-semibold px-2.5 py-2 border-b-2 border-border text-xs uppercase text-tinta-suave">Agora</th>
											<th class="text-left font-semibold px-2.5 py-2 border-b-2 border-border text-xs uppercase text-tinta-suave">Variação</th>
											<th class="text-left font-semibold px-2.5 py-2 border-b-2 border-border text-xs uppercase text-tinta-suave">Detectado</th>
											<th class="text-left font-semibold px-2.5 py-2 border-b-2 border-border text-xs uppercase text-tinta-suave"></th>
										</tr>
									</thead>
									<tbody>
										{#each novidades.variacoes as v (v.produto_id)}
											<tr>
												<td class="px-2.5 py-2 border-b border-border max-w-[200px] overflow-hidden text-ellipsis whitespace-nowrap">{v.nome}</td>
												<td class="px-2.5 py-2 border-b border-border">{brl(v.preco_anterior)}</td>
												<td class="px-2.5 py-2 border-b border-border font-bold {v.variacao_pct < 0 ? 'text-sucesso' : v.variacao_pct > 0 ? 'text-erro' : ''}">{brl(v.preco_atual)}</td>
												<td class="px-2.5 py-2 border-b border-border font-bold">
													<span
														class="inline-block px-2 py-0.5 rounded-full text-xs font-bold {v.variacao_pct < 0 ? 'bg-sucesso-fundo text-sucesso' : 'bg-erro-fundo text-erro'}"
													>
														{v.variacao_pct < 0 ? '↓' : '↑'}
														{Math.abs(v.variacao_pct * 100).toFixed(1)}%
													</span>
												</td>
												<td class="px-2.5 py-2 border-b border-border text-xs text-tinta-suave">{v.detectado_em?.split('T')[0]}</td>
												<td class="px-2.5 py-2 border-b border-border">
													<button
														class="border border-border bg-porcelana rounded-lg w-9 h-9 flex items-center justify-center cursor-pointer text-base shrink-0 hover:border-rosa hover:bg-[color-mix(in_srgb,var(--rosa)_8%,white)]"
														onclick={() =>
															irParaPublicar({
																id: v.produto_id,
																nome: v.nome,
																preco: v.preco_atual
															})}
														title="Publicar esta oferta">📤</button
													>
												</td>
											</tr>
										{/each}
									</tbody>
								</table>
							</div>
						{/if}
					{/if}
				</Tabs>
			{/if}
		{/if}
	{/if}
</section>
