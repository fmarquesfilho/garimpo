<script>
	/**
	 * TagInput: input com pílulas (tags). Reutilizável para keywords, shop IDs, etc.
	 * @prop tags — array de strings (bind:tags para two-way binding)
	 * @prop placeholder — texto do input
	 * @prop label — label acima do input
	 * @prop variant — 'default' | 'shop' (muda a cor da pílula)
	 * @prop parse — função opcional para processar o valor antes de adicionar
	 */
	import { Badge, Input, Button } from '$lib/components/ui';

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

	let badgeVariant = $derived(variant === 'shop' ? 'secondary' : 'default');
</script>

<div class="flex flex-col gap-2">
	{#if label}
		<span class="text-sm font-semibold text-muted-foreground">{label}</span>
	{/if}
	<div class="flex items-stretch gap-2">
		<div class="flex-1">
			<Input
				bind:value={valor}
				{placeholder}
				onkeydown={(e) => {
					if (e.key === 'Enter') {
						e.preventDefault();
						adicionar();
					}
				}}
			/>
		</div>
		<Button
			variant="secondary"
			size="md"
			class="h-10 shrink-0 px-4 text-lg"
			onclick={adicionar}
			aria-label="Adicionar tag"
		>
			+
		</Button>
	</div>
	{#if tags.length > 0}
		<div class="flex flex-wrap gap-2">
			{#each tags as tag (tag)}
				<Badge variant={badgeVariant}>
					{#if variant === 'shop'}🏪{/if}
					{tag}
					<button
						type="button"
						class="cursor-pointer border-none bg-transparent px-1 py-0.5 text-xs text-inherit opacity-70 hover:opacity-100"
						onclick={() => remover(tag)}
						aria-label="Remover {tag}">✕</button
					>
				</Badge>
			{/each}
		</div>
	{/if}
</div>
