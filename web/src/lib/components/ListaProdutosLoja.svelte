<script>
	/**
	 * Lista de produtos de uma loja monitorada.
	 * Usa ProductCard no layout compact.
	 */
	import ProductCard from '$lib/components/ProductCard.svelte';
	import { Loading, Alert } from '$lib/components/ui/index.js';

	let { produtos = [], carregando = false, erro = null, onpublicar = null } = $props();
</script>

{#if carregando}
	<Loading mensagem="Buscando produtos da loja…" />
{:else if erro}
	<Alert variant="error">{erro}</Alert>
{:else if produtos.length === 0}
	<p class="vazio-tab">Nenhum produto encontrado. A coleta periódica pode ainda não ter rodado.</p>
{:else}
	<div class="grade-produtos">
		{#each produtos as p (p.id)}
			<ProductCard produto={p} layout="compact" onpublicar={onpublicar} />
		{/each}
	</div>
{/if}

<style>
	.vazio-tab {
		color: var(--tinta-suave);
		font-size: var(--text-base);
		font-style: italic;
	}
	.grade-produtos {
		display: flex;
		flex-direction: column;
		gap: var(--r3);
	}
</style>
