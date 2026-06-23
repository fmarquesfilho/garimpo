<script>
	/**
	 * Select — dropdown padronizado.
	 * @prop value — bind:value para two-way
	 * @prop label — rótulo acima (opcional)
	 * @prop options — array de { value, label } ou passado via children (slot)
	 */
	let { value = $bindable(''), label = '', children } = $props();
</script>

<label class="select-wrapper">
	{#if label}
		<span class="select-label">{label}</span>
	{/if}
	<select class="select-field" bind:value>
		{@render children()}
	</select>
</label>

<style>
	.select-wrapper {
		display: flex;
		flex-direction: column;
		gap: 5px;
	}
	.select-label {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--tinta-suave);
	}
	.select-field {
		font-family: var(--mono);
		font-size: 0.9rem;
		padding: 9px 12px;
		border-radius: 10px;
		border: 1px solid var(--linha);
		background: var(--porcelana);
		color: var(--tinta);
		cursor: pointer;
	}
	.select-field:focus { outline: 2px solid var(--ouro); outline-offset: 1px; }
</style>
