<script>
	/**
	 * Aba de favoritos — lista produtos salvos com ações de publicar e remover.
	 */
	import { favoritos } from '$lib/favoritos.js';
	import { goto } from '$app/navigation';
	import ProductCard from './ProductCard.svelte';
	import { EmptyState } from '$lib/components/ui/index.js';

	function publicar(produto) {
		const dados = encodeURIComponent(JSON.stringify(produto));
		goto(`/publicar?dados=${dados}`);
	}
</script>

{#if $favoritos.length === 0}
	<EmptyState
		icone="⭐"
		mensagem="Nenhum produto favoritado ainda."
		dica="Clique no ⭐ em qualquer produto para salvá-lo aqui."
	/>
{:else}
	<p class="contagem">{$favoritos.length} {$favoritos.length === 1 ? 'produto salvo' : 'produtos salvos'}</p>
	<div class="grade">
		{#each $favoritos as fav (fav.produto_id || fav.id)}
			<ProductCard
				produto={fav}
				layout="compact"
				onpublicar={publicar}
				onfavoritar={(p) => favoritos.remover(p.produto_id || p.id)}
			/>
		{/each}
	</div>
{/if}

<style>
	.contagem { font-size: 0.82rem; color: var(--tinta-suave); margin-bottom: var(--r4); }
	.grade { display: flex; flex-direction: column; gap: var(--r3); }
</style>
