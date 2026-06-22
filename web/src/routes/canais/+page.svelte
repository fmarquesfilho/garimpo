<script>
	import { onMount } from 'svelte';
	import { listarDestinos, salvarDestino, deletarDestino, buscarConversoes } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';

	let destinos = $state([]);
	let conversoes = $state([]);
	let carregando = $state(true);
	let erro = $state(null);
	let sucesso = $state('');

	// Form de novo destino
	let nome = $state('');
	let config = $state('');
	let tipo = $state('telegram');
	let salvando = $state(false);

	// Abas
	let aba = $state('destinos'); // 'destinos' | 'conversoes'

	onMount(carregar);

	async function carregar() {
		carregando = true;
		erro = null;
		try {
			const [rd, rc] = await Promise.all([
				listarDestinos(),
				buscarConversoes({ dias: 30 }).catch(() => ({ conversoes: [] }))
			]);
			destinos = rd?.destinos ?? [];
			conversoes = rc?.conversoes ?? [];
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
	<title>Destinos — Garimpo</title>
</svelte:head>

<section class="destinos-page">
	<h1>📡 Destinos de Publicação</h1>
	<p class="subtitulo">
		Gerencie onde o Garimpo publica ofertas. Adicione grupos de Telegram, números
		de WhatsApp, etc. O mesmo bot/integração serve todos os destinos do mesmo tipo.
	</p>

	{#if !$usuario}
		<div class="aviso">Faça login para gerenciar destinos.</div>
	{:else}
		<!-- Abas -->
		<nav class="abas">
			<button class:ativa={aba === 'destinos'} onclick={() => (aba = 'destinos')}>Destinos</button>
			<button class:ativa={aba === 'conversoes'} onclick={() => (aba = 'conversoes')}>Conversões</button>
		</nav>

		{#if erro}
			<div class="erro">{erro}</div>
		{/if}
		{#if sucesso}
			<div class="sucesso">{sucesso}</div>
		{/if}

		{#if aba === 'destinos'}
			<!-- Formulário de novo destino -->
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

			<p class="dica">
				💡 Para Telegram, o bot precisa ser admin do canal/grupo.
				Para WhatsApp, a integração será via Business API (em breve).
			</p>

			<!-- Lista de destinos -->
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

		{:else if aba === 'conversoes'}
			<!-- Relatório de conversões -->
			<div class="conversoes-info">
				<p>
					Cada publicação gera um <code>sub_id</code> (canal + estratégia + data) que
					identifica de onde veio cada venda. Abaixo está o volume publicado por destino.
				</p>
			</div>

			{#if carregando}
				<p class="loading">Carregando…</p>
			{:else if conversoes.length === 0}
				<p class="vazio">Nenhuma publicação registrada nos últimos 30 dias.</p>
			{:else}
				<div class="tabela-conversoes">
					<table>
						<thead>
							<tr>
								<th>Canal</th>
								<th>Atribuição</th>
								<th>Produto</th>
								<th>Publicações</th>
								<th>Comissão est.</th>
								<th>Último</th>
							</tr>
						</thead>
						<tbody>
							{#each conversoes as c (c.sub_id)}
								<tr>
									<td><span class="tipo-badge mini">{tipoIcone[c.canal] ?? '📤'} {c.canal}</span></td>
									<td><code class="subid">{c.sub_id}</code></td>
									<td class="nome-produto">{c.nome}</td>
									<td class="num">{c.publicacoes}</td>
									<td class="num">R$ {c.comissao_estimada?.toFixed(2) ?? '—'}</td>
									<td class="data">{c.publicado_em}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		{/if}
	{/if}
</section>

<style>
	.destinos-page {
		max-width: 780px;
	}
	h1 {
		font-size: 1.5rem;
		margin-bottom: 0.25rem;
	}
	.subtitulo {
		color: var(--tinta-suave);
		font-size: 0.9rem;
		margin-bottom: var(--r6);
	}

	.abas {
		display: flex;
		gap: 2px;
		margin-bottom: var(--r5);
		border-bottom: 2px solid var(--linha);
	}
	.abas button {
		padding: 8px 20px;
		border: none;
		background: transparent;
		font-weight: 600;
		font-size: 0.88rem;
		color: var(--tinta-suave);
		cursor: pointer;
		border-bottom: 2px solid transparent;
		margin-bottom: -2px;
	}
	.abas button.ativa {
		color: var(--tinta);
		border-bottom-color: var(--ouro);
	}

	.aviso, .erro, .sucesso {
		padding: var(--r3) var(--r4);
		border-radius: 8px;
		font-size: 0.88rem;
		margin-bottom: var(--r4);
	}
	.aviso { background: var(--porcelana); color: var(--tinta-suave); }
	.erro { background: #fef2f2; color: #b91c1c; border: 1px solid #fecaca; }
	.sucesso { background: #f0fdf4; color: #166534; border: 1px solid #bbf7d0; }

	.form-destino {
		display: flex;
		flex-wrap: wrap;
		gap: var(--r3);
		align-items: flex-end;
		margin-bottom: var(--r4);
		padding: var(--r4);
		border: 1px solid var(--linha);
		border-radius: 12px;
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
	.campo input, .campo select {
		padding: 8px 12px;
		border: 1px solid var(--linha);
		border-radius: 8px;
		font-size: 0.9rem;
	}
	.campo input:focus, .campo select:focus {
		outline: 2px solid var(--ouro);
		outline-offset: 1px;
	}
	.form-destino > button {
		padding: 8px 20px;
		background: var(--ouro);
		color: white;
		font-weight: 600;
		font-size: 0.88rem;
		border: none;
		border-radius: 8px;
		cursor: pointer;
		white-space: nowrap;
	}
	.form-destino > button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.dica {
		font-size: 0.82rem;
		color: var(--tinta-suave);
		margin-bottom: var(--r6);
	}
	.dica code {
		background: var(--porcelana);
		padding: 2px 6px;
		border-radius: 4px;
		font-size: 0.8rem;
	}

	.loading, .vazio {
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
		border-radius: 10px;
		background: white;
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
	.tipo-badge.mini {
		font-size: 0.78rem;
	}
	.btn-remover {
		background: none;
		border: 1px solid var(--linha);
		border-radius: 6px;
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		color: var(--tinta-suave);
		font-size: 1rem;
	}
	.btn-remover:hover {
		color: #b91c1c;
		border-color: #fca5a5;
		background: #fef2f2;
	}

	/* Conversões */
	.conversoes-info {
		margin-bottom: var(--r4);
		font-size: 0.88rem;
		color: var(--tinta-suave);
	}
	.conversoes-info code {
		background: var(--porcelana);
		padding: 2px 6px;
		border-radius: 4px;
		font-size: 0.8rem;
	}
	.tabela-conversoes {
		overflow-x: auto;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.85rem;
	}
	th {
		text-align: left;
		font-weight: 600;
		padding: 8px 10px;
		border-bottom: 2px solid var(--linha);
		color: var(--tinta-suave);
		font-size: 0.78rem;
		text-transform: uppercase;
		letter-spacing: 0.03em;
	}
	td {
		padding: 8px 10px;
		border-bottom: 1px solid var(--linha);
	}
	.subid {
		font-size: 0.72rem;
		color: var(--tinta-suave);
	}
	.nome-produto {
		max-width: 180px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.num {
		text-align: right;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
	}
	.data {
		font-size: 0.8rem;
		color: var(--tinta-suave);
	}
</style>
