<script>
	import { onMount } from 'svelte';
	import { listarDestinos, salvarDestino, deletarDestino, listarGruposWhatsApp } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';
	import SeletorGrupo from '$lib/SeletorGrupo.svelte';

	let destinos = $state([]);
	let carregando = $state(true);
	let erro = $state(null);
	let sucesso = $state('');

	// Form de novo destino
	let nome = $state('');
	let config = $state('');
	let tipo = $state('telegram');
	let salvando = $state(false);

	// Edição
	let editandoId = $state(null);
	let editNome = $state('');
	let editConfig = $state('');
	let editSalvando = $state(false);

	// Grupos WhatsApp (carregados sob demanda)
	let gruposWA = $state([]);
	let carregandoGrupos = $state(false);
	let erroGrupos = $state(null);

	onMount(async () => {
		await carregar();
		carregarGruposWA();
	});

	async function carregar() {
		carregando = true;
		erro = null;
		try {
			const r = await listarDestinos();
			destinos = r?.destinos ?? [];
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}
	}

	async function carregarGruposWA() {
		if (gruposWA.length > 0 || carregandoGrupos) return;
		carregandoGrupos = true;
		erroGrupos = null;
		try {
			const r = await listarGruposWhatsApp();
			gruposWA = r?.grupos ?? [];
		} catch (e) {
			erroGrupos = e.message;
		} finally {
			carregandoGrupos = false;
		}
	}

	function aoMudarTipo() {
		config = '';
		if (tipo === 'whatsapp') {
			carregarGruposWA();
		}
	}

	async function adicionar() {
		if (!nome.trim() || !config.trim()) return;
		salvando = true;
		erro = null;
		sucesso = '';
		try {
			const r = await salvarDestino({ nome: nome.trim(), config: config.trim(), tipo });
			destinos = [...destinos, r.destino];
			nome = '';
			config = '';
			sucesso = 'Destino adicionado!';
			setTimeout(() => (sucesso = ''), 3000);
		} catch (e) {
			erro = e.message;
		} finally {
			salvando = false;
		}
	}

	function iniciarEdicao(d) {
		editandoId = d.id;
		editNome = d.nome;
		editConfig = d.config;
	}

	function cancelarEdicao() {
		editandoId = null;
		editNome = '';
		editConfig = '';
	}

	async function salvarEdicao(d) {
		if (!editNome.trim() || !editConfig.trim()) return;
		editSalvando = true;
		erro = null;
		try {
			const atualizado = { id: d.id, nome: editNome.trim(), config: editConfig.trim(), tipo: d.tipo };
			await salvarDestino(atualizado);
			destinos = destinos.map(x => x.id === d.id ? { ...x, nome: atualizado.nome, config: atualizado.config } : x);
			editandoId = null;
			sucesso = 'Destino atualizado!';
			setTimeout(() => (sucesso = ''), 3000);
		} catch (e) {
			erro = e.message;
		} finally {
			editSalvando = false;
		}
	}

	async function remover(id) {
		if (!confirm(`Remover o destino "${id}"?`)) return;
		erro = null;
		try {
			await deletarDestino(id);
			destinos = destinos.filter((d) => d.id !== id);
		} catch (e) {
			erro = e.message;
		}
	}

	const tipoIcone = { telegram: '✈️', whatsapp: '💬' };
	const tipoLabel = { telegram: 'Telegram', whatsapp: 'WhatsApp' };
</script>

<svelte:head>
	<title>Configurações — Garimpei</title>
</svelte:head>

<section class="config-page">
	<h1>⚙️ Configurações</h1>

	{#if !$usuario}
		<div class="aviso">Faça login para gerenciar configurações.</div>
	{:else}
		<h2>📡 Destinos de publicação</h2>
		<p class="subtitulo">
			Configure os grupos e canais onde o Garimpei publica. Cada destino usa o bot
			configurado para o tipo (Telegram, WhatsApp). Adicione o bot como admin do grupo.
		</p>

		{#if erro}
			<div class="erro">{erro}</div>
		{/if}
		{#if sucesso}
			<div class="sucesso">{sucesso}</div>
		{/if}

		<!-- Form novo destino -->
		<form class="form-destino" onsubmit={(e) => { e.preventDefault(); adicionar(); }}>
			<div class="campo">
				<label for="tipo">Tipo</label>
				<select id="tipo" bind:value={tipo} onchange={aoMudarTipo}>
					<option value="telegram">✈️ Telegram</option>
					<option value="whatsapp">💬 WhatsApp</option>
				</select>
			</div>
			<div class="campo">
				<label for="nome">Nome</label>
				<input id="nome" bind:value={nome} placeholder="ex.: Ofertas Beleza" required />
			</div>
			<div class="campo">
				<label for="config">Destino ({tipoLabel[tipo]})</label>
				{#if tipo === 'whatsapp'}
					<SeletorGrupo
						grupos={gruposWA}
						carregando={carregandoGrupos}
						erro={erroGrupos}
						onselect={(id) => { config = id; }}
					/>
				{:else}
					<input id="config" bind:value={config} placeholder="@meucanal ou -1001234567890" required />
				{/if}
			</div>
			<button type="submit" disabled={salvando || !nome.trim() || !config.trim()}>
				{salvando ? 'Salvando…' : '+ Adicionar'}
			</button>
		</form>

		<!-- Lista -->
		{#if carregando}
			<p class="loading">Carregando…</p>
		{:else if destinos.length === 0}
			<p class="vazio">Nenhum destino cadastrado. Adicione o primeiro acima.</p>
		{:else}
			<div class="lista">
				{#each destinos as d (d.id)}
					{#if editandoId === d.id}
						<!-- Modo edição -->
						<div class="card-destino editando">
							<div class="edit-form">
								<div class="campo-edit">
									<label>Nome</label>
									<input bind:value={editNome} placeholder="Nome do destino" />
								</div>
								<div class="campo-edit">
									<label>Grupos</label>
									{#if d.tipo === 'whatsapp'}
										<SeletorGrupo
											grupos={gruposWA}
											carregando={carregandoGrupos}
											erro={erroGrupos}
											onselect={(id) => { editConfig = id; }}
											inicial={editConfig}
										/>
									{:else}
										<input bind:value={editConfig} placeholder="@canal ou chat_id" />
									{/if}
								</div>
								<div class="edit-acoes">
									<button class="btn-salvar" onclick={() => salvarEdicao(d)} disabled={editSalvando || !editNome.trim() || !editConfig.trim()}>
										{editSalvando ? 'Salvando…' : 'Salvar'}
									</button>
									<button class="btn-cancelar" onclick={cancelarEdicao}>Cancelar</button>
								</div>
							</div>
						</div>
					{:else}
						<!-- Modo visualização -->
						<div class="card-destino">
							<div class="info">
								<span class="tipo-badge">{tipoIcone[d.tipo] ?? '📤'} {tipoLabel[d.tipo] ?? d.tipo}</span>
								<strong>{d.nome}</strong>
								{#if d.tipo === 'whatsapp' && gruposWA.length > 0}
									<div class="grupos-lista">
										{#each d.config.split(',') as gid (gid)}
											{@const grupo = gruposWA.find(g => g.id === gid.trim())}
											<span class="grupo-nome">{grupo?.nome ?? gid.trim()}</span>
										{/each}
									</div>
								{:else}
									<code>{d.config}</code>
								{/if}
							</div>
							<div class="card-acoes">
								<button class="btn-editar" onclick={() => iniciarEdicao(d)} title="Editar">✎</button>
								<button class="btn-remover" onclick={() => remover(d.id)} title="Remover">✕</button>
							</div>
						</div>
					{/if}
				{/each}
			</div>
		{/if}
	{/if}
</section>

<style>
	.config-page { max-width: 900px; }
	h1 { font-size: 1.5rem; margin-bottom: var(--r6); }
	h2 { font-size: 1.1rem; margin: 0 0 var(--r2); }
	.subtitulo { color: var(--tinta-suave); font-size: 0.88rem; margin-bottom: var(--r5); }

	.aviso, .erro, .sucesso { padding: var(--r3) var(--r4); border-radius: 8px; font-size: 0.88rem; margin-bottom: var(--r4); }
	.aviso { background: var(--porcelana); color: var(--tinta-suave); }
	.erro { background: var(--erro-fundo); color: var(--erro-texto); border: 1px solid var(--erro-borda); }
	.sucesso { background: var(--sucesso-fundo); color: var(--sucesso-texto); border: 1px solid var(--sucesso-borda); }

	.form-destino {
		display: flex; flex-wrap: wrap; gap: var(--r3); align-items: flex-end;
		margin-bottom: var(--r5); padding: var(--r4);
		border: 1px solid var(--linha); border-radius: var(--raio); background: var(--porcelana);
	}
	.campo { flex: 1; min-width: 140px; display: flex; flex-direction: column; gap: 4px; }
	.campo label { font-size: 0.78rem; font-weight: 600; color: var(--tinta-suave); }
	.campo :global(input), .campo :global(select) { padding: 8px 12px; border: 1px solid var(--linha); border-radius: 8px; font-size: 0.9rem; }
	.campo :global(input:focus), .campo :global(select:focus) { outline: 2px solid var(--ouro); outline-offset: 1px; }
	.form-destino > button {
		padding: 8px 20px; background: var(--ouro); color: white;
		font-weight: 600; font-size: 0.88rem; border: none; border-radius: 8px; cursor: pointer;
	}
	.form-destino > button:disabled { opacity: 0.5; cursor: not-allowed; }

	.loading, .vazio { color: var(--tinta-suave); font-size: 0.9rem; }
	.lista { display: flex; flex-direction: column; gap: var(--r3); }
	.card-destino {
		display: flex; align-items: center; justify-content: space-between;
		padding: var(--r3) var(--r4); border: 1px solid var(--linha); border-radius: var(--raio-sm); background: var(--branco);
	}
	.card-destino.editando {
		border-color: var(--ouro); background: var(--branco)beb;
	}
	.card-destino .info { display: flex; flex-direction: column; gap: 2px; }
	.card-destino strong { font-size: 0.92rem; }
	.card-destino code { font-size: 0.8rem; color: var(--tinta-suave); }
	.tipo-badge { font-size: 0.72rem; font-weight: 600; color: var(--tinta-suave); }

	.card-acoes { display: flex; gap: 4px; }
	.btn-editar, .btn-remover {
		background: none; border: 1px solid var(--linha); border-radius: 6px;
		width: 32px; height: 32px; display: flex; align-items: center; justify-content: center;
		cursor: pointer; color: var(--tinta-suave); font-size: 1rem;
	}
	.btn-editar:hover { color: var(--ouro); border-color: var(--ouro); background: var(--branco)beb; }
	.btn-remover:hover { color: var(--erro-texto); border-color: var(--erro-borda); background: var(--erro-fundo); }

	.edit-form { width: 100%; display: flex; flex-direction: column; gap: var(--r3); }
	.campo-edit { display: flex; flex-direction: column; gap: 4px; }
	.campo-edit label { font-size: 0.78rem; font-weight: 600; color: var(--tinta-suave); }
	.campo-edit input { padding: 8px 12px; border: 1px solid var(--linha); border-radius: 8px; font-size: 0.9rem; }
	.campo-edit input:focus { outline: 2px solid var(--ouro); outline-offset: 1px; }
	.edit-acoes { display: flex; gap: var(--r2); }
	.btn-salvar {
		padding: 6px 16px; background: var(--ouro); color: white;
		font-weight: 600; font-size: 0.82rem; border: none; border-radius: 6px; cursor: pointer;
	}
	.btn-salvar:disabled { opacity: 0.5; cursor: not-allowed; }
	.btn-cancelar {
		padding: 6px 16px; background: transparent; color: var(--tinta-suave);
		font-weight: 600; font-size: 0.82rem; border: 1px solid var(--linha); border-radius: 6px; cursor: pointer;
	}

	.grupos-lista { display: flex; flex-direction: column; gap: 2px; margin-top: 2px; }
	.grupo-nome { font-size: 0.78rem; color: var(--tinta-suave); padding: 1px 6px; background: var(--sucesso-fundo); border-radius: 4px; border: 1px solid var(--sucesso-borda); }
</style>
