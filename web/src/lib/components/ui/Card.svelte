<script>
	/**
	 * Card — container com borda e sombra padrão.
	 * @prop variant — 'default' | 'highlight' | 'success' | 'error'
	 * @prop padding — 'sm' | 'md' | 'lg'
	 */
	let { variant = 'default', padding = 'md', children } = $props();
</script>

<div class="card {variant} pad-{padding}">
	{@render children()}
</div>

<style>
	.card {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		box-shadow: var(--sombra);
	}

	.pad-sm { padding: var(--r3); }
	.pad-md { padding: var(--r4) var(--r6); }
	.pad-lg { padding: var(--r6) var(--r8); }

	.highlight { border-color: var(--ouro-claro); background: linear-gradient(180deg, #fffaf1, var(--nevoa)); }
	.success { border-color: #bbf7d0; background: #f0fdf4; }
	.error { border-color: #fecaca; background: #fef2f2; }
</style>
