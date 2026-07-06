<script>
	/**
	 * EmptyState — mensagem para estados vazios/sem dados.
	 * @prop icone — emoji ou ícone (opcional)
	 * @prop mensagem — texto principal
	 * @prop dica — texto secundário (opcional)
	 */
	let { icone = '', mensagem, dica = '', ...rest } = $props();
</script>

<div class="rounded-md border border-border bg-muted p-8 text-center text-muted-foreground" {...rest}>
	{#if icone}<span class="mb-3 block text-3xl">{icone}</span>{/if}
	<p class="m-0 text-base">{mensagem}</p>
	{#if dica}
		<!-- eslint-disable-next-line svelte/no-at-html-tags -->
		<p class="mt-2 text-sm text-muted-foreground [&_a]:text-primary [&_a]:underline">{@html dica}</p>
	{/if}
</div>
