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

<div class="rounded-sm border border-border bg-card px-4 py-3 text-center" {...rest}>
	<span
		class="block font-mono text-2xl font-bold leading-tight tabular-nums"
		class:text-ouro={resolvedVariant === 'gold'}
		class:text-[var(--sucesso-texto)]={resolvedVariant === 'green'}
		class:text-[var(--erro-texto)]={resolvedVariant === 'red'}>{valor}</span
	>
	<span class="text-xs font-semibold uppercase text-tinta-suave">{label}</span>
</div>
