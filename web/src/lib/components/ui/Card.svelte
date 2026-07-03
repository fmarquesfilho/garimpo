<script>
	/**
	 * Card — container com borda e sombra padrão.
	 * @prop variant — 'default' | 'highlight' | 'success' | 'error'
	 * @prop padding — 'sm' | 'md' | 'lg'
	 */
	const VARIANTS = ['default', 'highlight', 'success', 'error'];
	const PADDINGS = ['sm', 'md', 'lg'];

	let { variant = 'default', padding = 'md', children, ...rest } = $props();

	let resolvedVariant = $derived(VARIANTS.includes(variant) ? variant : 'default');
	let resolvedPadding = $derived(PADDINGS.includes(padding) ? padding : 'md');
</script>

<div class="card {resolvedVariant} pad-{resolvedPadding}" {...rest}>
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

	.highlight {
		border-color: var(--ouro-claro);
		background: linear-gradient(180deg, var(--ouro-fundo), var(--nevoa));
	}
	.success {
		border-left: 3px solid var(--sucesso-texto);
		background: var(--sucesso-fundo);
	}
	.error {
		border-left: 3px solid var(--erro-texto);
		background: var(--erro-fundo);
	}
</style>
