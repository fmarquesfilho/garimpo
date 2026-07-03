<script>
	import { onMount } from 'svelte';
	import { listarPublicacoes, buscarConversoes, buscarConversoesReais } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';
	import { dataHoraCompleta } from '$lib/formatters.js';
	import { Tabs, Loading, Button } from '$lib/components/ui/index.js';
	import PeriodSelector from '$lib/components/PeriodSelector.svelte';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';

	let publicacoes = $state([]);
	let conversoes = $state([]);
	let conversoesReais = $state(null);
	let carregando = $state(true);
	let carregandoReais = $state(false);
	let tentouReais = $state(false);
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
		tentouReais = true;
		erroReais = null;

		let timeoutId;
		const timeout = new Promise((_, reject) => {
			timeoutId = setTimeout(() => reject(new Error('A Shopee não respondeu a tempo. Tente novamente.')), 20000);
		});

		try {
			conversoesReais = await Promise.race([buscarConversoesReais({ dias: diasReais }), timeout]);
		} catch (e) {
			erroReais = { message: e.message ?? e, retry: true };
		} finally {
			clearTimeout(timeoutId);
			carregandoReais = false;
		}
	}

	$effect(() => {
		filtro;
		carregar();
	});

	$effect(() => {
		if (aba === 'desempenho' && !conversoesReais && !carregandoReais && !tentouReais) {
			carregarReais();
		}
	});

	const statusIcon = { agendada: '⏱', enviada: '✓', erro: '✕' };

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
		if (periodoInicial) {
			periodoInicial = false;
			return;
		}
		tentouReais = false;
		carregarReais();
	});
</script>

<svelte:head>
	<title>Publicações — Garimpei</title>
</svelte:head>

<section class="max-w-[900px]">
	<h1 class="text-2xl mb-1">📤 Publicações</h1>
	<p class="text-tinta-suave text-sm mb-5">Histórico e desempenho das publicações.</p>

	{#if !$usuario}
		<div class="py-3 px-4 rounded-lg text-sm bg-porcelana text-tinta-suave">Faça login para ver publicações.</div>
	{:else}
		<Tabs tabs={abasPrincipais} bind:active={aba}>
			{#if aba === 'historico'}
				<!-- Filtros de status -->
				<nav class="flex gap-0.5 mb-5 border-b-2 border-border">
					<button
						class="py-2 px-4 border-none bg-transparent font-semibold text-sm text-tinta-suave cursor-pointer border-b-2 border-b-transparent -mb-0.5 {filtro === '' ? '!text-tinta !border-b-ouro' : ''}"
						onclick={() => (filtro = '')}
					>Todas</button>
					<button
						class="py-2 px-4 border-none bg-transparent font-semibold text-sm text-tinta-suave cursor-pointer border-b-2 border-b-transparent -mb-0.5 {filtro === 'agendada' ? '!text-tinta !border-b-ouro' : ''}"
						onclick={() => (filtro = 'agendada')}
					>⏱ Agendadas</button>
					<button
						class="py-2 px-4 border-none bg-transparent font-semibold text-sm text-tinta-suave cursor-pointer border-b-2 border-b-transparent -mb-0.5 {filtro === 'enviada' ? '!text-tinta !border-b-ouro' : ''}"
						onclick={() => (filtro = 'enviada')}
					>✓ Enviadas</button>
					<button
						class="py-2 px-4 border-none bg-transparent font-semibold text-sm text-tinta-suave cursor-pointer border-b-2 border-b-transparent -mb-0.5 {filtro === 'erro' ? '!text-tinta !border-b-ouro' : ''}"
						onclick={() => (filtro = 'erro')}
					>✕ Erros</button>
				</nav>

				{#if erro}
					<ErrorMessage {erro} onretry={carregar} />
				{/if}

				{#if carregando}
					<Loading mensagem="Carregando…" />
				{:else if publicacoes.length === 0}
					<p class="text-tinta-suave text-sm">Nenhuma publicação {filtro ? `com status "${filtro}"` : ''} encontrada.</p>
				{:else}
					<div class="flex flex-col gap-3">
						{#each publicacoes as p (p.id)}
							<div class="py-3 px-4 border border-border rounded-sm bg-[var(--branco)] border-l-[3px] {p.status === 'enviada' ? 'border-l-sucesso' : p.status === 'agendada' ? 'border-l-ouro' : p.status === 'erro' ? 'border-l-erro' : 'border-l-border'}">
								<div class="flex items-center gap-3 flex-wrap">
									<span class="text-xs font-bold py-0.5 px-2 rounded-full bg-porcelana">{statusIcon[p.status] ?? '?'} {p.status}</span>
									<strong class="text-sm flex-1">{p.nome || '(sem título)'}</strong>
									{#if p.preco > 0}
										<span class="font-bold text-ouro text-sm">R$ {p.preco?.toFixed(2)}</span>
									{/if}
								</div>
								<div class="flex flex-wrap gap-2 mt-1 text-xs text-tinta-suave">
									{#if p.destino_id}<span>📡 {p.destino_id}</span>{/if}
									{#if p.estrategia}<span>🎯 {p.estrategia}</span>{/if}
									{#if p.agendada_em}<span>⏱ Agendada: {dataHoraCompleta(p.agendada_em)}</span>{/if}
									{#if p.enviada_em}<span>✓ Enviada: {dataHoraCompleta(p.enviada_em)}</span>{/if}
									{#if !p.enviada_em && p.criada_em}<span>📅 Criada: {dataHoraCompleta(p.criada_em)}</span>{/if}
								</div>
								{#if p.detalhe && p.status === 'erro'}
									<p class="text-xs mt-1 text-erro">{p.detalhe}</p>
								{:else if p.detalhe && p.status === 'enviada'}
									{@const canal = canalDoDetalhe(p.detalhe)}
									{#if canal}
										<p class="text-xs mt-1 text-tinta-suave">{canal} · {estrategiaDoDetalhe(p.detalhe)}</p>
									{/if}
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			{:else if aba === 'desempenho'}
				<div class="flex items-center justify-between mb-4 flex-wrap gap-3">
					<PeriodSelector bind:value={diasReais} options={[7, 30, 90]} />
					<Button
						variant="secondary"
						size="sm"
						onclick={() => {
							tentouReais = false;
							carregarReais();
						}}
						disabled={carregandoReais}
					>
						{carregandoReais ? '⏳' : '🔄'} Sincronizar
					</Button>
				</div>

				{#if erroReais}
					<ErrorMessage erro={erroReais} onretry={carregarReais} />
				{:else if carregandoReais}
					<div class="my-4">
						<Loading mensagem="Consultando relatório de conversões da Shopee…" />
						<p class="text-tinta-suave text-sm mt-1">Isso pode levar até 15 segundos. Se não responder, tente "Sincronizar" novamente.</p>
					</div>
				{:else if !conversoesReais || conversoesReais.total === 0}
					<div class="bg-nevoa border border-border rounded-md p-5">
						<h3 class="text-lg mb-3">📊 Nenhuma conversão nos últimos {diasReais} dias</h3>
						<p class="my-2 text-base">Quando alguém comprar pelo seu link de afiliado, a venda aparece aqui com:</p>
						<ul class="pl-5 my-3">
							<li class="my-2 text-base">📦 Nome do <strong>produto</strong> vendido</li>
							<li class="my-2 text-base">🏪 <strong>Loja</strong> que vendeu</li>
							<li class="my-2 text-base">💰 <strong>Comissão</strong> real recebida</li>
							<li class="my-2 text-base">📡 <strong>Canal</strong> da publicação (sub_id)</li>
							<li class="my-2 text-base">📅 Data da <strong>compra</strong></li>
							<li class="my-2 text-base">⏳ <strong>Status</strong> (pendente ou confirmada)</li>
						</ul>
						<p class="text-sm text-tinta-suave mt-2">💡 O sistema consulta os últimos {diasReais} dias do relatório de conversões da Shopee.</p>
					</div>
				{:else}
					<!-- Resumo -->
					<div class="flex gap-3 mb-5 flex-wrap">
						<div class="flex flex-col items-center p-4 border border-sucesso-borda rounded-sm min-w-[100px] bg-sucesso-fundo">
							<span class="text-xl font-bold font-mono">{conversoesReais.comissao_total?.toFixed(2) ? `R$ ${conversoesReais.comissao_total.toFixed(2)}` : 'R$ 0.00'}</span>
							<span class="text-[0.7rem] text-tinta-suave uppercase mt-0.5">Comissão total</span>
						</div>
						<div class="flex flex-col items-center p-4 border border-border rounded-sm min-w-[100px]">
							<span class="text-xl font-bold font-mono">{conversoesReais.total}</span>
							<span class="text-[0.7rem] text-tinta-suave uppercase mt-0.5">Conversões</span>
						</div>
						<div class="flex flex-col items-center p-4 border border-border rounded-sm min-w-[100px]">
							<span class="text-xl font-bold font-mono">{conversoesReais.confirmadas}</span>
							<span class="text-[0.7rem] text-tinta-suave uppercase mt-0.5">Confirmadas</span>
						</div>
						<div class="flex flex-col items-center p-4 border border-border rounded-sm min-w-[100px]">
							<span class="text-xl font-bold font-mono">{conversoesReais.pendentes}</span>
							<span class="text-[0.7rem] text-tinta-suave uppercase mt-0.5">Pendentes</span>
						</div>
					</div>

					<!-- Tabela detalhada -->
					<div class="overflow-x-auto">
						<table class="w-full border-collapse text-sm">
							<thead>
								<tr>
									<th class="text-left font-semibold py-2 px-2.5 border-b-2 border-border text-xs uppercase text-tinta-suave">Produto</th>
									<th class="text-left font-semibold py-2 px-2.5 border-b-2 border-border text-xs uppercase text-tinta-suave">Loja</th>
									<th class="text-left font-semibold py-2 px-2.5 border-b-2 border-border text-xs uppercase text-tinta-suave">Comissão</th>
									<th class="text-left font-semibold py-2 px-2.5 border-b-2 border-border text-xs uppercase text-tinta-suave">Status</th>
									<th class="text-left font-semibold py-2 px-2.5 border-b-2 border-border text-xs uppercase text-tinta-suave">Canal (sub_id)</th>
									<th class="text-left font-semibold py-2 px-2.5 border-b-2 border-border text-xs uppercase text-tinta-suave">Compra em</th>
								</tr>
							</thead>
							<tbody>
								{#each conversoesReais.conversoes as c (c.conversion_id)}
									<tr>
										<td class="py-2 px-2.5 border-b border-border max-w-[200px] overflow-hidden text-ellipsis whitespace-nowrap">{c.product_name || '—'}</td>
										<td class="py-2 px-2.5 border-b border-border text-sm max-w-[120px] overflow-hidden text-ellipsis whitespace-nowrap">{c.shop_name || '—'}</td>
										<td class="py-2 px-2.5 border-b border-border text-right font-semibold tabular-nums text-sucesso">R$ {c.total_commission?.toFixed(2)}</td>
										<td class="py-2 px-2.5 border-b border-border">
											<span
												class="text-xs font-bold py-0.5 px-2 rounded-full {c.status === 'PENDING' || c.status === 'UNPAID' ? 'bg-ouro-fundo text-ouro-escuro' : c.status === 'COMPLETED' || c.status === 'PAID' ? 'bg-sucesso-fundo text-sucesso' : c.status === 'CANCELLED' ? 'bg-erro-fundo text-erro' : ''}"
											>
												{c.status}
											</span>
										</td>
										<td class="py-2 px-2.5 border-b border-border text-xs font-mono max-w-[150px] overflow-hidden text-ellipsis">{c.utm_content || '—'}</td>
										<td class="py-2 px-2.5 border-b border-border text-xs text-tinta-suave">{dataHoraCompleta(c.purchase_time)}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			{/if}
		</Tabs>
	{/if}
</section>
