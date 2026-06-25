<script>
	/**
	 * FilterBar: barra de filtros reutilizável para curadoria e lojas.
	 * Usa bind: para two-way binding dos filtros.
	 */
	let {
		busca = $bindable(''),
		categoria = $bindable('cosméticos'),
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
</script>

<div class="filtros">
	{#if mostrarBusca}
		<label class="campo busca">
			<span class="rotulo">buscar na shopee</span>
			<input type="search" bind:value={busca} placeholder="perfume, sérum, batom…" class="entrada" />
		</label>
	{/if}
	<label class="campo">
		<span class="rotulo">categoria</span>
		<input type="text" list="categorias-sugeridas" bind:value={categoria} placeholder="ex.: cosméticos" class="entrada" />
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
	{#if mostrarQuantos && modo !== 'comparar'}
		<label class="campo">
			<span class="rotulo">quantos</span>
			<select bind:value={quantos} class="dado">
				<option value={6}>6</option>
				<option value={9}>9</option>
				<option value={12}>12</option>
			</select>
		</label>
	{/if}
	{#if mostrarExplorar && modo !== 'comparar'}
		<label class="campo-check" title="Reserva ~20% das vagas para produtos fora do topo">
			<input type="checkbox" bind:checked={explorar} />
			<span class="rotulo">explorar</span>
		</label>
	{/if}
</div>

<style>
	.filtros {
		display: flex; flex-wrap: wrap; align-items: flex-end; gap: var(--r4);
		margin-bottom: var(--r4); padding: var(--r4);
		background: var(--nevoa); border: 1px solid var(--linha); border-radius: var(--raio);
	}
	.campo { display: flex; flex-direction: column; gap: 5px; }
	.campo.busca { flex: 1 1 220px; }
	.campo-check {
		display: flex; align-items: center; gap: 6px;
		align-self: flex-end; padding-bottom: 9px; cursor: pointer;
	}
	.campo-check input { width: 16px; height: 16px; accent-color: var(--ouro); cursor: pointer; }
	.entrada {
		font-family: var(--ui); font-size: 0.95rem; padding: 9px 12px;
		border-radius: 10px; border: 1px solid var(--linha);
		background: var(--porcelana); color: var(--tinta); width: 100%;
	}
	.entrada::placeholder { color: var(--tinta-suave); opacity: 0.7; }
	.entrada.num { font-family: var(--mono); width: 5.5rem; }
	select {
		font-family: var(--mono); font-size: 0.9rem; padding: 9px 12px;
		border-radius: 10px; border: 1px solid var(--linha);
		background: var(--porcelana); color: var(--tinta);
	}
</style>
