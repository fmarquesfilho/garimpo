<script>
	/**
	 * SeletorGrupo — input com autocomplete dropdown para selecionar grupo WhatsApp.
	 * Controlado 100% por JavaScript (sem <select> nativo).
	 */
	let { grupos = [], carregando = false, erro = null, onselect = () => {} } = $props();

	let busca = $state('');
	let aberto = $state(false);
	let grupoSelecionado = $state(null);

	function filtrados() {
		if (!busca) return grupos;
		const lower = busca.toLowerCase();
		return grupos.filter((g) => g.nome.toLowerCase().includes(lower));
	}

	function selecionar(grupo) {
		grupoSelecionado = grupo;
		busca = grupo.nome;
		aberto = false;
		onselect(grupo.id);
	}

	function limpar() {
		grupoSelecionado = null;
		busca = '';
		onselect('');
	}

	function onInput() {
		aberto = true;
		// Se editou o texto depois de selecionar, limpa a seleção
		if (grupoSelecionado && busca !== grupoSelecionado.nome) {
			grupoSelecionado = null;
			onselect('');
		}
	}

	function onFocus() {
		aberto = true;
	}

	function onBlur(e) {
		// Se o clique foi dentro do container (dropdown), não fecha
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
		bind:value={busca}
		oninput={() => onselect(busca)}
		placeholder="ID do grupo (ex.: 123-456@g.us)"
	/>
{:else if grupos.length === 0}
	<div class="seletor-container">
		<input disabled placeholder="Nenhum grupo encontrado" />
	</div>
{:else}
	<div class="seletor-container">
		<div class="input-wrapper">
			<input
				type="text"
				bind:value={busca}
				oninput={onInput}
				onfocus={onFocus}
				onblur={onBlur}
				placeholder="Digite para buscar um grupo…"
				class:selecionado={grupoSelecionado}
			/>
			{#if grupoSelecionado}
				<button type="button" class="btn-limpar" onclick={limpar} title="Limpar">✕</button>
			{/if}
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
	input.selecionado {
		background: #f0fdf4;
		border-color: #86efac;
	}
	.btn-limpar {
		position: absolute;
		right: 8px;
		background: none;
		border: none;
		cursor: pointer;
		color: var(--tinta-suave);
		font-size: 0.9rem;
		padding: 4px;
	}
	.btn-limpar:hover { color: #b91c1c; }
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
	.erro-inline { font-size: 0.8rem; color: #b91c1c; margin-bottom: 4px; }
</style>
