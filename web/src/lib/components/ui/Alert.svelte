<script>
	/**
	 * Alert — mensagem de feedback (erro, sucesso, aviso).
	 * @prop variant — 'error' | 'success' | 'warning'
	 * @prop inline — se true, estilo mais compacto sem fundo (apenas cor)
	 */
	const VARIANTS = ['error', 'success', 'warning'];

	let { variant = 'error', inline = false, children, ...rest } = $props();

	let resolvedVariant = $derived(VARIANTS.includes(variant) ? variant : 'error');
</script>

{#if inline}
	<p class="inline {resolvedVariant}" role="alert" {...rest}>
		{@render children()}
	</p>
{:else}
	<div class="alert {resolvedVariant}" role="alert" {...rest}>
		{@render children()}
	</div>
{/if}

<style>
	.alert {
		padding: var(--r3) var(--r4);
		border-radius: var(--raio-sm);
		font-size: var(--text-base);
		border: 1px solid;
	}
	.alert.error {
		background: var(--erro-fundo);
		color: var(--erro-texto);
		border-color: var(--erro-borda);
	}
	.alert.success {
		background: var(--sucesso-fundo);
		color: var(--sucesso-texto);
		border-color: var(--sucesso-borda);
	}
	.alert.warning {
		background: var(--aviso-fundo);
		color: var(--aviso-texto);
		border-color: var(--aviso-borda);
	}

	.inline {
		font-size: var(--text-sm);
		margin-top: var(--r2);
	}
	.inline.error { color: var(--erro-texto); }
	.inline.success { color: var(--sucesso-texto); }
	.inline.warning { color: var(--aviso-texto); }
</style>
