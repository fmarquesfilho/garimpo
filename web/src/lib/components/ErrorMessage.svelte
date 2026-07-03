<script>
	/**
	 * Componente reutilizável de mensagem de erro.
	 * Suporta RFC 9457 (retry, status) e ação de retry.
	 */
	import { Card, Button } from '$lib/components/ui';

	let { erro, onretry = null } = $props();
</script>

{#if erro}
	<Card variant="error" padding="md">
		<div class="erro-content">
			<p class="erro-titulo">😕 {erro.title ?? 'Algo deu errado.'}</p>
			<p class="erro-msg">{erro.message ?? erro}</p>
			{#if erro.status === 502 || erro.retry}
				<p class="erro-dica">A API pode estar temporariamente fora. Tente novamente em alguns segundos.</p>
			{:else if erro.status === 401}
				<p class="erro-dica">Sua sessão pode ter expirado. Tente fazer logout e login novamente.</p>
			{/if}
			{#if onretry && (erro.retry || erro.status === 502)}
				<Button variant="primary" size="sm" onclick={onretry}>🔄 Tentar novamente</Button>
			{/if}
		</div>
	</Card>
{/if}

<style>
	.erro-content {
		text-align: center;
	}
	.erro-content p {
		margin: var(--r2) 0;
	}
	.erro-titulo {
		font-weight: var(--font-semi);
		font-size: var(--text-base);
	}
	.erro-msg {
		font-size: var(--text-sm);
	}
	.erro-dica {
		color: var(--tinta-suave);
		font-size: var(--text-sm);
	}
</style>
