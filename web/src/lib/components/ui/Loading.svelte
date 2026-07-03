<script>
	/**
	 * Loading — indicador de carregamento simples.
	 * @prop mensagem — texto exibido (default: "Carregando…")
	 */
	let { mensagem = 'Carregando…', ...rest } = $props();
</script>

<p class="loading" {...rest}>{mensagem}</p>

<style>
	.loading {
		color: var(--tinta-suave);
		font-style: italic;
		font-size: var(--text-base);
	}
</style>
