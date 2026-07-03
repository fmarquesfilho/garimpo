<script>
	/**
	 * Input — text input with label and error state.
	 * @prop label — optional label above
	 * @prop error — error message (renders below, colors border)
	 * @prop class — extra Tailwind classes on the input element
	 */
	import { cn } from '$lib/utils';

	let { label = '', error = '', class: className = '', value = $bindable(''), ...rest } = $props();
</script>

<div class="flex flex-col gap-1">
	{#if label}
		<!-- svelte-ignore a11y_label_has_associated_control -->
		<label class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
			{label}
		</label>
	{/if}
	<input
		bind:value
		class={cn(
			'flex h-9 w-full rounded-sm border border-input bg-background px-3 py-2 text-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
			error && 'border-erro focus-visible:ring-erro',
			className
		)}
		{...rest}
	/>
	{#if error}
		<span class="text-xs text-erro">{error}</span>
	{/if}
</div>
