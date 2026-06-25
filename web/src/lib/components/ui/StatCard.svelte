<script>
	/**
	 * StatCard — card compacto para métricas/KPIs.
	 * @prop label — rótulo da métrica
	 * @prop valor — valor formatado como string
	 * @prop variant — 'default' | 'positive' | 'negative' | 'gold'
	 */
	let { label, valor, variant = 'default' } = $props();
</script>

<div class="stat-card">
	<span class="label">{label}</span>
	<span class="valor {variant}">{valor}</span>
</div>

<style>
	.stat-card {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: 10px;
		padding: var(--r3) var(--r4);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
	}
	.label {
		font-size: 0.72rem;
		text-transform: uppercase;
		color: var(--tinta-suave);
		font-weight: 600;
		letter-spacing: 0.02em;
	}
	.valor {
		font-size: 1.2rem;
		font-weight: 700;
	}
	.positive { color: var(--erro-texto); }
	.negative { color: var(--sucesso-texto); }
	.gold { color: var(--ouro); }
</style>
