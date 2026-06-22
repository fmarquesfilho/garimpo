<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { buscarCandidatos, buscarNovidades } from '$lib/api.js';
	import { buscasSalvas } from '$lib/buscas.js';
	import { usuario } from '$lib/firebase.js';

	let buscasComLojas = $derived(($buscasSalvas ?? []).filter(b => b.shop_ids?.length > 0));
	let buscaSelecionada = $state(null);
	let aba = $state('produtos'); // 'produtos' | 'novidades' | 'precos'

	// Produtos da loja
	let produtos = $state([]);
	let carregandoProdutos = $state(false);
	let erroProdutos = $state(null);

	// Novidades
	let novidades = $state(null);
	let carregandoNovidades = $state(false);
	let erroNovidades = $state(null);

	onMount(() => buscasSalvas.sincronizarDoServidor());

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
		try {
			const r = await buscarCandidatos({
				fonte: 'shopee-shop',
				shopIds: buscaSelecionada.shop_ids,
				keyword: buscaSelecionada.keywords?.[0] ?? '',
				categoria: buscaSelecionada.categoria,
				estrategia: buscaSelecionada.estrategia ?? 'nicho',
				top: 50,
				comissaoMin: buscaSelecionada.comissao_min,
				vendasMin: buscaSelecionada.vendas_min,
				notaMin: buscaSelecionada.nota_min
			});
			produtos = r?.candidatos ?? [];
		} catch (e) {
			erroProdutos = e.message;
		} finally {
			carregandoProdutos = false;
		}
	}

	async function carregarNovidades() {
		if (!buscaSelecionada) return;
		carregandoNovidades = true;
		erroNovidades = null;
		try {
			const r = await buscarNovidades({ buscaId: buscaSelecionada.id, dias: 7 });
			novidades = r;
		} catch (e) {
			erroNovidades = e.message;
			novidades = null;
		} finally {
			carregandoNovidades = false;
		}
	}

	function irParaPublicar(c) {
		const dados = encodeURIComponent(JSON.stringify(c));
		goto(`/publicar?dados=${dados}`);
	}

	const brl = (v) => v?.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' }) ?? '';
	const pct = (v) => `${(v * 100).toFixed(1)}%`;
</script>

<svelte:head>
	<title>Lojas — Garimpo</title>
</svelte:head>

<section class="lojas-page">
	<h1>🏪 Lojas Monitoradas</h1>
	<p class="subtitulo">
		Acompanhe os produtos das lojas que você monitora. Veja novidades, variações de preço
		e publique ofertas diretamente.
	</p>

	{#if !$usuario}
		<div class="aviso">Faça login para ver as lojas monitoradas.</div>
	{:else if buscasComLojas.length === 0}
		<div class="vazio">
			<p>Nenhuma busca com lojas configurada.</p>
			<p class="dica">Na página de <a href="/">Curadoria</a>, crie uma nova busca e adicione IDs de lojas no campo "🏪 Lojas Shopee".</p>
		</div>
	{:else}
		<!-- Lista de buscas com lojas -->
		<div class="lojas-lista">
			{#each buscasComLojas as b (b.id)}
				<button
					class="loja-card"
					class:ativa={buscaSelecionada?.id === b.id}
					onclick={() => selecionarBusca(b)}
				>
					<strong>{b.id}</strong>
					<span class="loja-meta">
						🏪 {b.shop_ids.length} {b.shop_ids.length === 1 ? 'loja' : 'lojas'}
						{#if b.keywords?.length > 0}
							· 🔑 {b.keywords.join(', ')}
						{/if}
					</span>
				</button>
			{/each}
		</div>

		{#if buscaSelecionada}
			<!-- Abas -->
			<nav class="abas">
				<button class:ativa={aba === 'produtos'} onclick={() => (aba = 'produtos')}>
					Produtos {#if produtos.length > 0}<span class="badge-n">{produtos.length}</span>{/if}
				</button>
				<button class:ativa={aba === 'novidades'} onclick={() => (aba = 'novidades')}>
					🆕 Novidades {#if novidades?.produtos_novos?.length}<span class="badge-n alerta">{novidades.produtos_novos.length}</span>{/if}
				</button>
				<button class:ativa={aba === 'precos'} onclick={() => (aba = 'precos')}>
					📉 Preços {#if novidades?.variacoes?.length}<span class="badge-n">{novidades.variacoes.length}</span>{/if}
				</button>
			</nav>

			{#if aba === 'produtos'}
				{#if carregandoProdutos}
					<p class="loading">Buscando produtos da loja…</p>
				{:else if erroProdutos}
					<div class="msg-erro">{erroProdutos}</div>
				{:else if produtos.length === 0}
					<p class="vazio-tab">Nenhum produto encontrado. A coleta periódica pode ainda não ter rodado.</p>
				{:else}
					<div class="grade-produtos">
						{#each produtos as p, i (p.id)}
							<div class="card-produto-loja">
								{#if p.imagem}
									<img src={p.imagem} alt={p.nome} class="prod-thumb" />
								{/if}
								<div class="prod-info">
									<h4>{p.nome}</h4>
									<div class="prod-dados">
										<span class="prod-preco">{brl(p.preco)}</span>
										<span class="prod-comissao">{pct(p.comissao)}</span>
										<span class="prod-vendas">{p.vendas} vendas</span>
										<span class="prod-nota">★ {p.avaliacao?.toFixed(1)}</span>
									</div>
									<div class="prod-score">teor: {p.score?.toFixed(3)}</div>
								</div>
								<button class="btn-pub-mini" onclick={() => irParaPublicar(p)} title="Publicar este produto">
									📤
								</button>
							</div>
						{/each}
					</div>
				{/if}

			{:else if aba === 'novidades'}
				{#if carregandoNovidades}
					<p class="loading">Analisando novidades…</p>
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
					<p class="loading">Analisando variações…</p>
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
								</tr>
							</thead>
							<tbody>
								{#each novidades.variacoes as v (v.produto_id)}
									<tr class:baixou={v.variacao_pct < 0} class:subiu={v.variacao_pct > 0}>
										<td class="nome-col">{v.nome}</td>
										<td>{brl(v.preco_anterior)}</td>
										<td class="preco-atual">{brl(v.preco_atual)}</td>
										<td class="variacao">
											{v.variacao_pct > 0 ? '↑' : '↓'}
											{Math.abs(v.variacao_pct * 100).toFixed(1)}%
										</td>
										<td class="data">{v.detectado_em?.split('T')[0]}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			{/if}
		{/if}
	{/if}
</section>

<style>
	.lojas-page { max-width: 900px; }
	h1 { font-size: 1.5rem; margin-bottom: 0.25rem; }
	.subtitulo { color: var(--tinta-suave); font-size: 0.9rem; margin-bottom: var(--r6); }

	.aviso, .vazio { background: var(--porcelana); padding: var(--r4); border-radius: 10px; color: var(--tinta-suave); }
	.vazio a { color: var(--ouro); text-decoration: underline; }
	.dica { font-size: 0.85rem; margin-top: 4px; }
	.vazio-tab { color: var(--tinta-suave); font-size: 0.9rem; font-style: italic; }
	.msg-erro { background: #fef2f2; color: #b91c1c; padding: var(--r3) var(--r4); border-radius: 8px; margin-bottom: var(--r4); }
	.loading { color: var(--tinta-suave); font-style: italic; }

	/* Lista de buscas com lojas */
	.lojas-lista { display: flex; flex-wrap: wrap; gap: var(--r3); margin-bottom: var(--r5); }
	.loja-card {
		border: 1px solid var(--linha); background: var(--nevoa);
		border-radius: 10px; padding: var(--r3) var(--r4);
		cursor: pointer; text-align: left; display: flex; flex-direction: column; gap: 2px;
	}
	.loja-card:hover { border-color: var(--ouro); }
	.loja-card.ativa { border-color: var(--ouro); background: var(--ouro-fundo); }
	.loja-card strong { font-size: 0.9rem; }
	.loja-meta { font-size: 0.78rem; color: var(--tinta-suave); }

	/* Abas */
	.abas { display: flex; gap: 2px; margin-bottom: var(--r5); border-bottom: 2px solid var(--linha); }
	.abas button {
		padding: 8px 16px; border: none; background: transparent;
		font-weight: 600; font-size: 0.85rem; color: var(--tinta-suave);
		cursor: pointer; border-bottom: 2px solid transparent; margin-bottom: -2px;
		display: flex; align-items: center; gap: 6px;
	}
	.abas button.ativa { color: var(--tinta); border-bottom-color: var(--ouro); }
	.badge-n {
		font-size: 0.7rem; background: var(--ouro-fundo); color: #7a5a1e;
		padding: 1px 6px; border-radius: 999px; font-weight: 700;
	}
	.badge-n.alerta { background: #fef2f2; color: #b91c1c; }

	/* Grade de produtos */
	.grade-produtos { display: flex; flex-direction: column; gap: var(--r3); }
	.card-produto-loja {
		display: flex; gap: var(--r3); padding: var(--r3) var(--r4);
		border: 1px solid var(--linha); border-radius: 10px; background: white;
		align-items: center;
	}
	.prod-thumb { width: 56px; height: 56px; border-radius: 8px; object-fit: cover; flex-shrink: 0; }
	.prod-info { flex: 1; min-width: 0; }
	.prod-info h4 { font-size: 0.9rem; margin: 0 0 4px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
	.prod-dados { display: flex; flex-wrap: wrap; gap: var(--r2); font-size: 0.78rem; color: var(--tinta-suave); }
	.prod-preco { font-weight: 700; color: var(--ouro); }
	.prod-comissao { font-weight: 600; }
	.prod-score { font-size: 0.72rem; color: var(--tinta-suave); margin-top: 2px; }
	.btn-pub-mini {
		border: 1px solid var(--linha); background: var(--porcelana);
		border-radius: 8px; width: 36px; height: 36px;
		display: flex; align-items: center; justify-content: center;
		cursor: pointer; font-size: 1rem; flex-shrink: 0;
	}
	.btn-pub-mini:hover { border-color: var(--rosa); background: color-mix(in srgb, var(--rosa) 8%, white); }

	/* Novidades */
	.info-novidades { font-size: 0.88rem; margin-bottom: var(--r4); }
	.grade-novidades { display: flex; flex-direction: column; gap: var(--r3); }
	.card-novidade {
		display: flex; gap: var(--r3); padding: var(--r3) var(--r4);
		border: 1px solid #bbf7d0; border-left: 3px solid #22c55e;
		border-radius: 10px; background: #f0fdf4;
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
	tr.baixou .variacao { color: #16a34a; }
	tr.baixou .preco-atual { color: #16a34a; }
	tr.subiu .variacao { color: #dc2626; }
	tr.subiu .preco-atual { color: #dc2626; }
	.data { font-size: 0.78rem; color: var(--tinta-suave); }
</style>
