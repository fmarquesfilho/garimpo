<script>
	/**
	 * TagInput: input com pílulas (tags). Reutilizável para keywords, shop IDs, etc.
	 * @prop tags — array de strings (bind:tags para two-way binding)
	 * @prop placeholder — texto do input
	 * @prop label — label acima do input
	 * @prop variant — 'default' | 'shop' (muda a cor da pílula)
	 * @prop parse — função opcional para processar o valor antes de adicionar
	 */
	let { tags = $bindable([]), placeholder = '', label = '', variant = 'default', parse = null } = $props();

	let valor = $state('');

	function adicionar() {
		let v = valor.trim();
		if (!v) return;
		if (parse) v = parse(v);
		if (!v || tags.includes(v)) {
			valor = '';
			return;
		}
		tags = [...tags, v];
		valor = '';
	}

	function remover(tag) {
		tags = tags.filter((t) => t !== tag);
	}
</script>

<div class="tag-input">
	{#if label}
		<span class="tag-label">{label}</span>
	{/if}
	<div class="tag-field">
		<input
			class="tag-entrada"
			bind:value={valor}
			{placeholder}
			onkeydown={(e) => e.key === 'Enter' && adicionar()}
		/>
		<button type="button" class="tag-add" onclick={adicionar}>+</button>
	</div>
	{#if tags.length > 0}
		<div class="tag-list">
			{#each tags as tag}
				<span class="tag-pill" class:shop={variant === 'shop'}>
					{#if variant === 'shop'}🏪{/if} {tag}
					<button type="button" class="tag-x" onclick={() => remover(tag)}>✕</button>
				</span>
			{/each}
		</div>
	{/if}
</div>

<style>
	.tag-input { display: flex; flex-direction: column; gap: 6px; }
	.tag-label {
		font-size: 0.8rem; font-weight: 600; color: var(--tinta-suave);
	}
	.tag-field { display: flex; gap: var(--r2); }
	.tag-entrada {
		font-family: var(--ui); font-size: 0.95rem; padding: 9px 12px;
		border-radius: 10px; border: 1px solid var(--linha);
		background: var(--porcelana); color: var(--tinta); width: 100%; flex: 1;
	}
	.tag-entrada::placeholder { color: var(--tinta-suave); opacity: 0.7; }
	.tag-add {
		border: 1px solid var(--linha); background: var(--ouro-fundo);
		color: #7a5a1e; font-weight: 700; font-size: 1rem;
		padding: 0 14px; border-radius: 10px; cursor: pointer; flex-shrink: 0;
	}
	.tag-list { display: flex; flex-wrap: wrap; gap: var(--r2); }
	.tag-pill {
		display: inline-flex; align-items: center; gap: 5px;
		background: var(--ouro-fundo); border: 1px solid color-mix(in srgb, var(--ouro) 40%, var(--linha));
		border-radius: 999px; padding: 3px 6px 3px 10px;
		font-size: 0.85rem; font-weight: 600; color: #7a5a1e;
	}
	.tag-pill.shop {
		background: color-mix(in srgb, var(--rosa) 10%, var(--porcelana));
		border-color: color-mix(in srgb, var(--rosa) 30%, var(--linha));
		color: var(--rosa);
	}
	.tag-x {
		border: none; background: transparent; color: inherit;
		font-size: 0.72rem; cursor: pointer; padding: 2px 4px;
		opacity: 0.7;
	}
	.tag-x:hover { opacity: 1; }
</style>
