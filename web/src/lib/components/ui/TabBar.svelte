<script>
	/**
	 * TabBar — barra de abas horizontal com estilo padrão.
	 * @prop tabs — array de { id, label, badge? }
	 * @prop active — id da aba ativa (bindable)
	 */
	let { tabs = [], active = $bindable('') } = $props();
</script>

<nav class="mb-5 flex gap-0.5 overflow-x-auto border-b-2 border-border">
	{#each tabs as tab (tab.id)}
		<button
			class="-mb-[2px] flex cursor-pointer items-center gap-1.5 whitespace-nowrap border-b-2 border-b-transparent bg-transparent px-4 py-2 text-sm font-semibold text-muted-foreground max-sm:px-3 max-sm:text-xs"
			class:!text-foreground={active === tab.id}
			class:!border-b-primary={active === tab.id}
			onclick={() => (active = tab.id)}
		>
			{tab.label}
			{#if tab.badge}
				<span
					class="rounded-full bg-accent px-2 py-px text-xs font-bold text-accent-foreground"
					class:!bg-destructive={tab.badgeVariant === 'alert'}
					class:!text-destructive-foreground={tab.badgeVariant === 'alert'}
				>
					{tab.badge}
				</span>
			{/if}
		</button>
	{/each}
</nav>
