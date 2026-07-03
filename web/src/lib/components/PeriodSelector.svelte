<script>
	/**
	 * PeriodSelector — seletor de período tipo segmented control.
	 * Usa tokens do design system para visual consistente.
	 */
	let { value = $bindable(7), options = [1, 3, 7, 14, 30] } = $props();
</script>

<div class="flex gap-0.5 rounded-sm border border-border bg-porcelana p-[3px]" role="radiogroup" aria-label="Período">
	{#each options as dias}
		<button
			class="cursor-pointer rounded-sm border-none bg-transparent px-4 py-2 font-[var(--ui)] text-sm font-semibold text-tinta-suave transition-[background,color,box-shadow] duration-150 ease-linear focus-visible:outline-2 focus-visible:outline-ouro focus-visible:outline-offset-2 motion-reduce:transition-none"
			class:!bg-[var(--branco)]={value === dias}
			class:!text-foreground={value === dias}
			class:!shadow-sm={value === dias}
			onclick={() => (value = dias)}
			role="radio"
			aria-checked={value === dias}
		>
			{dias === 1 ? 'Hoje' : `${dias} dias`}
		</button>
	{/each}
</div>
