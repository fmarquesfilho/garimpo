<script>
	/**
	 * CategoriaCard — categoria adicionada ao filtro. Mostra o nome e os
	 * marketplaces a que a categoria pertence.
	 * @prop nome — nome da categoria
	 * @prop marketplaces — string[] de marketplaces
	 * @prop onremover — () => void
	 */
	let { nome = '', marketplaces = [], onremover = null } = $props();
</script>

<div class="relative min-w-[150px] rounded-sm border border-border bg-muted px-3 py-2.5">
	<button
		type="button"
		class="absolute right-1.5 top-1.5 rounded px-1 text-sm leading-none text-muted-foreground hover:text-destructive"
		onclick={() => onremover?.()}
		aria-label="Remover categoria {nome}">✕</button
	>
	<div class="pr-4 text-sm font-bold text-foreground">{nome}</div>
	{#if marketplaces.length}
		<div class="mt-1.5 flex flex-wrap gap-1">
			{#each marketplaces as m (m)}
				<span
					class="rounded border border-border bg-card px-1.5 py-px font-[var(--mono)] text-[0.68rem] text-muted-foreground"
					>{m}</span
				>
			{/each}
		</div>
	{/if}
</div>
