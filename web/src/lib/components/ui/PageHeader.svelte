<script>
	/**
	 * PageHeader — cabeçalho padrão de página.
	 * @prop rotulo — eyebrow text (opcional)
	 * @prop titulo — título principal (h1)
	 * @prop subtitulo — descrição abaixo do título (opcional)
	 */
	let { rotulo = '', titulo, subtitulo = '', ...rest } = $props();
</script>

<section class="mb-8 max-w-[42rem]" {...rest}>
	{#if rotulo}
		<p class="text-xs font-semibold uppercase tracking-widest text-muted-foreground">{rotulo}</p>
	{/if}
	<h1 class="my-2 mb-4 font-display text-[clamp(1.8rem,5vw,2.5rem)]">{titulo}</h1>
	{#if subtitulo}
		<p class="m-0 mb-4 leading-relaxed text-muted-foreground">{subtitulo}</p>
	{/if}
</section>
