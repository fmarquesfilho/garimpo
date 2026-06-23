<script>
	import { onMount } from 'svelte';
	import { listarPublicacoes, buscarConversoes } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';

	let publicacoes = $state([]);
	let conversoes = $state([]);
	let carregando = $state(true);
	let erro = $state(null);
	let filtro = $state(''); // '' | 'agendada' | 'enviada' | 'erro'
	let aba = $state('historico'); // 'historico' | 'desempenho'

	onMount(carregar);

	async function carregar() {
		carregando = true;
		erro = null;
		try {
			const [rp, rc] = await Promise.all([
				listarPublicacoes({ status: filtro }),
				buscarConversoes({ dias: 30 }).catch(() => ({ conversoes: [] }))
			]);
			publicacoes = rp?.publicacoes ?? [];
			conversoes = rc?.conversoes ?? [];
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}
	}

	$effect(() => {
		filtro;
		carregar();
	});

	const statusIcon = { agendada: '⏱', enviada: '✓', erro: '✕' };
	const statusClass = { agendada: 'agendada', enviada: 'enviada', erro: 'erro' };
</script>

<svelte:head>
	<title>Publicações — Garimpo</title>
</svelte:head>

<section class="publicacoes-page">
	<h1>📤 Publicações</h1>
	<p class="subtitulo">
		Histórico e desempenho das publicações.
	</p>

	{#if !$usuario}
		<div class="aviso">Faça login para ver publicações.</div>
	{:else}
		<!-- Abas principais -->
		<nav class="abas-pub">
			<button class:ativa={aba === 'historico'} onclick={() => (aba = 'historico')}>Histórico</button>
			<button class:ativa={aba === 'desempenho'} onclick={() => (aba = 'desempenho')}>
				Desempenho {#if conversoes.length > 0}<span class="badge-n">{conversoes.length}</span>{/if}
			</button>
		</nav>

		{#if aba === 'historico'}
			<!-- Filtros de status -->
			<nav class="filtros-pub">
			<button class:ativa={filtro === ''} onclick={() => (filtro = '')}>Todas</button>
			<button class:ativa={filtro === 'agendada'} onclick={() => (filtro = 'agendada')}>⏱ Agendadas</button>
			<button class:ativa={filtro === 'enviada'} onclick={() => (filtro = 'enviada')}>✓ Enviadas</button>
			<button class:ativa={filtro === 'erro'} onclick={() => (filtro = 'erro')}>✕ Erros</button>
		</nav>

		{#if erro}
			<div class="msg-erro">{erro}</div>
		{/if}

		{#if carregando}
			<p class="loading">Carregando…</p>
		{:else if publicacoes.length === 0}
			<p class="vazio">Nenhuma publicação {filtro ? `com status "${filtro}"` : ''} encontrada.</p>
		{:else}
			<div class="lista">
				{#each publicacoes as p (p.id)}
					<div class="card-pub {statusClass[p.status] ?? ''}">
						<div class="pub-principal">
							<span class="status-badge">{statusIcon[p.status] ?? '?'} {p.status}</span>
							<strong class="pub-nome">{p.nome}</strong>
							<span class="pub-preco">R$ {p.preco?.toFixed(2)}</span>
						</div>
						<div class="pub-meta">
							{#if p.destino_id}
								<span>📡 {p.destino_id}</span>
							{/if}
							{#if p.template_id}
								<span>🎨 {p.template_id}</span>
							{/if}
							{#if p.agendada_em}
								<span>⏱ {p.agendada_em}</span>
							{/if}
							{#if p.enviada_em}
								<span>✓ {p.enviada_em}</span>
							{/if}
						</div>
						{#if p.detalhe && p.status === 'erro'}
							<p class="pub-detalhe erro-txt">{p.detalhe}</p>
						{:else if p.detalhe}
							<p class="pub-detalhe"><code>{p.detalhe}</code></p>
						{/if}
					</div>
				{/each}
			</div>
		{/if}

		{:else if aba === 'desempenho'}
			<!-- Relatório de desempenho das publicações por canal -->
			{#if conversoes.length === 0}
				<p class="vazio">Nenhuma publicação rastreada nos últimos 30 dias.</p>
				<p class="dica">O desempenho aparece aqui quando publicações são enviadas com sub_id de rastreamento.</p>
			{:else}
				<div class="tabela-desemp">
					<table>
						<thead>
							<tr>
								<th>Canal</th>
								<th>Produto</th>
								<th>Publicações</th>
								<th>Comissão est.</th>
								<th>Último envio</th>
							</tr>
						</thead>
						<tbody>
							{#each conversoes as c (c.sub_id)}
								<tr>
									<td><span class="canal-badge">{c.canal}</span></td>
									<td class="nome-col">{c.nome}</td>
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
	.publicacoes-page { max-width: 780px; }
	h1 { font-size: 1.5rem; margin-bottom: 0.25rem; }
	.subtitulo { color: var(--tinta-suave); font-size: 0.9rem; margin-bottom: var(--r5); }

	.abas-pub {
		display: flex; gap: 2px; margin-bottom: var(--r4);
		border-bottom: 2px solid var(--linha);
	}
	.abas-pub button {
		padding: 8px 16px; border: none; background: transparent;
		font-weight: 600; font-size: 0.88rem; color: var(--tinta-suave);
		cursor: pointer; border-bottom: 2px solid transparent; margin-bottom: -2px;
		display: flex; align-items: center; gap: 6px;
	}
	.abas-pub button.ativa { color: var(--tinta); border-bottom-color: var(--ouro); }
	.badge-n { font-size: 0.7rem; background: var(--ouro-fundo); color: #7a5a1e; padding: 1px 6px; border-radius: 999px; font-weight: 700; }

	.filtros-pub {
		display: flex; gap: 2px; margin-bottom: var(--r5);
		border-bottom: 2px solid var(--linha);
	}
	.filtros-pub button {
		padding: 8px 16px; border: none; background: transparent;
		font-weight: 600; font-size: 0.85rem; color: var(--tinta-suave);
		cursor: pointer; border-bottom: 2px solid transparent; margin-bottom: -2px;
	}
	.filtros-pub button.ativa { color: var(--tinta); border-bottom-color: var(--ouro); }

	.aviso, .msg-erro { padding: var(--r3) var(--r4); border-radius: 8px; font-size: 0.88rem; margin-bottom: var(--r4); }
	.aviso { background: var(--porcelana); color: var(--tinta-suave); }
	.msg-erro { background: #fef2f2; color: #b91c1c; border: 1px solid #fecaca; }

	.loading, .vazio { color: var(--tinta-suave); font-size: 0.9rem; }

	.lista { display: flex; flex-direction: column; gap: var(--r3); }
	.card-pub {
		padding: var(--r3) var(--r4); border: 1px solid var(--linha);
		border-radius: 10px; background: white; border-left: 3px solid var(--linha);
	}
	.card-pub.enviada { border-left-color: #22c55e; }
	.card-pub.agendada { border-left-color: var(--ouro); }
	.card-pub.erro { border-left-color: #ef4444; }

	.pub-principal { display: flex; align-items: center; gap: var(--r3); flex-wrap: wrap; }
	.status-badge {
		font-size: 0.72rem; font-weight: 700; padding: 2px 8px;
		border-radius: 999px; background: var(--porcelana);
	}
	.pub-nome { font-size: 0.92rem; flex: 1; }
	.pub-preco { font-weight: 700; color: var(--ouro); font-size: 0.88rem; }

	.pub-meta {
		display: flex; flex-wrap: wrap; gap: var(--r2); margin-top: 4px;
		font-size: 0.78rem; color: var(--tinta-suave);
	}
	.pub-detalhe { font-size: 0.78rem; margin: 4px 0 0; color: var(--tinta-suave); }
	.pub-detalhe code { font-size: 0.72rem; background: var(--porcelana); padding: 2px 6px; border-radius: 4px; }
	.erro-txt { color: #b91c1c; }
	.dica { font-size: 0.82rem; color: var(--tinta-suave); margin-top: var(--r2); }

	/* Desempenho */
	.tabela-desemp { overflow-x: auto; }
	.tabela-desemp table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
	.tabela-desemp th { text-align: left; font-weight: 600; padding: 8px 10px; border-bottom: 2px solid var(--linha); font-size: 0.78rem; text-transform: uppercase; color: var(--tinta-suave); }
	.tabela-desemp td { padding: 8px 10px; border-bottom: 1px solid var(--linha); }
	.canal-badge { font-size: 0.78rem; font-weight: 600; }
	.nome-col { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.num { text-align: right; font-weight: 600; font-variant-numeric: tabular-nums; }
	.data { font-size: 0.78rem; color: var(--tinta-suave); }
</style>
