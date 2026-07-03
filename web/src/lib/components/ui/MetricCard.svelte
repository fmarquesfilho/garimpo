<script>
	/**
	 * MetricCard — card de métrica numérica para dashboards.
	 * @prop valor — o número/texto principal
	 * @prop label — descrição abaixo do valor
	 * @prop variant — 'default' | 'gold' | 'green' | 'red'
	 */
	const VARIANTS = ['default', 'gold', 'green', 'red'];

	let { valor, label, variant = 'default', ...rest } = $props();

	let resolvedVariant = $derived(VARIANTS.includes(variant) ? variant : 'default');
</script>

<div class="metrica {resolvedVariant}" {...rest}>
	<span class="metrica-valor">{valor}</span>
	<span class="metrica-label">{label}</span>
</div>

<style>
	.metrica {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		padding: var(--r3) var(--r4);
		text-align: center;
	}
	.metrica-valor {
		display: block;
		font-size: var(--text-2xl);
		font-weight: var(--font-bold);
		font-family: var(--mono);
		font-variant-numeric: tabular-nums;
		line-height: 1.2;
	}
	.metrica-label {
		font-size: var(--text-xs);
		color: var(--tinta-suave);
		text-transform: uppercase;
		font-weight: var(--font-semi);
	}
	.gold .metrica-valor { color: var(--ouro); }
	.green .metrica-valor { color: var(--sucesso-texto); }
	.red .metrica-valor { color: var(--erro-texto); }
</style>
