<script>
	import { onMount } from 'svelte';
	import { listarDestinos, salvarDestino, deletarDestino } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';
	import SeletorGrupo from '$lib/SeletorGrupo.svelte';
	import { Alert, Button, Dialog, DropdownMenu } from '$lib/components/ui';

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

	// Grupos WhatsApp (placeholder — carregados sob demanda futuramente)
	let gruposWA = $state([]);
	let carregandoGrupos = $state(false);
	let erroGrupos = $state(null);

	// Dialog de confirmação
	let dialogRemover = $state(false);
	let destinoParaRemover = $state(null);

	onMount(() => {
		carregar();
	});

	function aoMudarTipo() {
		config = '';
	}

	async function carregar() {
		carregando = true;
		erro = null;
		try {
			destinos = (await listarDestinos())?.destinos ?? [];
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
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
			destinos = destinos.map((x) => (x.id === d.id ? { ...x, nome: atualizado.nome, config: atualizado.config } : x));
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
		erro = null;
		try {
			await deletarDestino(id);
			destinos = destinos.filter((d) => d.id !== id);
			dialogRemover = false;
			destinoParaRemover = null;
		} catch (e) {
			erro = e.message;
		}
	}

	function pedirRemocao(d) {
		destinoParaRemover = d;
		dialogRemover = true;
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
			Configure os grupos e canais onde o Garimpei publica. Cada destino usa o bot configurado para o tipo (Telegram,
			WhatsApp). Adicione o bot como admin do grupo.
		</p>

		{#if erro}
			<Alert variant="error">{erro}</Alert>
		{/if}
		{#if sucesso}
			<Alert variant="success">{sucesso}</Alert>
		{/if}

		<!-- Form novo destino -->
		<form
			class="form-destino"
			onsubmit={(e) => {
				e.preventDefault();
				adicionar();
			}}
		>
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
						onselect={(id) => {
							config = id;
						}}
					/>
				{:else}
					<input id="config" bind:value={config} placeholder="@meucanal ou -1001234567890" required />
				{/if}
			</div>
			<Button type="submit" disabled={salvando || !nome.trim() || !config.trim()}>
				{salvando ? 'Salvando…' : '+ Adicionar'}
			</Button>
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
									<label
										>Nome
										<input bind:value={editNome} placeholder="Nome do destino" />
									</label>
								</div>
								<div class="campo-edit">
									<span class="campo-edit-label">Grupos</span>
									{#if d.tipo === 'whatsapp'}
										<SeletorGrupo
											grupos={gruposWA}
											carregando={carregandoGrupos}
											erro={erroGrupos}
											onselect={(id) => {
												editConfig = id;
											}}
											inicial={editConfig}
										/>
									{:else}
										<input bind:value={editConfig} placeholder="@canal ou chat_id" />
									{/if}
								</div>
								<div class="edit-acoes">
									<Button
										size="sm"
										onclick={() => salvarEdicao(d)}
										disabled={editSalvando || !editNome.trim() || !editConfig.trim()}
									>
										{editSalvando ? 'Salvando…' : 'Salvar'}
									</Button>
									<Button variant="ghost" size="sm" onclick={cancelarEdicao}>Cancelar</Button>
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
											{@const grupo = gruposWA.find((g) => g.id === gid.trim())}
											<span class="grupo-nome">{grupo?.nome ?? gid.trim()}</span>
										{/each}
									</div>
								{:else}
									<code>{d.config}</code>
								{/if}
							</div>
							<div class="card-acoes">
								<DropdownMenu
									items={[
										{ label: '✎ Editar', onclick: () => iniciarEdicao(d) },
										{ label: '✕ Remover', onclick: () => pedirRemocao(d), destructive: true }
									]}
								>
									<button class="btn-menu" aria-label="Ações">⋮</button>
								</DropdownMenu>
							</div>
						</div>
					{/if}
				{/each}
			</div>
		{/if}
	{/if}

	<Dialog bind:open={dialogRemover} title="Remover destino" description="Tem certeza que quer remover este destino?">
		{#if destinoParaRemover}
			<p>O destino <strong>{destinoParaRemover.nome}</strong> será removido permanentemente.</p>
			<div class="dialog-acoes">
				<Button variant="danger" onclick={() => remover(destinoParaRemover.id)}>Remover</Button>
				<Button
					variant="ghost"
					onclick={() => {
						dialogRemover = false;
					}}>Cancelar</Button
				>
			</div>
		{/if}
	</Dialog>
</section>

<style>
	.config-page {
		max-width: 900px;
	}
	h1 {
		font-size: 1.5rem;
		margin-bottom: var(--r6);
	}
	h2 {
		font-size: 1.1rem;
		margin: 0 0 var(--r2);
	}
	.subtitulo {
		color: var(--tinta-suave);
		font-size: 0.88rem;
		margin-bottom: var(--r5);
	}

	.aviso {
		padding: var(--r3) var(--r4);
		border-radius: var(--raio-sm);
		font-size: 0.88rem;
		margin-bottom: var(--r4);
		background: var(--porcelana);
		color: var(--tinta-suave);
	}

	.form-destino {
		display: flex;
		flex-wrap: wrap;
		gap: var(--r3);
		align-items: flex-end;
		margin-bottom: var(--r5);
		padding: var(--r4);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		background: var(--porcelana);
	}
	.campo {
		flex: 1;
		min-width: 140px;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.campo label {
		font-size: 0.78rem;
		font-weight: 600;
		color: var(--tinta-suave);
	}
	.campo :global(input) {
		padding: 8px 12px;
		border: 1px solid var(--linha);
		border-radius: 8px;
		font-size: 0.9rem;
	}
	.campo :global(input:focus) {
		outline: 2px solid var(--ouro);
		outline-offset: 1px;
	}
	.loading,
	.vazio {
		color: var(--tinta-suave);
		font-size: 0.9rem;
	}
	.lista {
		display: flex;
		flex-direction: column;
		gap: var(--r3);
	}
	.card-destino {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--r3) var(--r4);
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		background: var(--branco);
	}
	.card-destino.editando {
		border-color: var(--ouro);
		background: var(--ouro-fundo);
	}
	.card-destino .info {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.card-destino strong {
		font-size: 0.92rem;
	}
	.card-destino code {
		font-size: 0.8rem;
		color: var(--tinta-suave);
	}
	.tipo-badge {
		font-size: 0.72rem;
		font-weight: 600;
		color: var(--tinta-suave);
	}

	.card-acoes {
		display: flex;
		gap: 4px;
	}
	.btn-menu {
		background: none;
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		color: var(--tinta-suave);
		font-size: var(--text-lg);
	}
	.btn-menu:hover {
		border-color: var(--ouro);
		color: var(--ouro);
	}

	.edit-form {
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: var(--r3);
	}
	.campo-edit {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.campo-edit label {
		font-size: 0.78rem;
		font-weight: 600;
		color: var(--tinta-suave);
	}
	.campo-edit input {
		padding: 8px 12px;
		border: 1px solid var(--linha);
		border-radius: 8px;
		font-size: 0.9rem;
	}
	.campo-edit input:focus {
		outline: 2px solid var(--ouro);
		outline-offset: 1px;
	}
	.edit-acoes {
		display: flex;
		gap: var(--r2);
	}

	.grupos-lista {
		display: flex;
		flex-direction: column;
		gap: 2px;
		margin-top: 2px;
	}
	.grupo-nome {
		font-size: 0.78rem;
		color: var(--tinta-suave);
		padding: 1px 6px;
		background: var(--sucesso-fundo);
		border-radius: 4px;
		border: 1px solid var(--sucesso-borda);
	}
	.dialog-acoes {
		display: flex;
		gap: var(--r3);
		margin-top: var(--r4);
	}
</style>
