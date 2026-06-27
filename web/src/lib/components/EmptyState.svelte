<script>
	/**
	 * Estado vazio genérico com ícone, mensagem e dica opcional.
	 */
	let { icone = '📭', mensagem = '', dica = '' } = $props();
</script>

<div class="vazio">
	{#if icone}<span class="icone">{icone}</span>{/if}
	<p class="mensagem">{@html mensagem}</p>
	{#if dica}<p class="dica">{@html dica}</p>{/if}
</div>

<style>
	.vazio {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r8);
		text-align: center;
	}
	.icone { font-size: 2rem; display: block; margin-bottom: var(--r3); }
	.mensagem { font-size: 0.95rem; margin: 0; }
	.dica { color: var(--tinta-suave); font-size: 0.85rem; margin-top: var(--r2); }
</style>
