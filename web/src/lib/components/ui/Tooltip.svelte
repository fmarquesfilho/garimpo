<script>
	/**
	 * Tooltip — dica flutuante acessível com Bits UI + Tailwind.
	 * @prop content — texto do tooltip
	 * @prop side — 'top' | 'bottom' | 'left' | 'right'
	 * @prop children — elemento que recebe o tooltip
	 */
	import { Tooltip } from 'bits-ui';

	let { content = '', side = /** @type {'top'|'bottom'|'left'|'right'} */ ('top'), children, ...rest } = $props();
</script>

<Tooltip.Root {...rest}>
	<Tooltip.Trigger>
		{@render children()}
	</Tooltip.Trigger>
	<Tooltip.Portal>
		<Tooltip.Content
			{side}
			class="z-50 max-w-[240px] rounded-sm bg-foreground px-2 py-1 text-xs text-background shadow-md animate-in fade-in-0 zoom-in-95"
			sideOffset={4}
		>
			{content}
		</Tooltip.Content>
	</Tooltip.Portal>
</Tooltip.Root>
