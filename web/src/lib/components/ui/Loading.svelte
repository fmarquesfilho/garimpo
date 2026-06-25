<script>
	/**
	 * Loading — indicador de carregamento simples.
	 * @prop mensagem — texto exibido (default: "Carregando…")
	 */
	let { mensagem = 'Carregando…' } = $props();
</script>

<p class="loading">{mensagem}</p>

<style>
	.loading {
		color: var(--tinta-suave);
		font-style: italic;
		font-size: 0.9rem;
	}
</style>
