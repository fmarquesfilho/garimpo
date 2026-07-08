<script>
	/**
	 * Collapsible — seção expansível acessível (Bits UI + Tailwind).
	 * ARIA expanded, animação de abertura, keyboard accessible.
	 *
	 * @prop open — bind:open (boolean)
	 * @prop title — texto do trigger
	 * @prop class — classes extras no container
	 */
	import { Collapsible } from 'bits-ui';
	import { cn } from '$lib/utils';

	let { open = $bindable(false), title = '', class: className = '', children } = $props();
</script>

<Collapsible.Root bind:open class={cn('border-t border-border pt-6', className)}>
	<Collapsible.Trigger
		class="flex cursor-pointer items-center gap-2 text-sm font-semibold text-muted-foreground hover:text-foreground"
	>
		<span class="inline-block transition-transform duration-150" class:rotate-90={open}>▶</span>
		{title}
	</Collapsible.Trigger>

	<Collapsible.Content class="mt-4 space-y-6">
		{@render children()}
	</Collapsible.Content>
</Collapsible.Root>
