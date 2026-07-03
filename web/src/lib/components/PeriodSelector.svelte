<script>
	/**
	 * PeriodSelector — seletor de período tipo segmented control.
	 * Usa tokens do design system para visual consistente.
	 */
	let { value = $bindable(7), options = [1, 3, 7, 14, 30] } = $props();
</script>

<div class="periodo-selector" role="radiogroup" aria-label="Período">
	{#each options as dias}
		<button
			class:ativo={value === dias}
			onclick={() => (value = dias)}
			role="radio"
			aria-checked={value === dias}
		>
			{dias === 1 ? 'Hoje' : `${dias} dias`}
		</button>
	{/each}
</div>

<style>
	.periodo-selector {
		display: flex;
		gap: 2px;
		background: var(--porcelana);
		border-radius: var(--raio-sm);
		padding: 3px;
		border: 1px solid var(--linha);
	}
	button {
		padding: var(--r2) var(--r4);
		border: none;
		border-radius: var(--raio-sm);
		background: transparent;
		font-family: var(--ui);
		font-size: var(--text-sm);
		font-weight: var(--font-semi);
		color: var(--tinta-suave);
		cursor: pointer;
		transition: background 0.15s ease, color 0.15s ease, box-shadow 0.15s ease;
	}
	button:focus-visible {
		outline: 2px solid var(--ouro);
		outline-offset: 2px;
	}
	button.ativo {
		background: var(--branco);
		color: var(--tinta);
		box-shadow: var(--sombra);
	}
	button:hover:not(.ativo) {
		color: var(--tinta);
	}

	@media (prefers-reduced-motion: reduce) {
		button { transition-duration: 0ms; }
	}
</style>
