<script>
	/**
	 * RankList — lista ranqueada compacta (top N items).
	 * @prop items — array de { nome, valor }
	 * @prop vazio — mensagem quando lista está vazia
	 */
	let { items = [], vazio = 'Sem dados.', ...rest } = $props();
</script>

{#if items.length === 0}
	<p class="vazio">{vazio}</p>
{:else}
	<div class="lista" {...rest}>
		{#each items as item, i}
			<div class="item">
				<span class="pos">{i + 1}</span>
				<span class="nome">{item.nome}</span>
				<span class="valor">{item.valor}</span>
			</div>
		{/each}
	</div>
{/if}

<style>
	.vazio {
		font-size: var(--text-sm);
		color: var(--tinta-suave);
		font-style: italic;
		margin: 0;
	}
	.lista {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.item {
		display: flex;
		align-items: center;
		gap: var(--r2);
		padding: var(--r2) 0;
		border-bottom: 1px solid var(--linha);
	}
	.item:last-child {
		border-bottom: none;
	}
	.pos {
		font-size: var(--text-xs);
		font-weight: var(--font-bold);
		color: var(--tinta-suave);
		width: 18px;
	}
	.nome {
		flex: 1;
		font-size: var(--text-sm);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.valor {
		font-size: var(--text-xs);
		font-weight: var(--font-bold);
		color: var(--ouro);
		font-family: var(--mono);
		font-variant-numeric: tabular-nums;
	}
</style>
