<script>
	/**
	 * Tabs — acessível com Bits UI + Tailwind.
	 * @prop tabs — array de { id, label, badge?, badgeVariant? }
	 * @prop active — bind:active para two-way (id da aba ativa)
	 * @prop children — conteúdo renderizado abaixo
	 */
	import { Tabs } from 'bits-ui';
	import { cn } from '$lib/utils';
	import Badge from './Badge.svelte';

	let { tabs = [], active = $bindable(''), children, class: className = '', ...rest } = $props();

	$effect(() => {
		if (!active && tabs.length > 0) active = tabs[0].id;
	});
</script>

<Tabs.Root bind:value={active} class={cn('w-full', className)} {...rest}>
	<Tabs.List class="inline-flex h-9 items-center gap-1 rounded-md bg-muted p-1 text-muted-foreground">
		{#each tabs as tab (tab.id)}
			<Tabs.Trigger
				value={tab.id}
				class="inline-flex items-center justify-center gap-1.5 whitespace-nowrap rounded-sm px-3 py-1 text-sm font-medium transition-colors hover:text-foreground data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm"
			>
				{tab.label}
				{#if tab.badge}
					<Badge variant={tab.badgeVariant === 'alert' ? 'error' : 'secondary'} class="text-[0.6rem] px-1.5 py-0">
						{tab.badge}
					</Badge>
				{/if}
			</Tabs.Trigger>
		{/each}
	</Tabs.List>
	<div class="mt-4">
		{@render children()}
	</div>
</Tabs.Root>
