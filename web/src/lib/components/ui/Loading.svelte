<script>
	/**
	 * Loading — indicador de carregamento simples.
	 * @prop mensagem — texto exibido (default: "Carregando…")
	 */
	let { mensagem = 'Carregando…', ...rest } = $props();
</script>

<p class="italic text-tinta-suave" {...rest}>{mensagem}</p>
