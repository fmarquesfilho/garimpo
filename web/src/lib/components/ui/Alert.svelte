<script>
	/**
	 * Alert — mensagem de feedback (erro, sucesso, aviso).
	 * @prop variant — 'error' | 'success' | 'warning'
	 * @prop inline — se true, estilo mais compacto sem fundo (apenas cor)
	 */
	let { variant = 'error', inline = false, children } = $props();
</script>

{#if inline}
	<p class="inline {variant}">
		{@render children()}
	</p>
{:else}
	<div class="alert {variant}">
		{@render children()}
	</div>
{/if}

<style>
	.alert {
		padding: var(--r3) var(--r4);
		border-radius: 10px;
		font-size: 0.88rem;
		border: 1px solid;
	}
	.alert.error { background: var(--erro-fundo); color: var(--erro-texto); border-color: var(--erro-borda); }
	.alert.success { background: var(--sucesso-fundo); color: var(--sucesso-texto); border-color: var(--sucesso-borda); }
	.alert.warning { background: var(--aviso-fundo); color: var(--aviso-texto); border-color: var(--aviso-borda); }

	.inline { font-size: 0.85rem; margin-top: 6px; }
	.inline.error { color: var(--erro-texto); }
	.inline.success { color: var(--sucesso-texto); }
	.inline.warning { color: var(--aviso-texto); }
</style>
