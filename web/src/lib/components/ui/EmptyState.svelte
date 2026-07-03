<script>
	/**
	 * EmptyState — mensagem para estados vazios/sem dados.
	 * @prop icone — emoji ou ícone (opcional)
	 * @prop mensagem — texto principal
	 * @prop dica — texto secundário (opcional)
	 */
	let { icone = '', mensagem, dica = '', ...rest } = $props();
</script>

<div class="vazio" {...rest}>
	{#if icone}<span class="icone">{icone}</span>{/if}
	<p class="mensagem">{mensagem}</p>
	{#if dica}
		<!-- eslint-disable-next-line svelte/no-at-html-tags -->
		<p class="dica">{@html dica}</p>
	{/if}
</div>

<style>
	.vazio {
		background: var(--porcelana);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r8);
		text-align: center;
		color: var(--tinta-suave);
	}
	.icone {
		font-size: 2rem;
		display: block;
		margin-bottom: var(--r3);
	}
	.mensagem {
		margin: 0;
		font-size: var(--text-md);
	}
	.dica {
		font-size: var(--text-sm);
		margin-top: var(--r2);
		color: var(--tinta-suave);
	}
	.dica :global(a) {
		color: var(--ouro);
		text-decoration: underline;
	}
</style>
