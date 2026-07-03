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

<div class="flex flex-col items-center gap-1 rounded-sm border border-border bg-card px-4 py-3" {...rest}>
	<span class="text-xs font-semibold uppercase tracking-wide text-tinta-suave">{label}</span>
	<span
		class="font-mono text-xl font-bold tabular-nums"
		class:text-[var(--sucesso-texto)]={resolvedVariant === 'positive'}
		class:text-[var(--erro-texto)]={resolvedVariant === 'negative'}
		class:text-ouro={resolvedVariant === 'gold'}>{valor}</span
	>
</div>
