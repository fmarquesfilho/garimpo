<script>
	/**
	 * Estado vazio genérico com ícone, mensagem e dica opcional.
	 * Usa Card do UI como base.
	 */
	import { Card } from '$lib/components/ui';

	let { icone = '📭', mensagem = '', dica = '' } = $props();
</script>

<Card padding="lg">
	<div class="vazio-content">
		{#if icone}<span class="icone">{icone}</span>{/if}
		<p class="mensagem">{@html mensagem}</p>
		{#if dica}<p class="dica">{@html dica}</p>{/if}
	</div>
</Card>

<style>
	.vazio-content {
		text-align: center;
	}
	.icone {
		font-size: 2rem;
		display: block;
		margin-bottom: var(--r3);
	}
	.mensagem {
		font-size: var(--text-md);
		margin: 0;
		color: var(--tinta);
	}
	.dica {
		color: var(--tinta-suave);
		font-size: var(--text-sm);
		margin-top: var(--r2);
	}
	.dica :global(a) {
		color: var(--ouro);
		text-decoration: underline;
	}
</style>
