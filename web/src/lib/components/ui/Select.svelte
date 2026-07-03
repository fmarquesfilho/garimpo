<script>
	/**
	 * Select — dropdown acessível usando Bits UI + Tailwind.
	 * @prop value — bind:value para two-way
	 * @prop label — rótulo acima (opcional)
	 * @prop options — array de { value, label }
	 * @prop placeholder
	 * @prop size — 'sm' | 'md' | 'lg'
	 * @prop disabled
	 */
	import { Select } from 'bits-ui';
	import { cn } from '$lib/utils';

	const SIZES = { sm: 'h-8 text-xs px-2', md: 'h-9 text-sm px-3', lg: 'h-11 text-base px-4' };

	let {
		value = $bindable(''),
		label = '',
		options = [],
		placeholder = '',
		size = 'md',
		disabled = false,
		class: className = '',
		...rest
	} = $props();

	let selectedLabel = $derived(options.find((o) => o.value === value)?.label ?? '');
</script>

<div class={cn('flex flex-col gap-1', className)} {...rest}>
	{#if label}
		<span class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{label}</span>
	{/if}
	<Select.Root type="single" {value} onValueChange={(v) => (value = v)} {disabled}>
		<Select.Trigger
			class={cn(
				'inline-flex w-full items-center justify-between gap-2 rounded-sm border border-input bg-background font-sans transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 data-[state=open]:border-ring data-[state=open]:ring-2 data-[state=open]:ring-ring/20',
				SIZES[size] ?? SIZES.md
			)}
		>
			{#if selectedLabel}
				<span class="truncate">{selectedLabel}</span>
			{:else}
				<span class="text-muted-foreground">{placeholder}</span>
			{/if}
			<span class="text-muted-foreground text-xs">▾</span>
		</Select.Trigger>
		<Select.Portal>
			<Select.Content
				class="z-50 max-h-72 overflow-y-auto rounded-md border border-border bg-popover p-1 shadow-md animate-in fade-in-0 zoom-in-95"
			>
				<Select.Viewport>
					{#each options as opt (opt.value)}
						<Select.Item
							value={opt.value}
							class="relative cursor-pointer select-none rounded-sm px-3 py-2 text-sm outline-none transition-colors data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground data-[state=checked]:font-semibold data-[state=checked]:text-primary"
						>
							{opt.label}
						</Select.Item>
					{/each}
				</Select.Viewport>
			</Select.Content>
		</Select.Portal>
	</Select.Root>
</div>
