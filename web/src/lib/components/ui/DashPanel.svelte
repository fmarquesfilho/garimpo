<script>
	/**
	 * DashPanel — painel genérico para dashboard (título + conteúdo).
	 * @prop titulo — título do painel
	 */
	let { titulo = '', children, ...rest } = $props();
</script>

<div class="painel" {...rest}>
	{#if titulo}
		<h2>{titulo}</h2>
	{/if}
	{@render children()}
</div>

<style>
	.painel {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r4);
	}
	h2 {
		font-size: var(--text-base);
		font-weight: var(--font-semi);
		margin: 0 0 var(--r3);
	}
</style>
