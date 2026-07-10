<script>
	/**
	 * MarketplaceFilter — Ícones toggle para filtrar por marketplace (0..N).
	 *
	 * Cada marketplace suportado é exibido como um botão-ícone que o usuário
	 * ativa ou desativa. Zero selecionados = sem filtro (todos os marketplaces).
	 *
	 * Lê a config de `rules.marketplaces` via busca-config.js.
	 *
	 * @prop marketplaces — lista dos marketplaces atualmente selecionados
	 * @prop onchange — (marketplaces: string[]) => void
	 */
	import { MARKETPLACES } from '$lib/busca-config.js';
	import { cn } from '$lib/utils';

	let { marketplaces = [], onchange = null } = $props();

	const labels = {
		shopee: 'Shopee',
		mercado_livre: 'Mercado Livre',
		amazon: 'Amazon'
	};

	function toggle(mkt) {
		const ativos = marketplaces.includes(mkt)
			? marketplaces.filter((m) => m !== mkt)
			: [...marketplaces, mkt];
		onchange?.(ativos);
	}
</script>

<div class="flex items-center gap-1.5">
	<span class="font-[var(--mono)] text-[0.6rem] uppercase tracking-wider text-muted-foreground">marketplaces</span>
	<div class="flex items-center gap-1">
		{#each MARKETPLACES.suportados as mkt (mkt)}
			{@const ativo = marketplaces.includes(mkt)}
			{@const icone = MARKETPLACES.icones?.[mkt] ?? '🔘'}
			<button
				type="button"
				class={cn(
					'flex items-center gap-1 rounded-full border px-2.5 py-1 text-sm transition-colors',
					ativo
						? 'border-primary bg-[var(--ouro-fundo)] font-semibold text-[var(--ouro-escuro)]'
						: 'border-border bg-card text-muted-foreground hover:border-primary'
				)}
				title={labels[mkt] ?? mkt}
				onclick={() => toggle(mkt)}
			>
				<span class="text-base leading-none">{icone}</span>
				<span class="hidden text-xs sm:inline">{labels[mkt] ?? mkt}</span>
			</button>
		{/each}
	</div>
	{#if marketplaces.length === 0}
		<span class="text-xs italic text-muted-foreground">todos</span>
	{/if}
</div>
