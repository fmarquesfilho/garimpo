<script>
	/**
	 * StoreCard — card de resultado de busca por loja. Configurável por marketplace
	 * via rules/busca-rules.json (storeCard.camposVisiveis). Emite MONITORAR_LOJA.
	 */
	import { MARKETPLACES } from '$lib/busca-config.js';

	let { loja, engine } = $props();

	const mktIcone = $derived(MARKETPLACES?.icones?.[loja.marketplace] ?? '🛒');
	const monitorando = $derived(engine.ctx.resolucaoLoja.status === 'resolvendo');
	const erro = $derived(engine.ctx.resolucaoLoja.status === 'erro' ? engine.ctx.resolucaoLoja.erro : null);
</script>

<div class="flex items-center gap-3 rounded-md border border-border bg-card p-3">
	{#if loja.imagem}
		<img src={loja.imagem} alt={loja.nome} class="h-10 w-10 rounded-full object-cover" />
	{:else}
		<span class="flex h-10 w-10 items-center justify-center rounded-full bg-muted text-lg">{mktIcone}</span>
	{/if}

	<div class="flex-1 min-w-0">
		<div class="flex items-center gap-1.5">
			<p class="truncate font-medium text-sm">{loja.nome}</p>
			{#if loja.origem}
				<span class="text-xs" title="Origem">{loja.origem}</span>
			{/if}
		</div>
		<p class="text-xs text-muted-foreground">
			{loja.marketplace}
			{#if loja.total_produtos}
				&middot; {loja.total_produtos} produtos
			{/if}
			{#if loja.seguidores}
				&middot; {loja.seguidores.toLocaleString()} seguidores
			{/if}
			{#if loja.avaliacao}
				&middot; ⭐ {loja.avaliacao.toFixed(1)}
			{/if}
		</p>
		{#if erro}
			<p class="text-xs text-destructive mt-0.5">{erro}</p>
		{/if}
	</div>

	{#if loja.monitorada}
		<span class="text-green-600 text-sm" title="Monitorada">⏱ ✓</span>
	{:else}
		<button
			class="rounded-sm bg-primary/10 px-2 py-1 text-xs font-medium text-primary
				   hover:bg-primary/20 disabled:opacity-50"
			disabled={monitorando}
			onclick={() => engine.send({ type: 'MONITORAR_LOJA', loja })}
			aria-label="Monitorar loja {loja.nome}"
		>
			{monitorando ? '…' : '+ Monitorar'}
		</button>
	{/if}
</div>
