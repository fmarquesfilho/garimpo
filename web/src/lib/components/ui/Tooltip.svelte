<script>
	/**
	 * Tooltip — dica flutuante acessível usando Bits UI.
	 * @prop content — texto do tooltip
	 * @prop side — 'top' | 'bottom' | 'left' | 'right'
	 * @prop children — elemento que recebe o tooltip
	 */
	import { Tooltip } from 'bits-ui';

	let {
		content = '',
		side = 'top',
		children,
		...rest
	} = $props();
</script>

<Tooltip.Root {...rest}>
	<Tooltip.Trigger asChild>
		{@render children()}
	</Tooltip.Trigger>
	<Tooltip.Portal>
		<Tooltip.Content {side} class="tooltip-content">
			{content}
		</Tooltip.Content>
	</Tooltip.Portal>
</Tooltip.Root>

<style>
	:global(.tooltip-content) {
		font-family: var(--ui);
		font-size: var(--text-xs);
		background: var(--tinta);
		color: var(--branco);
		padding: var(--r1) var(--r2);
		border-radius: var(--raio-sm);
		max-width: 240px;
		z-index: 110;
		animation: tooltipFade 0.15s ease;
		box-shadow: var(--sombra);
	}

	@keyframes tooltipFade {
		from { opacity: 0; transform: scale(0.96); }
		to { opacity: 1; transform: scale(1); }
	}

	@media (prefers-reduced-motion: reduce) {
		:global(.tooltip-content) { animation-duration: 0ms; }
	}
</style>
