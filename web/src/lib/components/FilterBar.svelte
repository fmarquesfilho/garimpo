<script>
	/**
	 * FilterBar: barra de filtros com busca principal visível
	 * e filtros avançados colapsáveis.
	 * Categorias carregadas dinamicamente da API (com fallback).
	 */
	import { onMount } from 'svelte';
	import { buscarCategorias, filtrarCategorias } from '$lib/categorias.js';

	let {
		busca = $bindable(''),
		categoria = $bindable(''),
		comissaoMin = $bindable(0.07),
		vendasMin = $bindable(0),
		quantos = $bindable(9),
		explorar = $bindable(false),
		mostrarBusca = true
	} = $props();

	let avancadoAberto = $state(false);
	let todasCategorias = $state([]);
	let catInput = $state('');
	let sugestoes = $derived(filtrarCategorias(todasCategorias, catInput));

	// Conta filtros ativos (para badge)
	let filtrosAtivos = $derived(
		(comissaoMin > 0.07 ? 1 : 0) +
		(vendasMin > 0 ? 1 : 0) +
		(categoria !== '' ? 1 : 0)
	);

	onMount(async () => {
		todasCategorias = await buscarCategorias();
	});

	function selecionarCategoria(nome) {
		categoria = nome;
		catInput = nome;
	}

	function limparCategoria() {
		categoria = '';
		catInput = '';
	}

	// Sync catInput → categoria quando o usuário digita (e o valor bate com uma sugestão)
	$effect(() => {
		const match = todasCategorias.find(c => c.nome.toLowerCase() === catInput.trim().toLowerCase());
		if (match) {
			categoria = match.nome;
		} else if (!catInput.trim()) {
			categoria = '';
		}
	});
</script>

<div class="filtros">
	<!-- Busca principal — sempre visível -->
	{#if mostrarBusca}
		<div class="busca-row">
			<div class="busca-wrapper">
				<input
					type="search"
					bind:value={busca}
					placeholder="🔍 Buscar produto… (ex: sérum, perfume, batom)"
					class="busca-input"
					onkeydown={(e) => { if (e.key === 'Escape') { busca = ''; /** @type {HTMLInputElement} */ (e.target).blur(); } else if (e.key === 'Enter') /** @type {HTMLInputElement} */ (e.target).blur(); }}
				/>
				{#if busca}
					<button class="btn-limpar" onclick={() => (busca = '')} type="button" aria-label="Limpar busca">✕</button>
				{/if}
			</div>
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
				<div class="cat-wrapper">
					<input
						type="text"
						bind:value={catInput}
						placeholder="todas (digite para filtrar)"
						class="entrada"
						autocomplete="off"
					/>
					{#if categoria}
						<button class="btn-limpar-cat" onclick={limparCategoria} type="button" aria-label="Limpar categoria">✕</button>
					{/if}
				</div>
				{#if catInput && sugestoes.length > 0 && catInput.toLowerCase() !== categoria.toLowerCase()}
					<ul class="cat-sugestoes" role="listbox">
						{#each sugestoes.slice(0, 8) as cat (cat.id)}
							<li role="option" aria-selected={categoria === cat.nome}>
								<button type="button" class="cat-opcao" onclick={() => selecionarCategoria(cat.nome)}>
									{cat.nome}
									<span class="cat-mp">{cat.marketplace}</span>
								</button>
							</li>
						{/each}
					</ul>
				{/if}
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
	.busca-wrapper {
		flex: 1;
		position: relative;
	}
	.busca-input {
		width: 100%;
		font-family: var(--ui);
		font-size: 1rem;
		padding: 12px 40px 12px 16px;
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
	.btn-limpar {
		position: absolute;
		right: 10px;
		top: 50%;
		transform: translateY(-50%);
		border: none;
		background: var(--porcelana);
		color: var(--tinta-suave);
		width: 24px;
		height: 24px;
		border-radius: 50%;
		font-size: 0.75rem;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.btn-limpar:hover { background: var(--linha); color: var(--tinta); }

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
	.campo { display: flex; flex-direction: column; gap: 4px; position: relative; }
	.entrada {
		font-family: var(--ui); font-size: var(--text-base); padding: 8px 12px;
		border-radius: var(--raio-sm); border: 1px solid var(--linha);
		background: var(--porcelana); color: var(--tinta); width: 100%;
	}
	.entrada::placeholder { color: var(--tinta-suave); opacity: 0.7; }
	.entrada.num { font-family: var(--mono); width: 5.5rem; }
	.cat-wrapper { position: relative; }
	.btn-limpar-cat {
		position: absolute; right: 8px; top: 50%; transform: translateY(-50%);
		border: none; background: var(--porcelana); color: var(--tinta-suave);
		width: 20px; height: 20px; border-radius: 50%; font-size: 0.65rem;
		cursor: pointer; display: flex; align-items: center; justify-content: center;
	}
	.btn-limpar-cat:hover { background: var(--linha); color: var(--tinta); }
	.cat-sugestoes {
		position: absolute; z-index: 20; top: 100%; left: 0; right: 0;
		margin: 4px 0 0; padding: 4px; list-style: none;
		background: var(--branco); border: 1px solid var(--linha);
		border-radius: var(--raio-sm); box-shadow: 0 4px 12px rgba(0,0,0,0.08);
		max-height: 200px; overflow-y: auto;
	}
	.cat-opcao {
		width: 100%; padding: 8px 10px; border: none; background: none;
		text-align: left; font-size: var(--text-sm); color: var(--tinta);
		cursor: pointer; border-radius: 4px; display: flex; justify-content: space-between; align-items: center;
	}
	.cat-opcao:hover { background: var(--ouro-fundo); color: var(--ouro-escuro); }
	.cat-mp { font-size: 0.65rem; color: var(--tinta-suave); text-transform: uppercase; }
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
