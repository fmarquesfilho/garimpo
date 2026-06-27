<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { buscarNovidades } from '$lib/api.js';
	import { buscasSalvas } from '$lib/buscas.js';
	import { usuario } from '$lib/firebase.js';
	import PeriodSelector from '$lib/components/PeriodSelector.svelte';
	import ProductCard from '$lib/components/ProductCard.svelte';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import { PageHeader, Loading, EmptyState } from '$lib/components/ui/index.js';

	let dias = $state(7);
	let carregando = $state(true);
	let erro = $state(null);

	let quedas = $state([]);
	let altas = $state([]);
	let novos = $state([]);

	let buscasComLojas = $derived(($buscasSalvas ?? []).filter(b => b.shop_ids?.length > 0));
	let nomesLojas = $derived(Object.fromEntries(
		buscasComLojas.map(b => [b.id, b.nome || b.id])
	));

	onMount(async () => {
		await buscasSalvas.sincronizarDoServidor();
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

		let timeoutId;
		const timeout = new Promise((_, reject) => {
			timeoutId = setTimeout(() => reject(new Error('A análise demorou demais. Tente novamente ou reduza o período.')), 30000);
		});

		try {
			const promises = buscasComLojas.map(b =>
				buscarNovidades({ buscaId: b.id, dias })
					.then(r => ({ ...r, loja: b.id }))
					.catch(() => null)
			);
			const resultados = await Promise.race([Promise.all(promises), timeout]);

			const novasQuedas = [];
			const novasAltas = [];
			const novosItens = [];

			for (const r of resultados) {
				if (!r) continue;
				for (const v of (r.variacoes ?? [])) {
					const item = { ...v, loja: r.loja };
					if (v.variacao_pct < 0) novasQuedas.push(item);
					else novasAltas.push(item);
				}
				for (const p of (r.produtos_novos ?? [])) {
					novosItens.push({ ...p, loja: r.loja });
				}
			}

			novasQuedas.sort((a, b) => a.variacao_pct - b.variacao_pct);
			novasAltas.sort((a, b) => b.variacao_pct - a.variacao_pct);
			novosItens.sort((a, b) => (b.detectado_em ?? '').localeCompare(a.detectado_em ?? ''));

			quedas = novasQuedas;
			altas = novasAltas;
			novos = novosItens;
		} catch (e) {
			erro = e.message ?? e;
		} finally {
			clearTimeout(timeoutId);
			carregando = false;
		}
	}

	function irParaPublicar(item) {
		let link = item.link ?? '';
		if (!link && item.loja && item.produto_id) {
			const shopId = item.loja.replace('loja-', '');
			if (/^\d+$/.test(shopId)) {
				link = `https://shopee.com.br/product-i.${shopId}.${item.produto_id}`;
			}
		}

		const dados = encodeURIComponent(JSON.stringify({
			id: item.produto_id,
			nome: item.nome,
			preco: item.preco_atual ?? item.preco,
			comissao: item.comissao ?? 0,
			link,
			imagem: item.imagem ?? '',
			categoria: item.categoria ?? '',
			loja: item.loja ?? '',
			vendas: item.vendas ?? 0,
			avaliacao: item.nota ?? 0
		}));
		goto(`/publicar?dados=${dados}`);
	}

	// Recarrega quando o período muda (após a carga inicial)
	let primeiroRender = true;
	$effect(() => {
		dias; // track
		if (primeiroRender) { primeiroRender = false; return; }
		carregar();
	});
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
			<PeriodSelector bind:value={dias} options={[1, 3, 7, 14]} />
			<span class="meta">{buscasComLojas.length} {buscasComLojas.length === 1 ? 'loja' : 'lojas'} monitoradas</span>
		</div>

		{#if carregando}
			<Loading mensagem="Analisando variações de todas as lojas… (pode levar até 30s)" />
		{:else if erro}
			<ErrorMessage erro={{ message: erro, retry: true }} onretry={carregar} />
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
					<div class="resumo-item queda" title="Produtos que caíram de preço acima do threshold configurado">
						<span class="resumo-numero">{quedas.length}</span>
						<span class="resumo-label">↓ Quedas de preço</span>
					</div>
				{/if}
				{#if altas.length > 0}
					<div class="resumo-item alta" title="Produtos que subiram de preço — pode indicar fim de promoção">
						<span class="resumo-numero">{altas.length}</span>
						<span class="resumo-label">↑ Altas de preço</span>
					</div>
				{/if}
				{#if novos.length > 0}
					<div class="resumo-item novo" title="Produtos que apareceram pela primeira vez nas lojas monitoradas">
						<span class="resumo-numero">{novos.length}</span>
						<span class="resumo-label">🆕 Novos no catálogo</span>
					</div>
				{/if}
			</div>

			{#if quedas.length > 0}
				<section class="secao-feed">
					<h2>📉 Quedas de preço</h2>
					<div class="feed">
						{#each quedas as item (item.produto_id + item.loja)}
							<ProductCard
								produto={{ nome: item.nome, preco: item.preco_atual, imagem: item.imagem, link: item.link, comissao: item.comissao ?? 0, vendas: item.vendas ?? 0, produto_id: item.produto_id, loja: item.loja }}
								layout="feed"
								nomeLoja={nomesLojas[item.loja] ?? item.loja}
								variacao={{ tipo: 'queda', pct: item.variacao_pct, preco_anterior: item.preco_anterior, preco_atual: item.preco_atual, detectado_em: item.detectado_em }}
								onpublicar={() => irParaPublicar(item)}
							/>
						{/each}
					</div>
				</section>
			{/if}

			{#if novos.length > 0}
				<section class="secao-feed">
					<h2>🆕 Produtos novos nas lojas</h2>
					<p class="sub-secao">Apareceram pela primeira vez no catálogo das lojas monitoradas.</p>
					<div class="feed">
						{#each novos.slice(0, 20) as item (item.produto_id + item.loja)}
							<ProductCard
								produto={{ nome: item.nome, preco: item.preco, imagem: item.imagem, link: item.link, comissao: item.comissao ?? 0, vendas: item.vendas ?? 0, produto_id: item.produto_id, loja: item.loja }}
								layout="feed"
								nomeLoja={nomesLojas[item.loja] ?? item.loja}
								variacao={{ tipo: 'novo', detectado_em: item.detectado_em }}
								onpublicar={() => irParaPublicar(item)}
							/>
						{/each}
					</div>
				</section>
			{/if}

			{#if altas.length > 0}
				<section class="secao-feed">
					<h2>📈 Altas de preço</h2>
					<p class="sub-secao">Produtos que subiram — pode indicar fim de promoção ou escassez.</p>
					<div class="feed">
						{#each altas.slice(0, 10) as item (item.produto_id + item.loja)}
							<ProductCard
								produto={{ nome: item.nome, preco: item.preco_atual, imagem: item.imagem, link: item.link, comissao: item.comissao ?? 0, vendas: item.vendas ?? 0, produto_id: item.produto_id, loja: item.loja }}
								layout="feed"
								nomeLoja={nomesLojas[item.loja] ?? item.loja}
								variacao={{ tipo: 'alta', pct: item.variacao_pct, preco_anterior: item.preco_anterior, preco_atual: item.preco_atual, detectado_em: item.detectado_em }}
							/>
						{/each}
					</div>
				</section>
			{/if}
		{/if}
	{/if}
</section>

<style>
	.oportunidades-page { max-width: 900px; }

	.controles {
		display: flex; align-items: center; justify-content: space-between;
		margin-bottom: var(--r6); flex-wrap: wrap; gap: var(--r3);
	}
	.meta { font-size: 0.78rem; color: var(--tinta-suave); }
	.msg-erro { background: var(--erro-fundo); color: var(--erro-texto); padding: var(--r3) var(--r4); border-radius: 8px; }

	/* Resumo rápido */
	.resumo-rapido { display: flex; gap: var(--r3); margin-bottom: var(--r6); }
	.resumo-item {
		display: flex; align-items: center; gap: 6px;
		padding: 8px 16px; border-radius: var(--raio-sm);
		font-weight: 600; font-size: 0.85rem;
	}
	.resumo-item.queda { background: var(--sucesso-fundo); color: var(--sucesso-texto); }
	.resumo-item.alta { background: var(--erro-fundo); color: var(--erro-texto); }
	.resumo-item.novo { background: var(--ouro-fundo); color: var(--ouro-escuro); }
	.resumo-numero { font-size: 1.3rem; font-weight: 700; }

	/* Seções */
	.secao-feed { margin-bottom: var(--r8); }
	.secao-feed h2 { font-size: 1.2rem; margin-bottom: var(--r3); }
	.sub-secao { font-size: 0.82rem; color: var(--tinta-suave); margin-bottom: var(--r4); }
	.feed { display: flex; flex-direction: column; gap: var(--r3); }

	@media (max-width: 600px) {
		.resumo-rapido { flex-wrap: wrap; }
		.controles { flex-direction: column; align-items: stretch; }
	}
</style>
