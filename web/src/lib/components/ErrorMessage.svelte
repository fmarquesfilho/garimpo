<script>
	/**
	 * Componente reutilizável de mensagem de erro.
	 * Suporta RFC 9457 (retry, status) e ação de retry.
	 */
	let { erro, onretry = null } = $props();
</script>

{#if erro}
	<div class="msg-erro">
		<p><strong>😕 {erro.title ?? 'Algo deu errado.'}</strong></p>
		<p>{erro.message ?? erro}</p>
		{#if erro.status === 502 || erro.retry}
			<p class="dica">A API pode estar temporariamente fora. Tente novamente em alguns segundos.</p>
		{:else if erro.status === 401}
			<p class="dica">Sua sessão pode ter expirado. Tente fazer logout e login novamente.</p>
		{/if}
		{#if onretry && (erro.retry || erro.status === 502)}
			<button class="btn-retry" onclick={onretry}>🔄 Tentar novamente</button>
		{/if}
	</div>
{/if}

<style>
	.msg-erro {
		background: var(--nevoa);
		border: 1px solid color-mix(in srgb, var(--erro-texto) 30%, var(--linha));
		border-radius: var(--raio);
		padding: var(--r5);
		text-align: center;
	}
	.msg-erro p { margin: var(--r2) 0; }
	.dica { color: var(--tinta-suave); font-size: 0.85rem; }
	.btn-retry {
		margin-top: var(--r3);
		padding: 8px 16px;
		background: var(--ouro);
		color: white;
		border: none;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.85rem;
		cursor: pointer;
	}
	.btn-retry:hover { opacity: 0.9; }
</style>
