<script>
	/**
	 * SeletorGrupo — select com filtro para grupos de WhatsApp.
	 *
	 * Props:
	 *   grupos: Array<{id, nome}> — lista de grupos disponíveis
	 *   value: string — group ID selecionado (bind:value)
	 *   carregando: boolean
	 *   erro: string|null
	 */
	let { grupos = [], value = $bindable(''), carregando = false, erro = null } = $props();

	let filtro = $state('');

	// Filtra grupos pelo texto digitado
	function filtrar(lista, texto) {
		if (!texto) return lista;
		const lower = texto.toLowerCase();
		return lista.filter((g) => g.nome.toLowerCase().includes(lower));
	}

	function onSelect(e) {
		value = e.target.value;
	}
</script>

{#if carregando}
	<select disabled>
		<option>Carregando grupos…</option>
	</select>
{:else if erro}
	<div class="erro-inline">{erro}</div>
	<input bind:value placeholder="ID do grupo (ex.: 123-456@g.us)" />
{:else if grupos.length === 0}
	<select disabled>
		<option>Nenhum grupo encontrado</option>
	</select>
{:else}
	{@const visiveis = filtrar(grupos, filtro)}
	<input
		type="text"
		bind:value={filtro}
		placeholder="Filtrar grupos…"
		class="filtro-grupo"
	/>
	<select value={value} onchange={onSelect}>
		<option value="">Selecione um grupo… ({visiveis.length})</option>
		{#each visiveis as g (g.id)}
			<option value={g.id}>{g.nome}</option>
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
