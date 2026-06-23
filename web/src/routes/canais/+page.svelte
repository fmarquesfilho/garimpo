<script>
	import { onMount } from 'svelte';
	import { listarDestinos, salvarDestino, deletarDestino } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';

	let destinos = $state([]);
	let carregando = $state(true);
	let erro = $state(null);
	let sucesso = $state('');

	// Form de novo destino
	let nome = $state('');
	let config = $state('');
	let tipo = $state('telegram');
	let salvando = $state(false);

	onMount(carregar);

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
	const configPlaceholder = {
		telegram: '@meucanal ou -1001234567890',
		whatsapp: '+5511999999999'
	};
</script>

<svelte:head>
	<title>Configurações — Garimpo</title>
</svelte:head>

<section class="config-page">
	<h1>⚙️ Configurações</h1>

	{#if !$usuario}
		<div class="aviso">Faça login para gerenciar configurações.</div>
	{:else}
		<h2>📡 Destinos de publicação</h2>
		<p class="subtitulo">
			Configure os grupos e canais onde o Garimpo publica. Cada destino usa o bot
			configurado para o tipo (Telegram, WhatsApp). Adicione o bot como admin do grupo.
		</p>

		{#if erro}
			<div class="erro">{erro}</div>
		{/if}
		{#if sucesso}
			<div class="sucesso">{sucesso}</div>
		{/if}

		<!-- Form -->
		<form class="form-destino" onsubmit={(e) => { e.preventDefault(); adicionar(); }}>
			<div class="campo">
				<label for="tipo">Tipo</label>
				<select id="tipo" bind:value={tipo}>
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
				<input id="config" bind:value={config} placeholder={configPlaceholder[tipo]} required />
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
					<div class="card-destino">
						<div class="info">
							<span class="tipo-badge">{tipoIcone[d.tipo] ?? '📤'} {tipoLabel[d.tipo] ?? d.tipo}</span>
							<strong>{d.nome}</strong>
							<code>{d.config}</code>
						</div>
						<button class="btn-remover" onclick={() => remover(d.id)} title="Remover">✕</button>
					</div>
				{/each}
			</div>
		{/if}
	{/if}
</section>

<style>
	.config-page { max-width: 640px; }
	h1 { font-size: 1.5rem; margin-bottom: var(--r6); }
	h2 { font-size: 1.1rem; margin: 0 0 var(--r2); }
	.subtitulo { color: var(--tinta-suave); font-size: 0.88rem; margin-bottom: var(--r5); }

	.aviso, .erro, .sucesso { padding: var(--r3) var(--r4); border-radius: 8px; font-size: 0.88rem; margin-bottom: var(--r4); }
	.aviso { background: var(--porcelana); color: var(--tinta-suave); }
	.erro { background: #fef2f2; color: #b91c1c; border: 1px solid #fecaca; }
	.sucesso { background: #f0fdf4; color: #166534; border: 1px solid #bbf7d0; }

	.form-destino {
		display: flex; flex-wrap: wrap; gap: var(--r3); align-items: flex-end;
		margin-bottom: var(--r5); padding: var(--r4);
		border: 1px solid var(--linha); border-radius: 12px; background: var(--porcelana);
	}
	.campo { flex: 1; min-width: 140px; display: flex; flex-direction: column; gap: 4px; }
	.campo label { font-size: 0.78rem; font-weight: 600; color: var(--tinta-suave); }
	.campo input, .campo select { padding: 8px 12px; border: 1px solid var(--linha); border-radius: 8px; font-size: 0.9rem; }
	.campo input:focus, .campo select:focus { outline: 2px solid var(--ouro); outline-offset: 1px; }
	.form-destino > button {
		padding: 8px 20px; background: var(--ouro); color: white;
		font-weight: 600; font-size: 0.88rem; border: none; border-radius: 8px; cursor: pointer;
	}
	.form-destino > button:disabled { opacity: 0.5; cursor: not-allowed; }

	.loading, .vazio { color: var(--tinta-suave); font-size: 0.9rem; }
	.lista { display: flex; flex-direction: column; gap: var(--r3); }
	.card-destino {
		display: flex; align-items: center; justify-content: space-between;
		padding: var(--r3) var(--r4); border: 1px solid var(--linha); border-radius: 10px; background: white;
	}
	.card-destino .info { display: flex; flex-direction: column; gap: 2px; }
	.card-destino strong { font-size: 0.92rem; }
	.card-destino code { font-size: 0.8rem; color: var(--tinta-suave); }
	.tipo-badge { font-size: 0.72rem; font-weight: 600; color: var(--tinta-suave); }
	.btn-remover {
		background: none; border: 1px solid var(--linha); border-radius: 6px;
		width: 32px; height: 32px; display: flex; align-items: center; justify-content: center;
		cursor: pointer; color: var(--tinta-suave); font-size: 1rem;
	}
	.btn-remover:hover { color: #b91c1c; border-color: #fca5a5; background: #fef2f2; }
</style>
