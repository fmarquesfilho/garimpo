<script>
	/**
	 * PageHeader — cabeçalho padrão de página.
	 * @prop rotulo — eyebrow text (opcional)
	 * @prop titulo — título principal (h1)
	 * @prop subtitulo — descrição abaixo do título (opcional)
	 */
	let { rotulo = '', titulo, subtitulo = '', ...rest } = $props();
</script>

<section class="page-header" {...rest}>
	{#if rotulo}
		<p class="rotulo">{rotulo}</p>
	{/if}
	<h1>{titulo}</h1>
	{#if subtitulo}
		<p class="sub">{subtitulo}</p>
	{/if}
</section>

<style>
	.page-header {
		max-width: 42rem;
		margin-bottom: var(--r8);
	}
	.rotulo {
		font-size: var(--text-xs);
		font-weight: var(--font-semi);
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--tinta-suave);
	}
	h1 {
		font-family: var(--display);
		font-size: clamp(1.8rem, 5vw, 2.5rem);
		margin: var(--r2) 0 var(--r4);
	}
	.sub {
		color: var(--tinta-suave);
		margin: 0 0 var(--r4);
		font-size: var(--text-base);
		line-height: 1.5;
	}
</style>
