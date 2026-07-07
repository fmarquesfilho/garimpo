<script>
	/**
	 * Checkbox — caixa de seleção acessível (Bits UI + Tailwind, shadcn-svelte style).
	 * @prop checked — bind:checked (two-way)
	 * @prop label — rótulo à direita (opcional; clicável)
	 * @prop disabled
	 */
	import { Checkbox } from 'bits-ui';
	import { cn } from '$lib/utils';

	let { checked = $bindable(false), label = '', disabled = false, class: className = '', ...rest } = $props();

	const uid = $props.id();
</script>

<div class={cn('flex items-center gap-2', disabled && 'opacity-50', className)}>
	<Checkbox.Root
		id={uid}
		bind:checked
		{disabled}
		class="flex size-4 shrink-0 cursor-pointer items-center justify-center rounded-sm border border-input bg-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed data-[state=checked]:border-primary data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground"
		{...rest}
	>
		{#snippet children({ checked: isChecked })}
			{#if isChecked}<span class="text-[0.7rem] leading-none">✓</span>{/if}
		{/snippet}
	</Checkbox.Root>
	{#if label}
		<label for={uid} class="cursor-pointer text-sm text-foreground select-none">{label}</label>
	{/if}
</div>
