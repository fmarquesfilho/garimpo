<script>
	/**
	 * StatCard — card compacto para métricas/KPIs.
	 * @prop label — rótulo da métrica
	 * @prop valor — valor formatado como string
	 * @prop variant — 'default' | 'positive' | 'negative' | 'gold'
	 */
	const VARIANTS = ['default', 'positive', 'negative', 'gold'];

	let { label, valor, variant = 'default', ...rest } = $props();

	let resolvedVariant = $derived(VARIANTS.includes(variant) ? variant : 'default');
</script>

<div class="stat-card" {...rest}>
	<span class="label">{label}</span>
	<span class="valor {resolvedVariant}">{valor}</span>
</div>

<style>
	.stat-card {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		padding: var(--r3) var(--r4);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--r1);
	}
	.label {
		font-size: var(--text-xs);
		text-transform: uppercase;
		color: var(--tinta-suave);
		font-weight: var(--font-semi);
		letter-spacing: 0.02em;
	}
	.valor {
		font-size: var(--text-xl);
		font-weight: var(--font-bold);
		font-family: var(--mono);
		font-variant-numeric: tabular-nums;
	}
	.positive {
		color: var(--sucesso-texto);
	}
	.negative {
		color: var(--erro-texto);
	}
	.gold {
		color: var(--ouro);
	}
</style>
