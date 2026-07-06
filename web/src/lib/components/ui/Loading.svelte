<script>
	/**
	 * Loading — indicador de carregamento simples.
	 * @prop mensagem — texto exibido (default: "Carregando…")
	 */
	let { mensagem = 'Carregando…', ...rest } = $props();
</script>

<p class="italic text-muted-foreground" {...rest}>{mensagem}</p>
