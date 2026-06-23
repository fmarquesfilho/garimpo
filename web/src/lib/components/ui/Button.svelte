<script>
	/**
	 * Button — componente padrão de botão.
	 * @prop variant — 'primary' | 'secondary' | 'danger' | 'ghost'
	 * @prop size — 'sm' | 'md' | 'lg'
	 * @prop disabled
	 * @prop type — 'button' | 'submit'
	 */
	let {
		variant = 'primary',
		size = 'md',
		disabled = false,
		type = 'button',
		onclick = null,
		children
	} = $props();
</script>

<button
	class="btn {variant} {size}"
	{type}
	{disabled}
	{onclick}
>
	{@render children()}
</button>

<style>
	.btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		font-family: var(--ui);
		font-weight: 600;
		border: 1px solid transparent;
		border-radius: 10px;
		cursor: pointer;
		transition: background 0.15s ease, border-color 0.15s ease, opacity 0.15s ease;
		white-space: nowrap;
	}
	.btn:disabled { opacity: 0.5; cursor: not-allowed; }

	/* Sizes */
	.sm { font-size: 0.8rem; padding: 6px 12px; }
	.md { font-size: 0.88rem; padding: 10px 18px; }
	.lg { font-size: 0.95rem; padding: 12px 28px; }

	/* Variants */
	.primary { background: var(--ouro); color: white; }
	.primary:hover:not(:disabled) { background: #a3782f; }

	.secondary { background: var(--porcelana); border-color: var(--linha); color: var(--tinta); }
	.secondary:hover:not(:disabled) { border-color: var(--ouro); color: var(--ouro); }

	.danger { background: var(--rosa); color: white; }
	.danger:hover:not(:disabled) { background: #8f4c62; }

	.ghost { background: transparent; color: var(--tinta-suave); border-color: transparent; }
	.ghost:hover:not(:disabled) { color: var(--tinta); background: var(--porcelana); }
</style>
