<script>
	/**
	 * Input — campo de entrada padronizado.
	 * @prop value — bind:value para two-way
	 * @prop type — 'text' | 'search' | 'number' | 'url' | 'datetime-local'
	 * @prop placeholder
	 * @prop label — rótulo acima do input (opcional)
	 * @prop variant — 'default' | 'mono' (para números)
	 */
	let {
		value = $bindable(''),
		type = 'text',
		placeholder = '',
		label = '',
		variant = 'default',
		disabled = false,
		onkeydown = null
	} = $props();
</script>

<label class="input-wrapper">
	{#if label}
		<span class="input-label">{label}</span>
	{/if}
	<input
		class="input-field"
		class:mono={variant === 'mono'}
		{type}
		{placeholder}
		{disabled}
		{onkeydown}
		bind:value
	/>
</label>

<style>
	.input-wrapper {
		display: flex;
		flex-direction: column;
		gap: 5px;
	}
	.input-label {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--tinta-suave);
	}
	.input-field {
		font-family: var(--ui);
		font-size: 0.95rem;
		padding: 9px 12px;
		border-radius: 10px;
		border: 1px solid var(--linha);
		background: var(--porcelana);
		color: var(--tinta);
		width: 100%;
	}
	.input-field::placeholder { color: var(--tinta-suave); opacity: 0.7; }
	.input-field:focus { outline: 2px solid var(--ouro); outline-offset: 1px; }
	.input-field:disabled { opacity: 0.5; }
	.input-field.mono { font-family: var(--mono); }
</style>
