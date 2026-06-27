<script>
	import { onMount } from 'svelte';
	import { listarPublicacoes, buscarConversoes, buscarConversoesReais } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';
	import { dataHoraCompleta } from '$lib/formatters.js';
	import { TabBar, Loading } from '$lib/components/ui/index.js';
	import PeriodSelector from '$lib/components/PeriodSelector.svelte';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';

	let publicacoes = $state([]);
	let conversoes = $state([]);
	let conversoesReais = $state(null);
	let carregando = $state(true);
	let carregandoReais = $state(false);
	let erro = $state(null);
	let erroReais = $state(null);
	let filtro = $state('');
	let aba = $state('historico');
	let diasReais = $state(30);

	const abasPrincipais = $derived([
		{ id: 'historico', label: 'Histórico' },
		{ id: 'desempenho', label: 'Desempenho', badge: conversoes.length > 0 ? String(conversoes.length) : '' }
	]);

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
			erro = { message: e.message ?? e, retry: true };
		} finally {
			carregando = false;
		}
	}

	async function carregarReais() {
		carregandoReais = true;
		erroReais = null;
		try {
			conversoesReais = await buscarConversoesReais({ dias: diasReais });
		} catch (e) {
			erroReais = { message: e.message ?? e, retry: true };
		} finally {
			carregandoReais = false;
		}
	}

	$effect(() => {
		filtro;
		carregar();
	});

	$effect(() => {
		if (aba === 'desempenho' && !conversoesReais && !carregandoReais) {
			carregarReais();
		}
	});

	const statusIcon = { agendada: '⏱', enviada: '✓', erro: '✕' };
	const statusClass = { agendada: 'agendada', enviada: 'enviada', erro: 'erro' };

	function canalDoDetalhe(detalhe) {
		if (!detalhe) return '';
		if (detalhe.startsWith('whatsapp')) return '💬 WhatsApp';
		if (detalhe.startsWith('telegram')) return '✈️ Telegram';
		return '';
	}

	function estrategiaDoDetalhe(detalhe) {
		if (!detalhe) return '';
		const parts = detalhe.split('_');
		if (parts.length >= 2) return parts[1];
		return '';
	}

	// Recarrega conversões reais quando o período muda
	let periodoInicial = true;
	$effect(() => {
		diasReais; // track
		if (periodoInicial) { periodoInicial = false; return; }
		carregarReais();
	});
</script>

<svelte:head>
	<title>Publicações — Garimpei</title>
</svelte:head>

<section class="publicacoes-page">
	<h1>📤 Publicações</h1>
	<p class="subtitulo">Histórico e desempenho das publicações.</p>

	{#if !$usuario}
		<div class="aviso">Faça login para ver publicações.</div>
	{:else}
		<TabBar tabs={abasPrincipais} bind:active={aba} />

		{#if aba === 'historico'}
			<!-- Filtros de status -->
			<nav class="filtros-pub">
				<button class:ativa={filtro === ''} onclick={() => (filtro = '')}>Todas</button>
				<button class:ativa={filtro === 'agendada'} onclick={() => (filtro = 'agendada')}>⏱ Agendadas</button>
				<button class:ativa={filtro === 'enviada'} onclick={() => (filtro = 'enviada')}>✓ Enviadas</button>
				<button class:ativa={filtro === 'erro'} onclick={() => (filtro = 'erro')}>✕ Erros</button>
			</nav>

			{#if erro}
				<ErrorMessage {erro} onretry={carregar} />
			{/if}

			{#if carregando}
				<Loading mensagem="Carregando…" />
			{:else if publicacoes.length === 0}
				<p class="vazio">Nenhuma publicação {filtro ? `com status "${filtro}"` : ''} encontrada.</p>
			{:else}
				<div class="lista">
					{#each publicacoes as p (p.id)}
						<div class="card-pub {statusClass[p.status] ?? ''}">
							<div class="pub-principal">
								<span class="status-badge">{statusIcon[p.status] ?? '?'} {p.status}</span>
								<strong class="pub-nome">{p.nome || '(sem título)'}</strong>
								{#if p.preco > 0}
									<span class="pub-preco">R$ {p.preco?.toFixed(2)}</span>
								{/if}
							</div>
							<div class="pub-meta">
								{#if p.destino_id}<span>📡 {p.destino_id}</span>{/if}
								{#if p.estrategia}<span>🎯 {p.estrategia}</span>{/if}
								{#if p.agendada_em}<span>⏱ Agendada: {dataHoraCompleta(p.agendada_em)}</span>{/if}
								{#if p.enviada_em}<span>✓ Enviada: {dataHoraCompleta(p.enviada_em)}</span>{/if}
								{#if !p.enviada_em && p.criada_em}<span>📅 Criada: {dataHoraCompleta(p.criada_em)}</span>{/if}
							</div>
							{#if p.detalhe && p.status === 'erro'}
								<p class="pub-detalhe erro-txt">{p.detalhe}</p>
							{:else if p.detalhe && p.status === 'enviada'}
								{@const canal = canalDoDetalhe(p.detalhe)}
								{#if canal}
									<p class="pub-detalhe">{canal} · {estrategiaDoDetalhe(p.detalhe)}</p>
								{/if}
							{/if}
						</div>
					{/each}
				</div>
			{/if}

		{:else if aba === 'desempenho'}
			<div class="desemp-header">
				<PeriodSelector bind:value={diasReais} options={[7, 30, 90]} />
				<button class="btn-sync" onclick={carregarReais} disabled={carregandoReais}>
					{carregandoReais ? '⏳' : '🔄'} Sincronizar
				</button>
			</div>

			{#if erroReais}
				<ErrorMessage erro={erroReais} onretry={carregarReais} />
			{:else if carregandoReais}
				<Loading mensagem="Consultando relatório de conversões da Shopee…" />
			{:else if !conversoesReais || conversoesReais.total === 0}
				<div class="info-desempenho">
					<h3>📊 Nenhuma conversão nos últimos {diasReais} dias</h3>
					<p>Quando alguém comprar pelo seu link de afiliado, a venda aparece aqui com:</p>
					<ul class="lista-info">
						<li>📦 Nome do <strong>produto</strong> vendido</li>
						<li>🏪 <strong>Loja</strong> que vendeu</li>
						<li>💰 <strong>Comissão</strong> real recebida</li>
						<li>📡 <strong>Canal</strong> da publicação (sub_id)</li>
						<li>📅 Data da <strong>compra</strong></li>
						<li>⏳ <strong>Status</strong> (pendente ou confirmada)</li>
					</ul>
					<p class="dica">💡 O sistema consulta os últimos {diasReais} dias do relatório de conversões da Shopee.</p>
				</div>
			{:else}
				<!-- Resumo -->
				<div class="resumo-conversoes">
					<div class="resumo-card destaque">
						<span class="resumo-num">R$ {conversoesReais.comissao_total?.toFixed(2)}</span>
						<span class="resumo-label">Comissão total</span>
					</div>
					<div class="resumo-card">
						<span class="resumo-num">{conversoesReais.total}</span>
						<span class="resumo-label">Conversões</span>
					</div>
					<div class="resumo-card">
						<span class="resumo-num">{conversoesReais.confirmadas}</span>
						<span class="resumo-label">Confirmadas</span>
					</div>
					<div class="resumo-card">
						<span class="resumo-num">{conversoesReais.pendentes}</span>
						<span class="resumo-label">Pendentes</span>
					</div>
				</div>

				<!-- Tabela detalhada -->
				<div class="tabela-desemp">
					<table>
						<thead>
							<tr>
								<th>Produto</th>
								<th>Loja</th>
								<th>Comissão</th>
								<th>Status</th>
								<th>Canal (sub_id)</th>
								<th>Compra em</th>
							</tr>
						</thead>
						<tbody>
							{#each conversoesReais.conversoes as c (c.conversion_id)}
								<tr>
									<td class="nome-col">{c.product_name || '—'}</td>
									<td class="loja-col">{c.shop_name || '—'}</td>
									<td class="num comissao-val">R$ {c.total_commission?.toFixed(2)}</td>
									<td>
										<span class="status-badge-conv" class:pendente={c.status === 'PENDING' || c.status === 'UNPAID'} class:confirmada={c.status === 'COMPLETED' || c.status === 'PAID'} class:cancelada={c.status === 'CANCELLED'}>
											{c.status}
										</span>
									</td>
									<td class="sub-id-col">{c.utm_content || '—'}</td>
									<td class="data">{dataHoraCompleta(c.purchase_time)}</td>
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
	.publicacoes-page { max-width: 900px; }
	h1 { font-size: 1.5rem; margin-bottom: 0.25rem; }
	.subtitulo { color: var(--tinta-suave); font-size: 0.9rem; margin-bottom: var(--r5); }

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

	.aviso { padding: var(--r3) var(--r4); border-radius: 8px; font-size: 0.88rem; background: var(--porcelana); color: var(--tinta-suave); }
	.vazio { color: var(--tinta-suave); font-size: 0.9rem; }

	.lista { display: flex; flex-direction: column; gap: var(--r3); }
	.card-pub {
		padding: var(--r3) var(--r4); border: 1px solid var(--linha);
		border-radius: var(--raio-sm); background: var(--branco); border-left: 3px solid var(--linha);
	}
	.card-pub.enviada { border-left-color: var(--sucesso-texto); }
	.card-pub.agendada { border-left-color: var(--ouro); }
	.card-pub.erro { border-left-color: var(--erro-texto); }

	.pub-principal { display: flex; align-items: center; gap: var(--r3); flex-wrap: wrap; }
	.status-badge {
		font-size: 0.72rem; font-weight: 700; padding: 2px 8px;
		border-radius: var(--raio-full); background: var(--porcelana);
	}
	.pub-nome { font-size: 0.92rem; flex: 1; }
	.pub-preco { font-weight: 700; color: var(--ouro); font-size: 0.88rem; }

	.pub-meta {
		display: flex; flex-wrap: wrap; gap: var(--r2); margin-top: 4px;
		font-size: 0.78rem; color: var(--tinta-suave);
	}
	.pub-detalhe { font-size: 0.78rem; margin: 4px 0 0; color: var(--tinta-suave); }
	.erro-txt { color: var(--erro-texto); }
	.dica { font-size: 0.82rem; color: var(--tinta-suave); margin-top: var(--r2); }

	/* Desempenho */
	.desemp-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--r4); flex-wrap: wrap; gap: var(--r3); }
	.btn-sync { padding: 6px 14px; border: 1px solid var(--ouro); background: var(--ouro-fundo); color: var(--ouro-escuro); border-radius: var(--raio-sm); font-size: 0.82rem; font-weight: 600; cursor: pointer; }
	.btn-sync:hover:not(:disabled) { background: var(--ouro-claro); }
	.btn-sync:disabled { opacity: 0.5; }

	.info-desempenho {
		background: var(--nevoa); border: 1px solid var(--linha);
		border-radius: var(--raio); padding: var(--r5);
	}
	.info-desempenho h3 { font-size: 1.1rem; margin: 0 0 var(--r3); }
	.info-desempenho p { margin: var(--r2) 0; font-size: var(--text-base); }
	.lista-info { padding-left: var(--r5); margin: var(--r3) 0; }
	.lista-info li { margin: var(--r2) 0; font-size: var(--text-base); }

	.resumo-conversoes { display: flex; gap: var(--r3); margin-bottom: var(--r5); flex-wrap: wrap; }
	.resumo-card { display: flex; flex-direction: column; align-items: center; padding: var(--r4); border: 1px solid var(--linha); border-radius: var(--raio-sm); min-width: 100px; }
	.resumo-card.destaque { background: var(--sucesso-fundo); border-color: var(--sucesso-texto); }
	.resumo-num { font-size: 1.3rem; font-weight: 700; font-family: var(--mono); }
	.resumo-label { font-size: 0.7rem; color: var(--tinta-suave); text-transform: uppercase; margin-top: 2px; }

	.tabela-desemp { overflow-x: auto; }
	.tabela-desemp table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
	.tabela-desemp th { text-align: left; font-weight: 600; padding: 8px 10px; border-bottom: 2px solid var(--linha); font-size: 0.78rem; text-transform: uppercase; color: var(--tinta-suave); }
	.tabela-desemp td { padding: 8px 10px; border-bottom: 1px solid var(--linha); }
	.comissao-val { color: var(--sucesso-texto); }
	.status-badge-conv { font-size: 0.72rem; font-weight: 700; padding: 2px 8px; border-radius: var(--raio-full); }
	.status-badge-conv.pendente { background: var(--ouro-fundo); color: var(--ouro-escuro); }
	.status-badge-conv.confirmada { background: var(--sucesso-fundo); color: var(--sucesso-texto); }
	.status-badge-conv.cancelada { background: var(--erro-fundo); color: var(--erro-texto); }
	.sub-id-col { font-size: 0.72rem; font-family: var(--mono); max-width: 150px; overflow: hidden; text-overflow: ellipsis; }
	.loja-col { font-size: 0.82rem; max-width: 120px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.nome-col { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.num { text-align: right; font-weight: 600; font-variant-numeric: tabular-nums; }
	.data { font-size: 0.78rem; color: var(--tinta-suave); }
</style>
