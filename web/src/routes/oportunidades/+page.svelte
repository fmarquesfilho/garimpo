<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { buscarNovidades } from '$lib/api.js';
	import { buscasSalvas } from '$lib/buscas.js';
	import { usuario } from '$lib/firebase.js';
	import { brl, pct, apenasData, tempoAtras } from '$lib/formatters.js';
	import { PageHeader, Loading, EmptyState } from '$lib/components/ui/index.js';

	let dias = $state(7);
	let carregando = $state(true);
	let erro = $state(null);

	// Dados agregados de todas as lojas
	let quedas = $state([]);
	let altas = $state([]);
	let novos = $state([]);

	let buscasComLojas = $derived(($buscasSalvas ?? []).filter(b => b.shop_ids?.length > 0));

	// Mapa de busca_id → nome amigável
	let nomesLojas = $derived(Object.fromEntries(
		buscasComLojas.map(b => [b.id, b.nome || b.id])
	));

	onMount(async () => {
		await buscasSalvas.sincronizarDoServidor();
		// Aguarda um tick para o $derived atualizar
		await new Promise(r => setTimeout(r, 50));
		carregar();
	});

	async function carregar() {
		if (buscasComLojas.length === 0) {
			carregando = false;
			return;
		}
		carregando = true;
		erro = null;

		try {
			// Busca novidades de TODAS as lojas em paralelo
			const promises = buscasComLojas.map(b =>
				buscarNovidades({ buscaId: b.id, dias })
					.then(r => ({ ...r, loja: b.id }))
					.catch(() => null)
			);
			const resultados = await Promise.all(promises);

			const novasQuedas = [];
			const novasAltas = [];
			const novosItens = [];

			for (const r of resultados) {
				if (!r) continue;
				for (const v of (r.variacoes ?? [])) {
					const item = { ...v, loja: r.loja };
					if (v.variacao_pct < 0) {
						novasQuedas.push(item);
					} else {
						novasAltas.push(item);
					}
				}
				for (const p of (r.produtos_novos ?? [])) {
					novosItens.push({ ...p, loja: r.loja });
				}
			}

			// Ordena por magnitude (maiores variações primeiro)
			novasQuedas.sort((a, b) => a.variacao_pct - b.variacao_pct);
			novasAltas.sort((a, b) => b.variacao_pct - a.variacao_pct);
			novosItens.sort((a, b) => (b.detectado_em ?? '').localeCompare(a.detectado_em ?? ''));

			quedas = novasQuedas;
			altas = novasAltas;
			novos = novosItens;
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}
	}

	function irParaPublicar(item) {
		const dados = encodeURIComponent(JSON.stringify({
			id: item.produto_id,
			nome: item.nome,
			preco: item.preco_atual ?? item.preco,
			comissao: item.comissao ?? 0
		}));
		goto(`/publicar?dados=${dados}`);
	}

	function mudarPeriodo(novoDias) {
		dias = novoDias;
		carregar();
	}
</script>

<svelte:head>
	<title>Oportunidades — Garimpei</title>
</svelte:head>

<section class="oportunidades-page">
	<PageHeader
		rotulo="monitoramento de lojas"
		titulo="🎯 Oportunidades"
		subtitulo="Quedas de preço e produtos novos das suas lojas monitoradas. Atualizações a cada coleta."
	/>

	{#if !$usuario}
		<div class="msg-erro">Faça login para ver oportunidades.</div>
	{:else}
		<div class="controles">
			<div class="filtro-periodo">
				<button class:ativo={dias === 1} onclick={() => mudarPeriodo(1)}>Hoje</button>
				<button class:ativo={dias === 3} onclick={() => mudarPeriodo(3)}>3 dias</button>
				<button class:ativo={dias === 7} onclick={() => mudarPeriodo(7)}>7 dias</button>
				<button class:ativo={dias === 14} onclick={() => mudarPeriodo(14)}>14 dias</button>
			</div>
			<span class="meta">{buscasComLojas.length} {buscasComLojas.length === 1 ? 'loja' : 'lojas'} monitoradas</span>
		</div>

		{#if carregando}
			<Loading mensagem="Analisando variações de todas as lojas…" />
		{:else if erro}
			<div class="msg-erro">{erro}</div>
		{:else if quedas.length === 0 && altas.length === 0 && novos.length === 0}
			<EmptyState
				icone="📭"
				mensagem="Nenhuma variação de preço ou produto novo detectado nos últimos {dias} dias."
				dica='As coletas rodam a cada 4h. Adicione mais lojas em <a href="/lojas">Lojas</a> para ampliar o monitoramento.'
			/>
		{:else}
			<!-- Resumo rápido -->
			<div class="resumo-rapido">
				{#if quedas.length > 0}
					<div class="resumo-item queda">
						<span class="resumo-numero">{quedas.length}</span>
						<span class="resumo-label">↓ Quedas</span>
					</div>
				{/if}
				{#if altas.length > 0}
					<div class="resumo-item alta">
						<span class="resumo-numero">{altas.length}</span>
						<span class="resumo-label">↑ Altas</span>
					</div>
				{/if}
				{#if novos.length > 0}
					<div class="resumo-item novo">
						<span class="resumo-numero">{novos.length}</span>
						<span class="resumo-label">🆕 Novos</span>
					</div>
				{/if}
			</div>

			<!-- Quedas de preço (oportunidades!) -->
			{#if quedas.length > 0}
				<section class="secao-feed">
					<h2>📉 Quedas de preço</h2>
					<div class="feed">
						{#each quedas as item (item.produto_id + item.loja)}
							<div class="card-oportunidade queda">
								<div class="card-header">
									<span class="badge-variacao badge-queda">↓ {Math.abs(item.variacao_pct * 100).toFixed(0)}%</span>
									<span class="loja-tag">{nomesLojas[item.loja] ?? item.loja}</span>
									<span class="tempo">{tempoAtras(item.detectado_em)}</span>
								</div>
								<h3 class="card-nome">{item.nome}</h3>
								<div class="card-precos">
									<span class="preco-antes">{brl(item.preco_anterior)}</span>
									<span class="seta">→</span>
									<span class="preco-atual destaque-queda">{brl(item.preco_atual)}</span>
									<span class="economia">(-{brl(item.preco_anterior - item.preco_atual)})</span>
								</div>
								<div class="card-acoes">
									<button class="btn-publicar" onclick={() => irParaPublicar(item)}>
										📤 Publicar
									</button>
								</div>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			<!-- Produtos novos -->
			{#if novos.length > 0}
				<section class="secao-feed">
					<h2>🆕 Produtos novos nas lojas</h2>
					<p class="sub-secao">Apareceram pela primeira vez no catálogo das lojas monitoradas.</p>
					<div class="feed">
						{#each novos.slice(0, 20) as item (item.produto_id + item.loja)}
							<div class="card-oportunidade novo">
								<div class="card-header">
									<span class="badge-novo">Novo</span>
									<span class="loja-tag">{nomesLojas[item.loja] ?? item.loja}</span>
									<span class="tempo">{tempoAtras(item.detectado_em)}</span>
								</div>
								<h3 class="card-nome">{item.nome}</h3>
								<div class="card-precos">
									<span class="preco-atual">{brl(item.preco)}</span>
									{#if item.comissao > 0}
										<span class="comissao">{pct(item.comissao)} comissão</span>
									{/if}
									{#if item.vendas > 0}
										<span class="vendas">{item.vendas} vendas</span>
									{/if}
								</div>
								<div class="card-acoes">
									<button class="btn-publicar" onclick={() => irParaPublicar(item)}>
										📤 Publicar
									</button>
								</div>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			<!-- Altas de preço (informativo) -->
			{#if altas.length > 0}
				<section class="secao-feed">
					<h2>📈 Altas de preço</h2>
					<p class="sub-secao">Produtos que subiram — pode indicar fim de promoção ou escassez.</p>
					<div class="feed">
						{#each altas.slice(0, 10) as item (item.produto_id + item.loja)}
							<div class="card-oportunidade alta">
								<div class="card-header">
									<span class="badge-variacao badge-alta">↑ {Math.abs(item.variacao_pct * 100).toFixed(0)}%</span>
									<span class="loja-tag">{nomesLojas[item.loja] ?? item.loja}</span>
									<span class="tempo">{tempoAtras(item.detectado_em)}</span>
								</div>
								<h3 class="card-nome">{item.nome}</h3>
								<div class="card-precos">
									<span class="preco-antes">{brl(item.preco_anterior)}</span>
									<span class="seta">→</span>
									<span class="preco-atual destaque-alta">{brl(item.preco_atual)}</span>
								</div>
							</div>
						{/each}
					</div>
				</section>
			{/if}
		{/if}
	{/if}
</section>

<style>
	.oportunidades-page { max-width: 800px; }

	.controles {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--r6);
		flex-wrap: wrap;
		gap: var(--r3);
	}
	.filtro-periodo {
		display: flex;
		gap: 2px;
		background: var(--porcelana);
		border-radius: var(--raio-sm);
		padding: 3px;
		border: 1px solid var(--linha);
	}
	.filtro-periodo button {
		padding: 6px 14px;
		border: none;
		border-radius: var(--raio-sm);
		background: transparent;
		font-size: 0.82rem;
		font-weight: 600;
		color: var(--tinta-suave);
		cursor: pointer;
	}
	.filtro-periodo button.ativo {
		background: var(--branco);
		color: var(--tinta);
		box-shadow: 0 1px 3px rgba(0,0,0,0.08);
	}
	.meta { font-size: 0.78rem; color: var(--tinta-suave); }

	/* Resumo rápido */
	.resumo-rapido {
		display: flex;
		gap: var(--r3);
		margin-bottom: var(--r6);
	}
	.resumo-item {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 16px;
		border-radius: var(--raio-sm);
		font-weight: 600;
		font-size: 0.85rem;
	}
	.resumo-item.queda { background: var(--sucesso-fundo); color: var(--sucesso-texto); }
	.resumo-item.alta { background: var(--erro-fundo); color: var(--erro-texto); }
	.resumo-item.novo { background: var(--ouro-fundo); color: var(--ouro-escuro); }
	.resumo-numero { font-size: 1.3rem; font-weight: 700; }

	/* Seções */
	.secao-feed { margin-bottom: var(--r8); }
	.secao-feed h2 { font-size: 1.2rem; margin-bottom: var(--r3); }
	.sub-secao { font-size: 0.82rem; color: var(--tinta-suave); margin-bottom: var(--r4); }

	/* Feed de cards */
	.feed { display: flex; flex-direction: column; gap: var(--r3); }

	.card-oportunidade {
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r4);
		background: var(--branco);
		transition: border-color 0.15s;
	}
	.card-oportunidade:hover { border-color: var(--ouro-claro); }
	.card-oportunidade.queda { border-left: 3px solid var(--sucesso-texto); }
	.card-oportunidade.alta { border-left: 3px solid var(--erro-texto); }
	.card-oportunidade.novo { border-left: 3px solid var(--ouro); }

	.card-header {
		display: flex;
		align-items: center;
		gap: var(--r2);
		margin-bottom: 6px;
	}
	.badge-variacao {
		padding: 2px 8px;
		border-radius: var(--raio-full);
		font-size: 0.72rem;
		font-weight: 700;
	}
	.badge-queda { background: var(--sucesso-fundo); color: var(--sucesso-texto); }
	.badge-alta { background: var(--erro-fundo); color: var(--erro-texto); }
	.badge-novo {
		padding: 2px 8px;
		border-radius: var(--raio-full);
		font-size: 0.72rem;
		font-weight: 700;
		background: var(--ouro-fundo);
		color: var(--ouro-escuro);
	}
	.loja-tag {
		font-size: 0.72rem;
		color: var(--tinta-suave);
		background: var(--porcelana);
		padding: 1px 6px;
		border-radius: 4px;
	}
	.tempo { font-size: 0.72rem; color: var(--tinta-suave); margin-left: auto; }

	.card-nome {
		font-size: 0.95rem;
		font-weight: 600;
		margin: 0 0 8px;
		line-height: 1.3;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.card-precos {
		display: flex;
		align-items: center;
		gap: var(--r2);
		font-size: 0.88rem;
		flex-wrap: wrap;
	}
	.preco-antes { text-decoration: line-through; color: var(--tinta-suave); }
	.seta { color: var(--tinta-suave); font-size: 0.8rem; }
	.preco-atual { font-weight: 700; }
	.destaque-queda { color: var(--sucesso-texto); }
	.destaque-alta { color: var(--erro-texto); }
	.economia { font-size: 0.78rem; color: var(--sucesso-texto); font-weight: 600; }
	.comissao { font-size: 0.78rem; color: var(--ouro); font-weight: 600; }
	.vendas { font-size: 0.78rem; color: var(--tinta-suave); }

	.card-acoes { margin-top: 10px; }
	.btn-publicar {
		padding: 6px 14px;
		border: 1px solid var(--ouro-claro);
		background: var(--ouro-fundo);
		color: var(--ouro-escuro);
		border-radius: var(--raio-sm);
		font-size: 0.82rem;
		font-weight: 600;
		cursor: pointer;
	}
	.btn-publicar:hover { background: var(--ouro-claro); }

	@media (max-width: 600px) {
		.resumo-rapido { flex-wrap: wrap; }
		.controles { flex-direction: column; align-items: stretch; }
	}
</style>
