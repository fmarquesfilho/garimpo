<script>
	/**
	 * SeletorGrupo — select com filtro para grupos de WhatsApp.
	 *
	 * Em vez de usar bind:value no <select> (que tem bugs com opções dinâmicas
	 * no Svelte 5), usa um callback onselect para notificar o parent.
	 */

	/** @type {{grupos: Array<{id: string, nome: string}>, value: string, carregando: boolean, erro: string|null, onselect: (id: string) => void}} */
	let { grupos = [], value = '', carregando = false, erro = null, onselect = () => {} } = $props();

	let filtro = $state('');

	let visiveis = $derived(
		filtro
			? grupos.filter((g) => g.nome.toLowerCase().includes(filtro.toLowerCase()))
			: grupos
	);

	function handleChange(e) {
		onselect(e.target.value);
	}
</script>

{#if carregando}
	<select disabled>
		<option>Carregando grupos…</option>
	</select>
{:else if erro}
	<div class="erro-inline">{erro}</div>
	<input value={value} oninput={(e) => onselect(e.target.value)} placeholder="ID do grupo (ex.: 123-456@g.us)" />
{:else if grupos.length === 0}
	<select disabled>
		<option>Nenhum grupo encontrado</option>
	</select>
{:else}
	<input
		type="text"
		bind:value={filtro}
		placeholder="Filtrar grupos…"
		class="filtro-grupo"
	/>
	<select onchange={handleChange}>
		<option value="" selected={!value}>Selecione um grupo… ({visiveis.length})</option>
		{#each visiveis as g (g.id)}
			<option value={g.id} selected={g.id === value}>{g.nome}</option>
		{/each}
	</select>
{/if}

<style>
	.filtro-grupo {
		padding: 6px 10px;
		border: 1px solid var(--linha);
		border-radius: 6px;
		font-size: 0.82rem;
		margin-bottom: 4px;
		width: 100%;
	}
	.erro-inline { font-size: 0.8rem; color: #b91c1c; margin-bottom: 4px; }
	select, input {
		padding: 8px 12px;
		border: 1px solid var(--linha);
		border-radius: 8px;
		font-size: 0.9rem;
		width: 100%;
	}
	select:focus, input:focus { outline: 2px solid var(--ouro); outline-offset: 1px; }
</style>
