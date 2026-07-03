<script>
	/**
	 * Tabs — barra de abas acessível usando Bits UI.
	 * @prop tabs — array de { id, label, badge?, badgeVariant? }
	 * @prop active — id da aba ativa (bindable)
	 * @prop children — conteúdo das abas (snippet)
	 */
	import { Tabs } from 'bits-ui';

	let {
		tabs = [],
		active = $bindable(''),
		children,
		...rest
	} = $props();
</script>

<Tabs.Root bind:value={active} {...rest}>
	<Tabs.List class="tabs-list">
		{#each tabs as tab (tab.id)}
			<Tabs.Trigger value={tab.id} class="tabs-trigger">
				{tab.label}
				{#if tab.badge}
					<span class="tabs-badge" class:alerta={tab.badgeVariant === 'alert'}>
						{tab.badge}
					</span>
				{/if}
			</Tabs.Trigger>
		{/each}
	</Tabs.List>
	{@render children()}
</Tabs.Root>

<style>
	:global(.tabs-list) {
		display: flex;
		gap: 2px;
		margin-bottom: var(--r5);
		border-bottom: 2px solid var(--linha);
		overflow-x: auto;
	}
	:global(.tabs-trigger) {
		padding: var(--r2) var(--r4);
		border: none;
		background: transparent;
		font-family: var(--ui);
		font-weight: var(--font-semi);
		font-size: var(--text-sm);
		color: var(--tinta-suave);
		cursor: pointer;
		border-bottom: 2px solid transparent;
		margin-bottom: -2px;
		display: flex;
		align-items: center;
		gap: var(--r2);
		white-space: nowrap;
		transition: color 0.15s ease, border-color 0.15s ease;
	}
	:global(.tabs-trigger:focus-visible) {
		outline: 2px solid var(--ouro);
		outline-offset: 2px;
	}
	:global(.tabs-trigger[data-state="active"]) {
		color: var(--tinta);
		border-bottom-color: var(--ouro);
	}
	:global(.tabs-trigger:hover:not([data-state="active"])) {
		color: var(--tinta);
	}

	.tabs-badge {
		font-size: var(--text-xs);
		background: var(--ouro-fundo);
		color: var(--ouro-escuro);
		padding: 1px var(--r2);
		border-radius: var(--raio-full);
		font-weight: var(--font-bold);
	}
	.tabs-badge.alerta {
		background: var(--erro-fundo);
		color: var(--erro-texto);
	}

	@media (max-width: 600px) {
		:global(.tabs-trigger) {
			padding: var(--r2) var(--r3);
			font-size: var(--text-xs);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		:global(.tabs-trigger) { transition-duration: 0ms; }
	}
</style>
