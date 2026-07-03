<script>
	/**
	 * Select — dropdown acessível usando Bits UI.
	 * @prop value — bind:value para two-way
	 * @prop label — rótulo acima (opcional)
	 * @prop options — array de { value, label }
	 * @prop placeholder
	 * @prop size — 'sm' | 'md' | 'lg'
	 * @prop disabled
	 */
	import { Select } from 'bits-ui';

	const SIZES = ['sm', 'md', 'lg'];

	let {
		value = $bindable(''),
		label = '',
		options = [],
		placeholder = '',
		size = 'md',
		disabled = false,
		...rest
	} = $props();

	let resolvedSize = $derived(SIZES.includes(size) ? size : 'md');
</script>

<div class="select-wrapper" {...rest}>
	{#if label}
		<span class="select-label">{label}</span>
	{/if}
	<Select.Root type="single" bind:value {disabled}>
		<Select.Trigger class="select-trigger size-{resolvedSize}">
			<Select.Value {placeholder} />
		</Select.Trigger>
		<Select.Portal>
			<Select.Content class="select-content">
				<Select.Viewport>
					{#each options as opt (opt.value)}
						<Select.Item value={opt.value} class="select-item">
							{opt.label}
						</Select.Item>
					{/each}
				</Select.Viewport>
			</Select.Content>
		</Select.Portal>
	</Select.Root>
</div>

<style>
	.select-wrapper {
		display: flex;
		flex-direction: column;
		gap: var(--r1);
	}
	.select-label {
		font-size: var(--text-xs);
		font-weight: var(--font-semi);
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--tinta-suave);
	}

	:global(.select-trigger) {
		display: inline-flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--r2);
		font-family: var(--ui);
		padding: var(--r3) var(--r4);
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		background: var(--branco);
		color: var(--tinta);
		cursor: pointer;
		width: 100%;
		transition: border-color 0.15s ease, box-shadow 0.15s ease;
	}
	:global(.select-trigger:focus-visible) {
		outline: 2px solid var(--ouro);
		outline-offset: 2px;
	}
	:global(.select-trigger[data-state="open"]) {
		border-color: var(--ouro);
		box-shadow: 0 0 0 2px var(--ouro-fundo);
	}
	:global(.select-trigger[data-disabled]) {
		opacity: 0.5;
		cursor: not-allowed;
	}
	:global(.select-trigger.size-sm) {
		font-size: var(--text-sm);
		padding: var(--r2) var(--r3);
	}
	:global(.select-trigger.size-md) {
		font-size: var(--text-base);
		padding: var(--r3) var(--r4);
	}
	:global(.select-trigger.size-lg) {
		font-size: var(--text-md);
		padding: var(--r4) var(--r5);
	}

	:global(.select-content) {
		background: var(--branco);
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		box-shadow: var(--sombra);
		padding: var(--r1) 0;
		max-height: 300px;
		overflow-y: auto;
		z-index: 50;
	}

	:global(.select-item) {
		font-family: var(--ui);
		font-size: var(--text-base);
		padding: var(--r2) var(--r4);
		cursor: pointer;
		outline: none;
		transition: background 0.1s ease;
	}
	:global(.select-item[data-highlighted]) {
		background: var(--porcelana);
		color: var(--tinta);
	}
	:global(.select-item[data-state="checked"]) {
		font-weight: var(--font-semi);
		color: var(--ouro);
	}

	@media (prefers-reduced-motion: reduce) {
		:global(.select-trigger),
		:global(.select-item) {
			transition-duration: 0ms;
		}
	}
</style>
