<script>
	let { valor = $bindable('nicho'), opcoes = ['nicho', 'diversificada', 'comparar'] } = $props();
	const legenda = {
		nicho: 'Nicho',
		diversificada: 'Diversificada',
		comparar: 'Comparar'
	};
</script>

<div class="alternador" role="tablist" aria-label="estratégia de curadoria">
	{#each opcoes as op}
		<button
			role="tab"
			aria-selected={valor === op}
			class:ativo={valor === op}
			class:nicho={op === 'nicho'}
			class:div={op === 'diversificada'}
			onclick={() => (valor = op)}
		>
			{legenda[op] ?? op}
		</button>
	{/each}
</div>

<style>
	.alternador {
		display: inline-flex;
		gap: 4px;
		padding: 4px;
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: 999px;
	}
	button {
		border: none;
		background: transparent;
		color: var(--tinta-suave);
		font-weight: 600;
		font-size: 0.9rem;
		padding: 8px 18px;
		border-radius: 999px;
		transition: all 0.15s ease;
	}
	button:hover {
		color: var(--tinta);
	}
	.ativo {
		color: #fff;
		background: var(--tinta);
	}
	.ativo.nicho {
		background: var(--rosa);
	}
	.ativo.div {
		background: var(--ardosia);
	}
</style>
