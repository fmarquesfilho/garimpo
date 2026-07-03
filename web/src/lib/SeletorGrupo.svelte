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
			const ids = inicial.split(',').map(id => id.trim()).filter(Boolean);
			const encontrados = ids.map(id => grupos.find(g => g.id === id)).filter(Boolean);
			if (encontrados.length > 0) {
				selecionados = encontrados;
			}
		}
	});

	function filtrados() {
		const idsJaSelecionados = new Set(selecionados.map(g => g.id));
		let lista = grupos.filter(g => !idsJaSelecionados.has(g.id));
		if (busca) {
			const lower = busca.toLowerCase();
			lista = lista.filter(g => g.nome.toLowerCase().includes(lower));
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
		selecionados = selecionados.filter(g => g.id !== id);
		emitir();
	}

	function emitir() {
		// Emite os IDs separados por vírgula (formato do config)
		const ids = selecionados.map(g => g.id).join(',');
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
		setTimeout(() => { aberto = false; }, 150);
	}
</script>

{#if carregando}
	<div class="seletor-container">
		<input disabled placeholder="Carregando grupos…" />
	</div>
{:else if erro}
	<div class="erro-inline">{erro}</div>
	<input
		value={selecionados.map(g => g.id).join(',')}
		oninput={(e) => onselect(e.target.value)}
		placeholder="IDs dos grupos separados por vírgula"
	/>
{:else if grupos.length === 0}
	<div class="seletor-container">
		<input disabled placeholder="Nenhum grupo encontrado" />
	</div>
{:else}
	<div class="seletor-container">
		{#if selecionados.length > 0}
			<div class="chips">
				{#each selecionados as g (g.id)}
					<span class="chip">
						{g.nome}
						<button type="button" onclick={() => removerGrupo(g.id)} title="Remover">✕</button>
					</span>
				{/each}
			</div>
		{/if}

		{#if selecionados.length < MAX_GRUPOS}
			<div class="input-wrapper">
				<input
					type="text"
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
					<ul class="dropdown" onpointerdown={(e) => e.preventDefault()}>
						{#each lista as g (g.id)}
							<li>
								<button type="button" onclick={() => selecionar(g)}>
									{g.nome}
								</button>
							</li>
						{/each}
					</ul>
				{:else}
					<ul class="dropdown">
						<li class="vazio">Nenhum grupo encontrado</li>
					</ul>
				{/if}
			{/if}
		{:else}
			<p class="limite">Limite de {MAX_GRUPOS} grupos atingido</p>
		{/if}
	</div>
{/if}

<style>
	.seletor-container {
		position: relative;
		width: 100%;
	}
	.input-wrapper {
		position: relative;
		display: flex;
		align-items: center;
	}
	input {
		padding: 8px 12px;
		border: 1px solid var(--linha);
		border-radius: 8px;
		font-size: 0.9rem;
		width: 100%;
	}
	input:focus { outline: 2px solid var(--ouro); outline-offset: 1px; }
	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		margin-bottom: 6px;
	}
	.chip {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 3px 8px;
		background: var(--sucesso-fundo);
		border: 1px solid var(--sucesso-borda);
		border-radius: 6px;
		font-size: 0.78rem;
		color: var(--sucesso-texto);
	}
	.chip button {
		background: none;
		border: none;
		cursor: pointer;
		color: var(--sucesso-texto);
		font-size: 0.75rem;
		padding: 0 2px;
		line-height: 1;
	}
	.chip button:hover { color: var(--erro-texto); }
	.limite {
		font-size: 0.78rem;
		color: var(--tinta-suave);
		margin: 4px 0 0;
	}
	.dropdown {
		position: absolute;
		top: 100%;
		left: 0;
		right: 0;
		max-height: 200px;
		overflow-y: auto;
		background: white;
		border: 1px solid var(--linha);
		border-radius: 8px;
		box-shadow: 0 4px 12px rgba(0,0,0,0.1);
		margin: 4px 0 0;
		padding: 4px;
		list-style: none;
		z-index: 100;
	}
	.dropdown li button {
		display: block;
		width: 100%;
		text-align: left;
		padding: 8px 12px;
		border: none;
		background: none;
		cursor: pointer;
		font-size: 0.88rem;
		border-radius: 6px;
	}
	.dropdown li button:hover {
		background: var(--porcelana);
	}
	.dropdown li.vazio {
		padding: 8px 12px;
		color: var(--tinta-suave);
		font-size: 0.85rem;
	}
	.erro-inline { font-size: 0.8rem; color: var(--erro-texto); margin-bottom: 4px; }
</style>
