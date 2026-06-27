<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { buscarCandidatos, buscarNovidades, removerLoja } from '$lib/api.js';
	import { buscasSalvas } from '$lib/buscas.js';
	import { usuario } from '$lib/firebase.js';
	import { brl, pct } from '$lib/formatters.js';
	import { TabBar, Loading } from '$lib/components/ui/index.js';
	import FormAdicionarLoja from '$lib/components/FormAdicionarLoja.svelte';
	import PainelAlertas from '$lib/components/PainelAlertas.svelte';
	import ListaProdutosLoja from '$lib/components/ListaProdutosLoja.svelte';

	let buscasComLojas = $derived(($buscasSalvas ?? []).filter(b => b.shop_ids?.length > 0));
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
		{ id: 'novidades', label: '🆕 Novidades', badge: novidades?.produtos_novos?.length ? String(novidades.produtos_novos.length) : '', badgeVariant: novidades?.produtos_novos?.length ? 'alert' : '' },
		{ id: 'precos', label: '📉 Preços', badge: novidades?.variacoes?.length ? String(novidades.variacoes.length) : '' }
	]);

	onMount(() => { buscasSalvas.sincronizarDoServidor(); });

	function handleLojaAdicionada(r) {
		setTimeout(() => {
			const nova = buscasComLojas.find(b => b.id === r.id);
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
			const r = await Promise.race([buscarCandidatos({
				fonte: 'shopee-shop',
				shopIds: buscaSelecionada.shop_ids,
				keyword: buscaSelecionada.keywords?.[0] ?? '',
				categoria: buscaSelecionada.categoria,
				estrategia: buscaSelecionada.estrategia ?? 'nicho',
				top: 50,
				semFiltro: true
			}), timeout]);
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
		const dados = encodeURIComponent(JSON.stringify(c));
		goto(`/publicar?dados=${dados}`);
	}
</script>

<svelte:head>
	<title>Lojas — Garimpei</title>
</svelte:head>

<section class="lojas-page">
	<h1>🏪 Lojas Monitoradas</h1>
	<p class="subtitulo">
		Acompanhe os produtos das lojas que você monitora. Veja novidades, variações de preço
		e publique ofertas diretamente.
	</p>

	{#if !$usuario}
		<div class="aviso">Faça login para ver as lojas monitoradas.</div>
	{:else}
		<FormAdicionarLoja onadicionada={handleLojaAdicionada} />
		<PainelAlertas {buscaSelecionada} />

		{#if buscasComLojas.length === 0}
			<div class="vazio">
				<p>Nenhuma loja monitorada ainda.</p>
				<p class="dica">Use o formulário acima para adicionar uma loja Shopee.</p>
			</div>
		{:else}
			<!-- Lista de buscas com lojas -->
			<div class="lojas-lista">
				{#each buscasComLojas as b (b.id)}
					<div class="loja-card-wrapper">
						<button
							class="loja-card"
							class:ativa={buscaSelecionada?.id === b.id}
							onclick={() => selecionarBusca(b)}
						>
							<strong>{b.nome || b.id}</strong>
							{#if b.cron}
								<span class="loja-meta">⏱ coleta automática</span>
							{/if}
						</button>
						<button
							class="btn-remover"
							onclick={() => handleRemoverLoja(b)}
							title="Remover monitoramento"
						>✕</button>
					</div>
				{/each}
			</div>

			{#if buscaSelecionada}
				<TabBar {tabs} bind:active={aba} />

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
						<div class="msg-erro">{erroNovidades}</div>
					{:else if !novidades || novidades.produtos_novos?.length === 0}
						<p class="vazio-tab">Nenhum produto novo detectado nos últimos {novidades?.dias_janela ?? 7} dias.</p>
					{:else}
						<p class="info-novidades">
							<strong>{novidades.produtos_novos.length}</strong> produto(s) novo(s) detectado(s)
							nos últimos {novidades.dias_janela} dias.
						</p>
						<div class="grade-novidades">
							{#each novidades.produtos_novos as p (p.produto_id)}
								<div class="card-novidade">
									<div class="novidade-badge">🆕</div>
									<div class="novidade-info">
										<strong>{p.nome}</strong>
										<div class="novidade-dados">
											<span>{brl(p.preco)}</span>
											<span>{pct(p.comissao)} comissão</span>
											<span>{p.vendas} vendas</span>
										</div>
										<span class="novidade-data">Detectado: {p.detectado_em?.split('T')[0]}</span>
									</div>
								</div>
							{/each}
						</div>
					{/if}

				{:else if aba === 'precos'}
					{#if carregandoNovidades}
						<Loading mensagem="Analisando variações…" />
					{:else if !novidades || novidades.variacoes?.length === 0}
						<p class="vazio-tab">Nenhuma variação de preço detectada nos últimos {novidades?.dias_janela ?? 7} dias.</p>
					{:else}
						<p class="info-novidades">
							<strong>{novidades.variacoes.length}</strong> variação(ões) de preço detectada(s).
						</p>
						<div class="tabela-variacoes">
							<table>
								<thead>
									<tr>
										<th>Produto</th>
										<th>Antes</th>
										<th>Agora</th>
										<th>Variação</th>
										<th>Detectado</th>
										<th></th>
									</tr>
								</thead>
								<tbody>
									{#each novidades.variacoes as v (v.produto_id)}
										<tr class:baixou={v.variacao_pct < 0} class:subiu={v.variacao_pct > 0}>
											<td class="nome-col">{v.nome}</td>
											<td>{brl(v.preco_anterior)}</td>
											<td class="preco-atual">{brl(v.preco_atual)}</td>
											<td class="variacao">
												<span class="badge-variacao" class:badge-baixou={v.variacao_pct < 0} class:badge-subiu={v.variacao_pct > 0}>
													{v.variacao_pct < 0 ? '↓' : '↑'}
													{Math.abs(v.variacao_pct * 100).toFixed(1)}%
												</span>
											</td>
											<td class="data">{v.detectado_em?.split('T')[0]}</td>
											<td>
												<button class="btn-pub-mini" onclick={() => irParaPublicar({
													id: v.produto_id,
													nome: v.nome,
													preco: v.preco_atual
												})} title="Publicar esta oferta">📤</button>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				{/if}
			{/if}
		{/if}
	{/if}
</section>

<style>
	.lojas-page { max-width: 900px; }
	h1 { font-size: 1.5rem; margin-bottom: 0.25rem; }
	.subtitulo { color: var(--tinta-suave); font-size: 0.9rem; margin-bottom: var(--r6); }

	.aviso, .vazio { background: var(--porcelana); padding: var(--r4); border-radius: var(--raio-sm); color: var(--tinta-suave); }
	.dica { font-size: 0.85rem; margin-top: 4px; }
	.vazio-tab { color: var(--tinta-suave); font-size: 0.9rem; font-style: italic; }
	.msg-erro { background: var(--erro-fundo); color: var(--erro-texto); padding: var(--r3) var(--r4); border-radius: 8px; margin-bottom: var(--r4); }

	/* Lista de lojas */
	.lojas-lista { display: flex; flex-wrap: wrap; gap: var(--r3); margin-bottom: var(--r5); }
	.loja-card-wrapper { position: relative; }
	.loja-card {
		border: 1px solid var(--linha); background: var(--nevoa);
		border-radius: var(--raio-sm); padding: var(--r3) var(--r4);
		padding-right: 32px; cursor: pointer; text-align: left;
		display: flex; flex-direction: column; gap: 2px;
	}
	.loja-card:hover { border-color: var(--ouro); }
	.loja-card.ativa { border-color: var(--ouro); background: var(--ouro-fundo); }
	.loja-card strong { font-size: 0.9rem; }
	.loja-meta { font-size: 0.78rem; color: var(--tinta-suave); }
	.btn-remover {
		position: absolute; top: 4px; right: 4px;
		width: 22px; height: 22px; border-radius: 50%;
		border: none; background: transparent; color: var(--tinta-suave);
		font-size: 0.75rem; cursor: pointer;
		display: flex; align-items: center; justify-content: center;
	}
	.btn-remover:hover { background: var(--erro-fundo); color: var(--erro-texto); }

	/* Novidades */
	.info-novidades { font-size: 0.88rem; margin-bottom: var(--r4); }
	.grade-novidades { display: flex; flex-direction: column; gap: var(--r3); }
	.card-novidade {
		display: flex; gap: var(--r3); padding: var(--r3) var(--r4);
		border: 1px solid var(--sucesso-borda); border-left: 3px solid var(--sucesso-texto);
		border-radius: var(--raio-sm); background: var(--sucesso-fundo);
	}
	.novidade-badge { font-size: 1.2rem; }
	.novidade-info { flex: 1; }
	.novidade-info strong { font-size: 0.9rem; }
	.novidade-dados { display: flex; gap: var(--r3); font-size: 0.78rem; color: var(--tinta-suave); margin-top: 2px; }
	.novidade-data { font-size: 0.72rem; color: var(--tinta-suave); }

	/* Variações de preço */
	.tabela-variacoes { overflow-x: auto; }
	table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
	th { text-align: left; font-weight: 600; padding: 8px 10px; border-bottom: 2px solid var(--linha); font-size: 0.78rem; text-transform: uppercase; color: var(--tinta-suave); }
	td { padding: 8px 10px; border-bottom: 1px solid var(--linha); }
	.nome-col { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.preco-atual { font-weight: 700; }
	.variacao { font-weight: 700; }
	.badge-variacao { display: inline-block; padding: 2px 8px; border-radius: var(--raio-full); font-size: 0.78rem; font-weight: 700; }
	.badge-baixou { background: var(--sucesso-fundo); color: var(--sucesso-texto); }
	.badge-subiu { background: var(--erro-fundo); color: var(--erro-texto); }
	tr.baixou .preco-atual { color: var(--sucesso-texto); }
	tr.subiu .preco-atual { color: var(--erro-texto); }
	.data { font-size: 0.78rem; color: var(--tinta-suave); }
	.btn-pub-mini {
		border: 1px solid var(--linha); background: var(--porcelana);
		border-radius: 8px; width: 36px; height: 36px;
		display: flex; align-items: center; justify-content: center;
		cursor: pointer; font-size: 1rem; flex-shrink: 0;
	}
	.btn-pub-mini:hover { border-color: var(--rosa); background: color-mix(in srgb, var(--rosa) 8%, white); }
</style>
