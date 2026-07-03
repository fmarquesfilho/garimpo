<script>
	/**
	 * RankList — lista ranqueada compacta (top N items).
	 * @prop items — array de { nome, valor }
	 * @prop vazio — mensagem quando lista está vazia
	 */
	let { items = [], vazio = 'Sem dados.', ...rest } = $props();
</script>

{#if items.length === 0}
	<p class="m-0 text-sm italic text-tinta-suave">{vazio}</p>
{:else}
	<div class="flex flex-col gap-0.5" {...rest}>
		{#each items as item, i}
			<div class="flex items-center gap-2 border-b border-border py-2 last:border-b-0">
				<span class="w-[18px] text-xs font-bold text-tinta-suave">{i + 1}</span>
				<span class="flex-1 truncate text-sm">{item.nome}</span>
				<span class="font-mono text-xs font-bold tabular-nums text-ouro">{item.valor}</span>
			</div>
		{/each}
	</div>
{/if}
