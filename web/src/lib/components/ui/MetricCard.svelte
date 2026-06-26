<script>
	/**
	 * MetricCard — card de métrica numérica para dashboards.
	 * @prop valor — o número/texto principal
	 * @prop label — descrição abaixo do valor
	 * @prop variant — 'default' | 'gold' | 'green' | 'red'
	 */
	let { valor, label, variant = 'default' } = $props();
</script>

<div class="metrica {variant}">
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
		font-size: 1.5rem;
		font-weight: 700;
		font-family: var(--mono);
		line-height: 1.2;
	}
	.metrica-label {
		font-size: var(--text-xs);
		color: var(--tinta-suave);
		text-transform: uppercase;
		font-weight: 600;
	}
	.gold .metrica-valor { color: var(--ouro); }
	.green .metrica-valor { color: var(--sucesso-texto); }
	.red .metrica-valor { color: var(--erro-texto); }
</style>
