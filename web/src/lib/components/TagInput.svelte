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

	let badgeVariant = $derived(variant === 'shop' ? 'secondary' : 'default');
</script>

<div class="flex flex-col gap-2">
	{#if label}
		<span class="text-sm font-semibold text-muted-foreground">{label}</span>
	{/if}
	<div class="flex gap-2">
		<input
			class="flex-1 rounded-sm border border-border bg-background px-4 py-3 font-[var(--ui)] text-foreground transition-[border-color] duration-150 ease-linear placeholder:text-muted-foreground/70 focus:border-primary focus:ring-2 focus:ring-ring/20 focus:outline-none motion-reduce:transition-none"
			bind:value={valor}
			{placeholder}
			onkeydown={(e) => {
				if (e.key === 'Enter') {
					e.preventDefault();
					adicionar();
				}
			}}
		/>
		<button
			type="button"
			class="shrink-0 rounded-sm border border-border bg-accent px-4 text-lg font-bold text-accent-foreground transition-[border-color] duration-150 ease-linear hover:border-primary motion-reduce:transition-none"
			onclick={adicionar}
			aria-label="Adicionar tag">+</button
		>
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
