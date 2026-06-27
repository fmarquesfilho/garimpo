<script>
	let { value = $bindable(7), options = [1, 3, 7, 14, 30] } = $props();
</script>

<div class="periodo-selector">
	{#each options as dias}
		<button class:ativo={value === dias} onclick={() => (value = dias)}>
			{dias === 1 ? 'Hoje' : `${dias} dias`}
		</button>
	{/each}
</div>

<style>
	.periodo-selector {
		display: flex;
		gap: 2px;
		background: var(--porcelana);
		border-radius: var(--raio-sm);
		padding: 3px;
		border: 1px solid var(--linha);
	}
	button {
		padding: 6px 14px;
		border: none;
		border-radius: var(--raio-sm);
		background: transparent;
		font-size: 0.82rem;
		font-weight: 600;
		color: var(--tinta-suave);
		cursor: pointer;
	}
	button.ativo {
		background: var(--branco);
		color: var(--tinta);
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
	}
</style>
