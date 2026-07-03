<script>
	/**
	 * DashPanel — painel genérico para dashboard (título + conteúdo).
	 * @prop titulo — título do painel
	 */
	let { titulo = '', children, ...rest } = $props();
</script>

<div class="rounded-md border border-border bg-card p-4" {...rest}>
	{#if titulo}
		<h2 class="m-0 mb-3 text-base font-semibold">{titulo}</h2>
	{/if}
	{@render children()}
</div>
