<script>
	import { goto } from '$app/navigation';
	import { favoritos } from '$lib/favoritos.js';
	import { prepararPublicacao } from '$lib/publicar-store.js';
	import ProductCard from '$lib/components/ProductCard.svelte';
	import BuscaUnificada from '$lib/components/BuscaUnificada.svelte';
	import { Loading, EmptyState, Button } from '$lib/components/ui/index.js';

	let carregando = $state(false);
	let erro = $state(null);
	let resultados = $state([]);

	function publicar(c) {
		goto(prepararPublicacao(c));
	}
</script>

<svelte:head>
	<title>Garimpar — Garimpei</title>
</svelte:head>

<section class="max-w-[900px] space-y-4">
	<div>
		<h1 class="mb-2 text-[clamp(1.8rem,5vw,2.5rem)]">O que publicar hoje?</h1>
		<p class="text-[0.95rem] text-tinta-suave">Busque produtos, monitore lojas e publique com um clique.</p>
	</div>

	<BuscaUnificada
		onresultados={(r) => (resultados = r)}
		oncarregando={(v) => (carregando = v)}
		onerro={(e) => (erro = e)}
	/>

	<!-- Resultados -->
	{#if carregando}
		<Loading mensagem="Buscando produtos…" />
	{:else if erro}
		<div
			class="msg-erro rounded-md border border-[color-mix(in_srgb,var(--erro-texto)_30%,var(--linha))] bg-card p-5 text-center"
		>
			<p class="my-2"><strong>😕 {erro.message ?? erro}</strong></p>
			<Button size="sm" onclick={() => {}}>🔄 Tentar novamente</Button>
		</div>
	{:else if resultados.length === 0}
		<EmptyState
			icone="🔍"
			mensagem="Nenhum resultado com os filtros atuais."
			dica="Ajuste keywords, fontes ou filtros acima."
		/>
	{:else}
		<p class="contagem text-[0.82rem] text-tinta-suave">
			{resultados.length}
			{resultados.length === 1 ? 'produto' : 'produtos'}
		</p>
		<div class="grade grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
			{#each resultados as produto, i (produto.id || produto.produto_id || i)}
				<ProductCard
					{produto}
					posicao={produto._fonte === 'curadoria' ? i + 1 : null}
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
