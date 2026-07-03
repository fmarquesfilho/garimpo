<script>
	/**
	 * DropdownMenu — menu suspenso acessível usando Bits UI.
	 * @prop items — array de { label, onclick, disabled?, destructive? }
	 * @prop children — trigger content (botão que abre o menu)
	 */
	import { DropdownMenu } from 'bits-ui';

	let {
		items = [],
		children,
		...rest
	} = $props();
</script>

<DropdownMenu.Root {...rest}>
	<DropdownMenu.Trigger asChild>
		{@render children()}
	</DropdownMenu.Trigger>
	<DropdownMenu.Portal>
		<DropdownMenu.Content class="dropdown-content" sideOffset={4}>
			{#each items as item, i (i)}
				<DropdownMenu.Item
					class="dropdown-item{item.destructive ? ' destructive' : ''}"
					disabled={item.disabled}
					onSelect={item.onclick}
				>
					{item.label}
				</DropdownMenu.Item>
			{/each}
		</DropdownMenu.Content>
	</DropdownMenu.Portal>
</DropdownMenu.Root>

<style>
	:global(.dropdown-content) {
		background: var(--branco);
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		box-shadow: var(--sombra);
		padding: var(--r1) 0;
		min-width: 160px;
		z-index: 50;
		animation: dropdownSlide 0.15s ease;
	}

	:global(.dropdown-item) {
		font-family: var(--ui);
		font-size: var(--text-base);
		padding: var(--r2) var(--r4);
		cursor: pointer;
		outline: none;
		display: flex;
		align-items: center;
		gap: var(--r2);
		color: var(--tinta);
		transition: background 0.1s ease;
	}
	:global(.dropdown-item[data-highlighted]) {
		background: var(--porcelana);
	}
	:global(.dropdown-item[data-disabled]) {
		opacity: 0.5;
		cursor: not-allowed;
	}
	:global(.dropdown-item.destructive) {
		color: var(--erro-texto);
	}
	:global(.dropdown-item.destructive[data-highlighted]) {
		background: var(--erro-fundo);
	}

	@keyframes dropdownSlide {
		from { opacity: 0; transform: translateY(-4px); }
		to { opacity: 1; transform: translateY(0); }
	}

	@media (prefers-reduced-motion: reduce) {
		:global(.dropdown-content) { animation-duration: 0ms; }
		:global(.dropdown-item) { transition-duration: 0ms; }
	}
</style>
