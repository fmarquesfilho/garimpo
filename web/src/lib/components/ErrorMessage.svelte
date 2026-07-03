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
		<div class="text-center">
			<p class="my-2 font-semibold">{erro.title ?? 'Algo deu errado.'}</p>
			<p class="my-2 text-sm">{erro.message ?? erro}</p>
			{#if erro.status === 502 || erro.retry}
				<p class="my-2 text-sm text-tinta-suave">A API pode estar temporariamente fora. Tente novamente em alguns segundos.</p>
			{:else if erro.status === 401}
				<p class="my-2 text-sm text-tinta-suave">Sua sessão pode ter expirado. Tente fazer logout e login novamente.</p>
			{/if}
			{#if onretry && (erro.retry || erro.status === 502)}
				<Button variant="primary" size="sm" onclick={onretry}>🔄 Tentar novamente</Button>
			{/if}
		</div>
	</Card>
{/if}
