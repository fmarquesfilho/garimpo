<script>
	/**
	 * Badge — etiqueta/pílula de informação.
	 * @prop variant — 'default' | 'gold' | 'pink' | 'green' | 'red'
	 */
	let { variant = 'default', children } = $props();
</script>

<span class="badge {variant}">
	{@render children()}
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		font-size: 0.72rem;
		font-weight: 700;
		padding: 2px 8px;
		border-radius: 999px;
		border: 1px solid var(--linha);
		background: var(--porcelana);
		color: var(--tinta-suave);
		white-space: nowrap;
	}
	.gold { background: var(--ouro-fundo); border-color: var(--ouro-claro); color: #7a5a1e; }
	.pink { background: color-mix(in srgb, var(--rosa) 10%, var(--porcelana)); border-color: color-mix(in srgb, var(--rosa) 30%, var(--linha)); color: var(--rosa); }
	.green { background: #f0fdf4; border-color: #bbf7d0; color: #166534; }
	.red { background: #fef2f2; border-color: #fecaca; color: #b91c1c; }
</style>
