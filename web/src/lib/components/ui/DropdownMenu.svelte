<script>
	/**
	 * DropdownMenu — menu de contexto acessível com Bits UI + Tailwind.
	 * @prop items — array de { label, onclick, destructive? }
	 * @prop children — trigger element (slot)
	 */
	import { DropdownMenu } from 'bits-ui';

	let { items = [], children, ...rest } = $props();
</script>

<DropdownMenu.Root {...rest}>
	<DropdownMenu.Trigger>
		{@render children()}
	</DropdownMenu.Trigger>
	<DropdownMenu.Portal>
		<DropdownMenu.Content
			class="z-50 min-w-[8rem] overflow-hidden rounded-md border border-border bg-popover p-1 shadow-md animate-in fade-in-0 zoom-in-95"
			sideOffset={4}
		>
			{#each items as item (item.label)}
				<DropdownMenu.Item
					class="relative flex cursor-pointer select-none items-center rounded-sm px-3 py-2 text-sm outline-none transition-colors data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground {item.destructive
						? 'text-destructive data-[highlighted]:text-destructive'
						: ''}"
					onSelect={item.onclick}
				>
					{item.label}
				</DropdownMenu.Item>
			{/each}
		</DropdownMenu.Content>
	</DropdownMenu.Portal>
</DropdownMenu.Root>
