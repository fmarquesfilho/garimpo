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

<section class="max-w-[900px]">
	<h1 class="text-2xl mb-6">⚙️ Configurações</h1>

	{#if !$usuario}
		<div class="py-3 px-4 rounded-sm text-sm mb-4 bg-porcelana text-tinta-suave">Faça login para gerenciar configurações.</div>
	{:else}
		<h2 class="text-lg mb-2">📡 Destinos de publicação</h2>
		<p class="text-tinta-suave text-sm mb-5">
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
			class="flex flex-wrap gap-3 items-end mb-5 p-4 border border-border rounded-md bg-porcelana"
			onsubmit={(e) => {
				e.preventDefault();
				adicionar();
			}}
		>
			<div class="flex-1 min-w-[140px] flex flex-col gap-1">
				<label for="tipo" class="text-xs font-semibold text-tinta-suave">Tipo</label>
				<select id="tipo" bind:value={tipo} onchange={aoMudarTipo}>
					<option value="telegram">✈️ Telegram</option>
					<option value="whatsapp">💬 WhatsApp</option>
				</select>
			</div>
			<div class="flex-1 min-w-[140px] flex flex-col gap-1">
				<label for="nome" class="text-xs font-semibold text-tinta-suave">Nome</label>
				<input id="nome" class="px-3 py-2 border border-border rounded-lg text-sm focus:outline-2 focus:outline-ouro focus:outline-offset-1" bind:value={nome} placeholder="ex.: Ofertas Beleza" required />
			</div>
			<div class="flex-1 min-w-[140px] flex flex-col gap-1">
				<label for="config" class="text-xs font-semibold text-tinta-suave">Destino ({tipoLabel[tipo]})</label>
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
					<input id="config" class="px-3 py-2 border border-border rounded-lg text-sm focus:outline-2 focus:outline-ouro focus:outline-offset-1" bind:value={config} placeholder="@meucanal ou -1001234567890" required />
				{/if}
			</div>
			<Button type="submit" disabled={salvando || !nome.trim() || !config.trim()}>
				{salvando ? 'Salvando…' : '+ Adicionar'}
			</Button>
		</form>

		<!-- Lista -->
		{#if carregando}
			<p class="text-tinta-suave text-sm">Carregando…</p>
		{:else if destinos.length === 0}
			<p class="text-tinta-suave text-sm">Nenhum destino cadastrado. Adicione o primeiro acima.</p>
		{:else}
			<div class="flex flex-col gap-3">
				{#each destinos as d (d.id)}
					{#if editandoId === d.id}
						<!-- Modo edição -->
						<div class="flex items-center justify-between p-3 px-4 border border-ouro rounded-sm bg-ouro-fundo">
							<div class="w-full flex flex-col gap-3">
								<div class="flex flex-col gap-1">
									<label class="text-xs font-semibold text-tinta-suave"
										>Nome
										<input class="px-3 py-2 border border-border rounded-lg text-sm focus:outline-2 focus:outline-ouro focus:outline-offset-1" bind:value={editNome} placeholder="Nome do destino" />
									</label>
								</div>
								<div class="flex flex-col gap-1">
									<span class="text-xs font-semibold text-tinta-suave">Grupos</span>
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
										<input class="px-3 py-2 border border-border rounded-lg text-sm focus:outline-2 focus:outline-ouro focus:outline-offset-1" bind:value={editConfig} placeholder="@canal ou chat_id" />
									{/if}
								</div>
								<div class="flex gap-2">
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
						<div class="flex items-center justify-between p-3 px-4 border border-border rounded-sm bg-[var(--branco)]">
							<div class="flex flex-col gap-0.5">
								<span class="text-xs font-semibold text-tinta-suave">{tipoIcone[d.tipo] ?? '📤'} {tipoLabel[d.tipo] ?? d.tipo}</span>
								<strong class="text-sm">{d.nome}</strong>
								{#if d.tipo === 'whatsapp' && gruposWA.length > 0}
									<div class="flex flex-col gap-0.5 mt-0.5">
										{#each d.config.split(',') as gid (gid)}
											{@const grupo = gruposWA.find((g) => g.id === gid.trim())}
											<span class="text-xs text-tinta-suave px-1.5 bg-sucesso-fundo rounded border border-sucesso-borda">{grupo?.nome ?? gid.trim()}</span>
										{/each}
									</div>
								{:else}
									<code class="text-xs text-tinta-suave">{d.config}</code>
								{/if}
							</div>
							<div class="flex gap-1">
								<DropdownMenu
									items={[
										{ label: '✎ Editar', onclick: () => iniciarEdicao(d) },
										{ label: '✕ Remover', onclick: () => pedirRemocao(d), destructive: true }
									]}
								>
									<button class="bg-transparent border border-border rounded-sm w-8 h-8 flex items-center justify-center cursor-pointer text-tinta-suave text-lg hover:border-ouro hover:text-ouro" aria-label="Ações">⋮</button>
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
			<div class="flex gap-3 mt-4">
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
