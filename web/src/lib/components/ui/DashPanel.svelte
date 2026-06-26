<script>
	/**
	 * DashPanel — painel genérico para dashboard (título + conteúdo).
	 * @prop titulo — título do painel
	 */
	let { titulo = '', children } = $props();
</script>

<div class="painel">
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
	h2 { font-size: 1rem; margin: 0 0 var(--r3); font-weight: 600; }
</style>
