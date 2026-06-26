<script>
	/**
	 * FilterBar: barra de filtros com busca principal visível
	 * e filtros avançados colapsáveis.
	 */
	let {
		busca = $bindable(''),
		categoria = $bindable(''),
		comissaoMin = $bindable(0.07),
		vendasMin = $bindable(0),
		notaMin = $bindable(0),
		quantos = $bindable(9),
		explorar = $bindable(false),
		modo = 'nicho',
		mostrarQuantos = true,
		mostrarExplorar = true,
		mostrarBusca = true
	} = $props();

	let avancadoAberto = $state(false);

	// Conta filtros ativos (para badge)
	let filtrosAtivos = $derived(
		(comissaoMin > 0.07 ? 1 : 0) +
		(vendasMin > 0 ? 1 : 0) +
		(notaMin > 0 ? 1 : 0) +
		(categoria !== '' ? 1 : 0)
	);
</script>

<div class="filtros">
	<!-- Busca principal — sempre visível -->
	{#if mostrarBusca}
		<div class="busca-row">
			<input
				type="search"
				bind:value={busca}
				placeholder="🔍 Buscar produto… (ex: sérum, perfume, batom)"
				class="busca-input"
				onkeydown={(e) => { if (e.key === 'Enter') e.target.blur(); }}
			/>
			<button
				class="btn-avancado"
				class:ativo={avancadoAberto}
				onclick={() => (avancadoAberto = !avancadoAberto)}
				type="button"
			>
				⚙️ Filtros
				{#if filtrosAtivos > 0 && !avancadoAberto}
					<span class="filtro-badge">{filtrosAtivos}</span>
				{/if}
			</button>
		</div>
	{/if}

	<!-- Filtros avançados — colapsáveis -->
	{#if avancadoAberto}
		<div class="avancados">
			<label class="campo">
				<span class="rotulo">categoria</span>
				<input type="text" list="categorias-sugeridas" bind:value={categoria} placeholder="todas (opcional)" class="entrada" />
				<datalist id="categorias-sugeridas">
					<option value="cosméticos" />
					<option value="perfumaria" />
					<option value="skincare" />
					<option value="maquiagem" />
					<option value="bem-estar" />
					<option value="eletrônicos" />
					<option value="casa" />
					<option value="moda" />
				</datalist>
			</label>
			<label class="campo">
				<span class="rotulo">comissão mín.</span>
				<select bind:value={comissaoMin} class="dado">
					<option value={0.05}>5%</option>
					<option value={0.07}>7%</option>
					<option value={0.1}>10%</option>
					<option value={0.15}>15%</option>
				</select>
			</label>
			<label class="campo">
				<span class="rotulo">vendas mín.</span>
				<input type="number" min="0" step="1" bind:value={vendasMin} class="entrada num" />
			</label>
			<label class="campo">
				<span class="rotulo">nota mín.</span>
				<select bind:value={notaMin} class="dado">
					<option value={0}>todas</option>
					<option value={4}>4,0+</option>
					<option value={4.5}>4,5+</option>
				</select>
			</label>
		</div>
	{/if}
</div>

<style>
	.filtros {
		margin-bottom: var(--r5);
	}

	/* Busca principal */
	.busca-row {
		display: flex;
		gap: var(--r2);
	}
	.busca-input {
		flex: 1;
		font-family: var(--ui);
		font-size: 1rem;
		padding: 12px 16px;
		border-radius: var(--raio);
		border: 1px solid var(--linha);
		background: var(--branco);
		color: var(--tinta);
	}
	.busca-input:focus {
		outline: none;
		border-color: var(--ouro);
		box-shadow: 0 0 0 3px var(--ouro-fundo);
	}
	.busca-input::placeholder { color: var(--tinta-suave); opacity: 0.6; }

	.btn-avancado {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 10px 14px;
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		background: var(--porcelana);
		color: var(--tinta-suave);
		font-size: var(--text-sm);
		font-weight: 600;
		cursor: pointer;
		white-space: nowrap;
	}
	.btn-avancado:hover { border-color: var(--ouro); color: var(--tinta); }
	.btn-avancado.ativo { border-color: var(--ouro); background: var(--ouro-fundo); color: var(--ouro-escuro); }
	.filtro-badge {
		font-size: 0.65rem;
		background: var(--ouro);
		color: var(--branco);
		width: 16px;
		height: 16px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
	}

	/* Avançados */
	.avancados {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-end;
		gap: var(--r3);
		margin-top: var(--r3);
		padding: var(--r4);
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
	}
	.campo { display: flex; flex-direction: column; gap: 4px; }
	.campo-check {
		display: flex; align-items: center; gap: 6px;
		align-self: flex-end; padding-bottom: 8px; cursor: pointer;
	}
	.campo-check input { width: 16px; height: 16px; accent-color: var(--ouro); cursor: pointer; }
	.entrada {
		font-family: var(--ui); font-size: var(--text-base); padding: 8px 12px;
		border-radius: var(--raio-sm); border: 1px solid var(--linha);
		background: var(--porcelana); color: var(--tinta); width: 100%;
	}
	.entrada::placeholder { color: var(--tinta-suave); opacity: 0.7; }
	.entrada.num { font-family: var(--mono); width: 5.5rem; }
	select {
		font-family: var(--mono); font-size: var(--text-sm); padding: 8px 12px;
		border-radius: var(--raio-sm); border: 1px solid var(--linha);
		background: var(--porcelana); color: var(--tinta);
	}

	@media (max-width: 480px) {
		.busca-row { flex-direction: column; }
		.btn-avancado { justify-content: center; }
	}
</style>
