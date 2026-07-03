<script>
	/**
	 * TabBar — barra de abas horizontal com estilo padrão.
	 * @prop tabs — array de { id, label, badge? }
	 * @prop active — id da aba ativa (bindable)
	 */
	let { tabs = [], active = $bindable('') } = $props();
</script>

<nav class="tab-bar">
	{#each tabs as tab (tab.id)}
		<button
			class:ativa={active === tab.id}
			onclick={() => (active = tab.id)}
		>
			{tab.label}
			{#if tab.badge}
				<span class="badge-tab" class:alerta={tab.badgeVariant === 'alert'}>
					{tab.badge}
				</span>
			{/if}
		</button>
	{/each}
</nav>

<style>
	.tab-bar {
		display: flex;
		gap: 2px;
		margin-bottom: var(--r5);
		border-bottom: 2px solid var(--linha);
		overflow-x: auto;
	}
	button {
		padding: 8px 16px;
		border: none;
		background: transparent;
		font-weight: 600;
		font-size: 0.85rem;
		color: var(--tinta-suave);
		cursor: pointer;
		border-bottom: 2px solid transparent;
		margin-bottom: -2px;
		display: flex;
		align-items: center;
		gap: 6px;
		white-space: nowrap;
	}
	button.ativa {
		color: var(--tinta);
		border-bottom-color: var(--ouro);
	}
	button:hover:not(.ativa) { color: var(--tinta); }
	.badge-tab {
		font-size: var(--text-xs);
		background: var(--ouro-fundo);
		color: var(--ouro-escuro);
		padding: 1px var(--r2);
		border-radius: var(--raio-full);
		font-weight: var(--font-bold);
	}
	.badge-tab.alerta { background: var(--erro-fundo); color: var(--erro-texto); }

	@media (max-width: 600px) {
		button { padding: 8px 12px; font-size: 0.8rem; }
	}
</style>
