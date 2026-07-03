<script>
	/**
	 * SeletorGrupo — input com autocomplete para selecionar 1-5 grupos de WhatsApp.
	 * Seleções aparecem como chips/badges acima do input.
	 */
	const MAX_GRUPOS = 5;

	let { grupos = [], carregando = false, erro = null, onselect = () => {}, inicial = '' } = $props();

	let busca = $state('');
	let aberto = $state(false);
	let selecionados = $state([]);

	// Inicializa com grupos pré-selecionados (modo edição)
	$effect(() => {
		if (inicial && grupos.length > 0 && selecionados.length === 0) {
			const ids = inicial
				.split(',')
				.map((id) => id.trim())
				.filter(Boolean);
			const encontrados = ids.map((id) => grupos.find((g) => g.id === id)).filter(Boolean);
			if (encontrados.length > 0) {
				selecionados = encontrados;
			}
		}
	});

	function filtrados() {
		const idsJaSelecionados = new Set(selecionados.map((g) => g.id));
		let lista = grupos.filter((g) => !idsJaSelecionados.has(g.id));
		if (busca) {
			const lower = busca.toLowerCase();
			lista = lista.filter((g) => g.nome.toLowerCase().includes(lower));
		}
		return lista;
	}

	function selecionar(grupo) {
		if (selecionados.length >= MAX_GRUPOS) return;
		selecionados = [...selecionados, grupo];
		busca = '';
		aberto = false;
		emitir();
	}

	function removerGrupo(id) {
		selecionados = selecionados.filter((g) => g.id !== id);
		emitir();
	}

	function emitir() {
		// Emite os IDs separados por vírgula (formato do config)
		const ids = selecionados.map((g) => g.id).join(',');
		onselect(ids);
	}

	function onInput() {
		aberto = true;
	}

	function onFocus() {
		aberto = true;
	}

	function onBlur(e) {
		const container = e.target.closest('.seletor-container');
		if (container && container.contains(e.relatedTarget)) return;
		setTimeout(() => {
			aberto = false;
		}, 150);
	}
</script>

{#if carregando}
	<div class="seletor-container relative w-full">
		<input
			disabled
			placeholder="Carregando grupos…"
			class="w-full rounded-lg border border-border px-3 py-2 text-[0.9rem]"
		/>
	</div>
{:else if erro}
	<div class="mb-1 text-xs text-[var(--erro-texto)]">{erro}</div>
	<input
		class="w-full rounded-lg border border-border px-3 py-2 text-[0.9rem] focus:outline-2 focus:outline-ouro focus:outline-offset-1"
		value={selecionados.map((g) => g.id).join(',')}
		oninput={(e) => onselect(/** @type {HTMLInputElement} */ (e.target).value)}
		placeholder="IDs dos grupos separados por vírgula"
	/>
{:else if grupos.length === 0}
	<div class="seletor-container relative w-full">
		<input
			disabled
			placeholder="Nenhum grupo encontrado"
			class="w-full rounded-lg border border-border px-3 py-2 text-[0.9rem]"
		/>
	</div>
{:else}
	<div class="seletor-container relative w-full">
		{#if selecionados.length > 0}
			<div class="mb-1.5 flex flex-wrap gap-1">
				{#each selecionados as g (g.id)}
					<span
						class="inline-flex items-center gap-1 rounded-md border border-[var(--sucesso-borda)] bg-[var(--sucesso-fundo)] px-2 py-0.5 text-xs text-[var(--sucesso-texto)]"
					>
						{g.nome}
						<button
							type="button"
							class="cursor-pointer border-none bg-transparent p-0 px-0.5 text-xs leading-none text-[var(--sucesso-texto)] hover:text-[var(--erro-texto)]"
							onclick={() => removerGrupo(g.id)}
							title="Remover">✕</button
						>
					</span>
				{/each}
			</div>
		{/if}

		{#if selecionados.length < MAX_GRUPOS}
			<div class="relative flex items-center">
				<input
					type="text"
					class="w-full rounded-lg border border-border px-3 py-2 text-[0.9rem] focus:outline-2 focus:outline-ouro focus:outline-offset-1"
					bind:value={busca}
					oninput={onInput}
					onfocus={onFocus}
					onblur={onBlur}
					placeholder={selecionados.length === 0 ? 'Digite para buscar um grupo…' : 'Adicionar outro grupo…'}
				/>
			</div>
			{#if aberto}
				{@const lista = filtrados()}
				{#if lista.length > 0}
					<ul
						class="absolute top-full right-0 left-0 z-100 mt-1 max-h-[200px] list-none overflow-y-auto rounded-lg border border-border bg-[var(--branco)] p-1 shadow-[var(--sombra)]"
						onpointerdown={(e) => e.preventDefault()}
					>
						{#each lista as g (g.id)}
							<li>
								<button
									type="button"
									class="block w-full cursor-pointer rounded-md border-none bg-transparent px-3 py-2 text-left text-[0.88rem] hover:bg-porcelana"
									onclick={() => selecionar(g)}
								>
									{g.nome}
								</button>
							</li>
						{/each}
					</ul>
				{:else}
					<ul
						class="absolute top-full right-0 left-0 z-100 mt-1 max-h-[200px] list-none overflow-y-auto rounded-lg border border-border bg-[var(--branco)] p-1 shadow-[var(--sombra)]"
					>
						<li class="px-3 py-2 text-sm text-tinta-suave">Nenhum grupo encontrado</li>
					</ul>
				{/if}
			{/if}
		{:else}
			<p class="mt-1 text-xs text-tinta-suave">Limite de {MAX_GRUPOS} grupos atingido</p>
		{/if}
	</div>
{/if}
