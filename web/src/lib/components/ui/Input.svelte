<script>
	/**
	 * Input — campo de entrada padronizado.
	 * @prop value — bind:value para two-way
	 * @prop type — 'text' | 'search' | 'number' | 'url' | 'email' | 'datetime-local'
	 * @prop placeholder
	 * @prop label — rótulo acima do input (opcional)
	 * @prop variant — 'default' | 'mono'
	 * @prop size — 'sm' | 'md' | 'lg'
	 */
	const VARIANTS = ['default', 'mono'];
	const SIZES = ['sm', 'md', 'lg'];

	let {
		value = $bindable(''),
		type = 'text',
		placeholder = '',
		label = '',
		variant = 'default',
		size = 'md',
		disabled = false,
		...rest
	} = $props();

	let resolvedVariant = $derived(VARIANTS.includes(variant) ? variant : 'default');
	let resolvedSize = $derived(SIZES.includes(size) ? size : 'md');
</script>

<label class="input-wrapper">
	{#if label}
		<span class="input-label">{label}</span>
	{/if}
	<input
		class="input-field {resolvedVariant} size-{resolvedSize}"
		{type}
		{placeholder}
		{disabled}
		bind:value
		{...rest}
	/>
</label>

<style>
	.input-wrapper {
		display: flex;
		flex-direction: column;
		gap: var(--r1);
	}
	.input-label {
		font-size: var(--text-xs);
		font-weight: var(--font-semi);
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--tinta-suave);
	}
	.input-field {
		font-family: var(--ui);
		font-size: var(--text-base);
		padding: var(--r3) var(--r4);
		border-radius: var(--raio-sm);
		border: 1px solid var(--linha);
		background: var(--branco);
		color: var(--tinta);
		width: 100%;
		transition: border-color 0.15s ease, box-shadow 0.15s ease;
	}
	.input-field::placeholder {
		color: var(--tinta-suave);
		opacity: 0.7;
	}
	.input-field:focus {
		outline: none;
		border-color: var(--ouro);
		box-shadow: 0 0 0 2px var(--ouro-fundo);
	}
	.input-field:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	/* Variant: mono */
	.mono {
		font-family: var(--mono);
		font-variant-numeric: tabular-nums;
	}

	/* Sizes */
	.size-sm { font-size: var(--text-sm); padding: var(--r2) var(--r3); }
	.size-md { font-size: var(--text-base); padding: var(--r3) var(--r4); }
	.size-lg { font-size: var(--text-md); padding: var(--r4) var(--r5); }

	@media (prefers-reduced-motion: reduce) {
		.input-field { transition-duration: 0ms; }
	}
</style>
