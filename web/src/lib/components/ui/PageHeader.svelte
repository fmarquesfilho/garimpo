<script>
	/**
	 * PageHeader — cabeçalho padrão de página.
	 * @prop rotulo — eyebrow text (opcional)
	 * @prop titulo — título principal (h1)
	 * @prop subtitulo — descrição abaixo do título (opcional)
	 */
	let { rotulo = '', titulo, subtitulo = '' } = $props();
</script>

<section class="page-header">
	{#if rotulo}
		<p class="rotulo">{rotulo}</p>
	{/if}
	<h1>{titulo}</h1>
	{#if subtitulo}
		<p class="sub">{subtitulo}</p>
	{/if}
</section>

<style>
	.page-header { max-width: 42rem; margin-bottom: var(--r8); }
	h1 { font-size: clamp(1.8rem, 5vw, 2.5rem); margin: var(--r2) 0 var(--r4); }
	.sub { color: var(--tinta-suave); margin: 0 0 var(--r4); font-size: 0.9rem; line-height: 1.5; }
</style>
