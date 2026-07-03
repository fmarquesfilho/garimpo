<script>
	/**
	 * Button — componente padrão de botão.
	 * @prop variant — 'primary' | 'secondary' | 'danger' | 'ghost'
	 * @prop size — 'sm' | 'md' | 'lg'
	 * @prop disabled
	 * @prop type — 'button' | 'submit' | 'reset'
	 */
	const VARIANTS = ['primary', 'secondary', 'danger', 'ghost'];
	const SIZES = ['sm', 'md', 'lg'];

	let {
		variant = 'primary',
		size = 'md',
		disabled = false,
		type = 'button',
		onclick = null,
		children,
		...rest
	} = $props();

	let resolvedVariant = $derived(VARIANTS.includes(variant) ? variant : 'primary');
	let resolvedSize = $derived(SIZES.includes(size) ? size : 'md');
</script>

<button
	class="btn {resolvedVariant} {resolvedSize}"
	{type}
	{disabled}
	{onclick}
	{...rest}
>
	{@render children()}
</button>

<style>
	.btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: var(--r2);
		font-family: var(--ui);
		font-weight: var(--font-semi);
		border: 1px solid transparent;
		border-radius: var(--raio-sm);
		cursor: pointer;
		white-space: nowrap;
		transition: background 0.15s ease, border-color 0.15s ease, opacity 0.15s ease;
	}
	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.btn:focus-visible {
		outline: 2px solid var(--ouro);
		outline-offset: 2px;
	}

	/* Sizes */
	.sm { font-size: var(--text-sm); padding: var(--r2) var(--r3); }
	.md { font-size: var(--text-base); padding: var(--r3) var(--r5); }
	.lg { font-size: var(--text-md); padding: var(--r3) var(--r8); }

	/* Variants */
	.primary { background: var(--ouro); color: var(--branco); }
	.primary:hover:not(:disabled) { background: var(--ouro-hover); }

	.secondary { background: var(--porcelana); border-color: var(--linha); color: var(--tinta); }
	.secondary:hover:not(:disabled) { border-color: var(--ouro); color: var(--ouro); }

	.danger { background: var(--rosa); color: var(--branco); }
	.danger:hover:not(:disabled) { background: var(--rosa-hover); }

	.ghost { background: transparent; color: var(--tinta-suave); }
	.ghost:hover:not(:disabled) { background: var(--porcelana); color: var(--tinta); }

	@media (prefers-reduced-motion: reduce) {
		.btn { transition-duration: 0ms; }
	}
</style>
