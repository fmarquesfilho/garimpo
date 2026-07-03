<script>
	/**
	 * Badge — etiqueta/pílula de informação.
	 * @prop variant — 'default' | 'gold' | 'pink' | 'green' | 'red'
	 */
	const VARIANTS = ['default', 'gold', 'pink', 'green', 'red'];

	let { variant = 'default', children, ...rest } = $props();

	let resolvedVariant = $derived(VARIANTS.includes(variant) ? variant : 'default');
</script>

<span class="badge {resolvedVariant}" {...rest}>
	{@render children()}
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		gap: var(--r1);
		font-size: var(--text-xs);
		font-weight: var(--font-bold);
		padding: 2px var(--r2);
		border-radius: var(--raio-full);
		white-space: nowrap;
	}

	.default {
		background: var(--porcelana);
		color: var(--tinta-suave);
		border: 1px solid var(--linha);
	}
	.gold {
		background: var(--ouro-fundo);
		border: 1px solid var(--ouro-claro);
		color: var(--ouro-escuro);
	}
	.pink {
		background: var(--rosa-fundo);
		border: 1px solid color-mix(in srgb, var(--rosa) 30%, var(--linha));
		color: var(--rosa);
	}
	.green {
		background: var(--sucesso-fundo);
		border: 1px solid var(--sucesso-borda);
		color: var(--sucesso-texto);
	}
	.red {
		background: var(--erro-fundo);
		border: 1px solid var(--erro-borda);
		color: var(--erro-texto);
	}
</style>
