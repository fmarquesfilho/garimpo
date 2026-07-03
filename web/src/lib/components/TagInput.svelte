<script>
	/**
	 * TagInput: input com pílulas (tags). Reutilizável para keywords, shop IDs, etc.
	 * @prop tags — array de strings (bind:tags para two-way binding)
	 * @prop placeholder — texto do input
	 * @prop label — label acima do input
	 * @prop variant — 'default' | 'shop' (muda a cor da pílula)
	 * @prop parse — função opcional para processar o valor antes de adicionar
	 */
	import { Badge } from '$lib/components/ui';

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

	let badgeVariant = $derived(variant === 'shop' ? 'pink' : 'gold');
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
			onkeydown={(e) => {
				if (e.key === 'Enter') {
					e.preventDefault();
					adicionar();
				}
			}}
		/>
		<button type="button" class="tag-add" onclick={adicionar} aria-label="Adicionar tag">+</button>
	</div>
	{#if tags.length > 0}
		<div class="tag-list">
			{#each tags as tag (tag)}
				<Badge variant={badgeVariant}>
					{#if variant === 'shop'}🏪{/if}
					{tag}
					<button type="button" class="tag-x" onclick={() => remover(tag)} aria-label="Remover {tag}">✕</button>
				</Badge>
			{/each}
		</div>
	{/if}
</div>

<style>
	.tag-input {
		display: flex;
		flex-direction: column;
		gap: var(--r2);
	}
	.tag-label {
		font-size: var(--text-sm);
		font-weight: var(--font-semi);
		color: var(--tinta-suave);
	}
	.tag-field {
		display: flex;
		gap: var(--r2);
	}
	.tag-entrada {
		font-family: var(--ui);
		font-size: var(--text-base);
		padding: var(--r3) var(--r4);
		border-radius: var(--raio-sm);
		border: 1px solid var(--linha);
		background: var(--porcelana);
		color: var(--tinta);
		width: 100%;
		flex: 1;
		transition: border-color 0.15s ease;
	}
	.tag-entrada::placeholder {
		color: var(--tinta-suave);
		opacity: 0.7;
	}
	.tag-entrada:focus {
		outline: none;
		border-color: var(--ouro);
		box-shadow: 0 0 0 2px var(--ouro-fundo);
	}
	.tag-add {
		border: 1px solid var(--linha);
		background: var(--ouro-fundo);
		color: var(--ouro-escuro);
		font-weight: var(--font-bold);
		font-size: var(--text-lg);
		padding: 0 var(--r4);
		border-radius: var(--raio-sm);
		cursor: pointer;
		flex-shrink: 0;
		transition: border-color 0.15s ease;
	}
	.tag-add:hover {
		border-color: var(--ouro);
	}
	.tag-list {
		display: flex;
		flex-wrap: wrap;
		gap: var(--r2);
	}
	.tag-x {
		border: none;
		background: transparent;
		color: inherit;
		font-size: var(--text-xs);
		cursor: pointer;
		padding: 2px var(--r1);
		opacity: 0.7;
	}
	.tag-x:hover {
		opacity: 1;
	}

	@media (prefers-reduced-motion: reduce) {
		.tag-entrada,
		.tag-add {
			transition-duration: 0ms;
		}
	}
</style>
