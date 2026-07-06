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
	let filtrosAtivos = $derived((comissaoMin > 0.07 ? 1 : 0) + (vendasMin > 0 ? 1 : 0) + (categoria !== '' ? 1 : 0));

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
		const match = todasCategorias.find((c) => c.nome.toLowerCase() === catInput.trim().toLowerCase());
		if (match) {
			categoria = match.nome;
		} else if (!catInput.trim()) {
			categoria = '';
		}
	});
</script>

<div class="mb-5">
	<!-- Busca principal — sempre visível -->
	{#if mostrarBusca}
		<div class="flex gap-2 max-[480px]:flex-col">
			<div class="flex-1 relative">
				<input
					type="search"
					bind:value={busca}
					placeholder="🔍 Buscar produto… (ex: sérum, perfume, batom)"
					class="w-full font-sans text-base py-3 pr-10 pl-4 rounded-md border border-border bg-background text-foreground placeholder:text-muted-foreground placeholder:opacity-60 focus:outline-none focus:border-primary focus:ring-2 focus:ring-ring/20"
					onkeydown={(e) => {
						if (e.key === 'Escape') {
							busca = '';
							/** @type {HTMLInputElement} */ (e.target).blur();
						} else if (e.key === 'Enter') /** @type {HTMLInputElement} */ (e.target).blur();
					}}
				/>
				{#if busca}
					<button
						class="absolute right-2.5 top-1/2 -translate-y-1/2 border-none bg-muted text-muted-foreground w-6 h-6 rounded-full text-xs cursor-pointer flex items-center justify-center hover:bg-border hover:text-foreground"
						onclick={() => (busca = '')}
						type="button"
						aria-label="Limpar busca">✕</button
					>
				{/if}
			</div>
			<button
				class="flex items-center gap-1 py-2.5 px-3.5 border border-border rounded-sm bg-muted text-muted-foreground text-sm font-semibold cursor-pointer whitespace-nowrap hover:border-primary hover:text-foreground max-[480px]:justify-center {avancadoAberto
					? 'border-primary bg-accent text-accent-foreground'
					: ''}"
				onclick={() => (avancadoAberto = !avancadoAberto)}
				type="button"
			>
				⚙️ Filtros
				{#if filtrosAtivos > 0 && !avancadoAberto}
					<span
						class="text-[0.65rem] bg-primary text-primary-foreground w-4 h-4 rounded-full flex items-center justify-center font-bold"
						>{filtrosAtivos}</span
					>
				{/if}
			</button>
		</div>
	{/if}

	<!-- Filtros avançados — colapsáveis -->
	{#if avancadoAberto}
		<div class="flex flex-wrap items-end gap-3 mt-3 p-4 bg-card border border-border rounded-sm">
			<label class="flex flex-col gap-1 relative">
				<span class="rotulo">categoria</span>
				<div class="relative">
					<input
						type="text"
						bind:value={catInput}
						placeholder="todas (digite para filtrar)"
						class="font-sans text-base py-2 px-3 rounded-sm border border-border bg-muted text-foreground w-full placeholder:text-muted-foreground placeholder:opacity-70"
						autocomplete="off"
					/>
					{#if categoria}
						<button
							class="absolute right-2 top-1/2 -translate-y-1/2 border-none bg-muted text-muted-foreground w-5 h-5 rounded-full text-[0.65rem] cursor-pointer flex items-center justify-center hover:bg-border hover:text-foreground"
							onclick={limparCategoria}
							type="button"
							aria-label="Limpar categoria">✕</button
						>
					{/if}
				</div>
				{#if catInput && sugestoes.length > 0 && catInput.toLowerCase() !== categoria.toLowerCase()}
					<ul
						class="absolute z-20 top-full left-0 right-0 mt-1 p-1 list-none bg-background border border-border rounded-sm shadow-md max-h-[200px] overflow-y-auto"
						role="listbox"
					>
						{#each sugestoes.slice(0, 8) as cat (cat.id)}
							<li role="option" aria-selected={categoria === cat.nome}>
								<button
									type="button"
									class="w-full py-2 px-2.5 border-none bg-transparent text-left text-sm text-foreground cursor-pointer rounded flex justify-between items-center hover:bg-accent hover:text-accent-foreground"
									onclick={() => selecionarCategoria(cat.nome)}
								>
									{cat.nome}
									<span class="text-[0.65rem] text-muted-foreground uppercase">{cat.marketplace}</span>
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			</label>
			<label class="flex flex-col gap-1">
				<span class="rotulo">comissão mín.</span>
				<select
					bind:value={comissaoMin}
					class="dado font-mono text-sm py-2 px-3 rounded-sm border border-border bg-muted text-foreground"
				>
					<option value={0.05}>5%</option>
					<option value={0.07}>7%</option>
					<option value={0.1}>10%</option>
					<option value={0.15}>15%</option>
				</select>
			</label>
			<label class="flex flex-col gap-1">
				<span class="rotulo">vendas mín.</span>
				<input
					type="number"
					min="0"
					step="1"
					bind:value={vendasMin}
					class="font-mono text-base py-2 px-3 rounded-sm border border-border bg-muted text-foreground w-[5.5rem]"
				/>
			</label>
		</div>
	{/if}
</div>
