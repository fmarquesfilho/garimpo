<script>
	/**
	 * PainelNovidades — exibe produtos novos e variações de preço coletados
	 * pelo Scheduler para uma busca agendada por keyword (sem loja).
	 *
	 * Props:
	 *   buscaId  — UUID da busca no servidor
	 *   keywords — array de keywords para exibir no header
	 */
	import { onMount } from 'svelte';
	import { buscarNovidades } from '$lib/api.js';
	import { Badge, Loading, EmptyState } from '$lib/components/ui';
	import ProductCard from './ProductCard.svelte';

	let { buscaId, keywords = [] } = $props();

	let carregando = $state(true);
	let produtosNovos = $state([]);
	let variacoes = $state([]);
	let erro = $state(null);

	async function carregar() {
		carregando = true;
		erro = null;
		try {
			const r = await buscarNovidades({ buscaId, dias: 7 });
			produtosNovos = r?.produtos_novos ?? [];
			variacoes = r?.variacoes ?? [];
		} catch {
			erro = 'Falha ao carregar novidades.';
			produtosNovos = [];
			variacoes = [];
		} finally {
			carregando = false;
		}
	}

	onMount(carregar);

	// Recarrega quando buscaId muda
	$effect(() => {
		if (buscaId) carregar();
	});

	let totalResultados = $derived(produtosNovos.length + variacoes.length);
</script>

<div class="mt-4 rounded-md border border-border bg-card p-4">
	<div class="mb-3 flex items-center gap-2">
		<h3 class="m-0 text-base font-semibold text-foreground">📊 Resultados da coleta</h3>
		{#each keywords.slice(0, 3) as kw}
			<Badge variant="secondary">{kw}</Badge>
		{/each}
		{#if keywords.length > 3}
			<Badge variant="secondary">+{keywords.length - 3}</Badge>
		{/if}
	</div>

	{#if carregando}
		<Loading />
	{:else if erro}
		<p class="text-sm text-destructive">{erro}</p>
	{:else if totalResultados === 0}
		<EmptyState
			icone="⏳"
			mensagem="Aguardando primeira coleta"
			dica="Os resultados aparecerão aqui após a próxima execução do agendamento."
		/>
	{:else}
		{#if produtosNovos.length > 0}
			<div class="mb-4">
				<h4 class="mb-2 text-sm font-semibold text-foreground">
					🆕 Produtos novos <Badge>{produtosNovos.length}</Badge>
				</h4>
				<div class="flex flex-col gap-2">
					{#each produtosNovos.slice(0, 10) as p (p.produto_id ?? p.id)}
						<ProductCard
							produto={{
								nome: p.nome,
								preco: p.preco,
								comissao: p.comissao,
								vendas: p.vendas ?? 0,
								imagem: p.imagem,
								link: p.link,
								loja: p.loja
							}}
							layout="compact"
							variacao={{ tipo: 'novo', detectado_em: p.detectado_em }}
						/>
					{/each}
					{#if produtosNovos.length > 10}
						<p class="text-sm text-muted-foreground">
							+{produtosNovos.length - 10} produtos novos não exibidos
						</p>
					{/if}
				</div>
			</div>
		{/if}

		{#if variacoes.length > 0}
			<div>
				<h4 class="mb-2 text-sm font-semibold text-foreground">
					📉 Variações de preço <Badge>{variacoes.length}</Badge>
				</h4>
				<div class="flex flex-col gap-2">
					{#each variacoes.slice(0, 10) as v (v.produto_id ?? v.id)}
						<ProductCard
							produto={{
								nome: v.nome,
								preco: v.preco_atual ?? v.preco,
								comissao: v.comissao,
								vendas: v.vendas ?? 0,
								imagem: v.imagem,
								link: v.link,
								loja: v.loja
							}}
							layout="compact"
							variacao={{
								tipo: v.variacao_pct < 0 ? 'queda' : 'alta',
								pct: v.variacao_pct,
								preco_anterior: v.preco_primeiro ?? v.preco_anterior,
								preco_atual: v.preco_atual ?? v.preco,
								detectado_em: v.detectado_em
							}}
						/>
					{/each}
					{#if variacoes.length > 10}
						<p class="text-sm text-muted-foreground">
							+{variacoes.length - 10} variações não exibidas
						</p>
					{/if}
				</div>
			</div>
		{/if}
	{/if}
</div>
