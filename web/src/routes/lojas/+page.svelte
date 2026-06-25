<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { buscarCandidatos, buscarNovidades, adicionarLoja, removerLoja, buscarAlertasConfig, testarAlertas, configurarAlertas } from '$lib/api.js';
	import { buscasSalvas } from '$lib/buscas.js';
	import { usuario } from '$lib/firebase.js';
	import { brl, pct } from '$lib/formatters.js';

	let buscasComLojas = $derived(($buscasSalvas ?? []).filter(b => b.shop_ids?.length > 0));
	let buscaSelecionada = $state(null);
	let aba = $state('produtos'); // 'produtos' | 'novidades' | 'precos'

	// Formulário de adicionar loja
	let inputLoja = $state('');
	let adicionando = $state(false);
	let erroAdicionar = $state(null);
	let sucessoAdicionar = $state(null);

	// Alertas
	let alertasConfig = $state(null);
	let mostraAlertas = $state(false);
	let alertaChatId = $state('');
	let alertaThreshold = $state(15);
	let alertaApenasQuedas = $state(true);
	let salvandoAlerta = $state(false);
	let testandoAlerta = $state(false);
	let msgAlerta = $state(null);

	// Produtos da loja
	let produtos = $state([]);
	let carregandoProdutos = $state(false);
	let erroProdutos = $state(null);

	// Novidades
	let novidades = $state(null);
	let carregandoNovidades = $state(false);
	let erroNovidades = $state(null);

	onMount(async () => {
		buscasSalvas.sincronizarDoServidor();
		try {
			alertasConfig = await buscarAlertasConfig();
			if (alertasConfig) {
				alertaThreshold = Math.round((alertasConfig.threshold ?? 0.15) * 100);
				alertaApenasQuedas = alertasConfig.apenas_quedas ?? true;
			}
		} catch { /* sem alertas configurados */ }
	});

	async function handleSalvarAlertas() {
		salvandoAlerta = true;
		msgAlerta = null;
		try {
			await configurarAlertas({
				chatId: alertaChatId || undefined,
				threshold: alertaThreshold / 100,
				apenasQuedas: alertaApenasQuedas
			});
			alertasConfig = await buscarAlertasConfig();
			msgAlerta = { tipo: 'sucesso', texto: 'Configuração salva!' };
		} catch (e) {
			msgAlerta = { tipo: 'erro', texto: e.message };
		} finally {
			salvandoAlerta = false;
		}
	}

	async function handleTestarAlertas() {
		testandoAlerta = true;
		msgAlerta = null;
		try {
			const r = await testarAlertas({ buscaId: buscaSelecionada?.id });
			msgAlerta = { tipo: 'sucesso', texto: r.status };
		} catch (e) {
			msgAlerta = { tipo: 'erro', texto: e.message };
		} finally {
			testandoAlerta = false;
		}
	}

	async function handleAdicionarLoja() {
		const valor = inputLoja.trim();
		if (!valor) return;

		adicionando = true;
		erroAdicionar = null;
		sucessoAdicionar = null;

		try {
			const r = await adicionarLoja({ input: valor });
			sucessoAdicionar = `Loja ${r.shop_id} adicionada com sucesso!`;
			inputLoja = '';
			// Recarrega a lista de buscas do servidor
			await buscasSalvas.sincronizarDoServidor();
			// Seleciona a nova loja automaticamente
			setTimeout(() => {
				const nova = buscasComLojas.find(b => b.id === r.id);
				if (nova) selecionarBusca(nova);
				sucessoAdicionar = null;
			}, 2000);
		} catch (e) {
			erroAdicionar = e.message;
		} finally {
			adicionando = false;
		}
	}

	async function handleRemoverLoja(b) {
		if (!confirm(`Remover monitoramento da loja "${b.id}"?`)) return;
		try {
			await removerLoja(b.id);
			await buscasSalvas.sincronizarDoServidor();
			if (buscaSelecionada?.id === b.id) {
				buscaSelecionada = null;
			}
		} catch (e) {
			alert('Erro ao remover: ' + e.message);
		}
	}

	async function selecionarBusca(b) {
		buscaSelecionada = b;
		aba = 'produtos';
		await carregarProdutos();
		carregarNovidades();
	}

	async function carregarProdutos() {
		if (!buscaSelecionada) return;
		carregandoProdutos = true;
		erroProdutos = null;
		try {
			const r = await buscarCandidatos({
				fonte: 'shopee-shop',
				shopIds: buscaSelecionada.shop_ids,
				keyword: buscaSelecionada.keywords?.[0] ?? '',
				categoria: buscaSelecionada.categoria,
				estrategia: buscaSelecionada.estrategia ?? 'nicho',
				top: 50,
				semFiltro: true // monitoramento: mostra tudo, sem elegibilidade
			});
			produtos = r?.candidatos ?? [];
		} catch (e) {
			erroProdutos = e.message;
		} finally {
			carregandoProdutos = false;
		}
	}

	async function carregarNovidades() {
		if (!buscaSelecionada) return;
		carregandoNovidades = true;
		erroNovidades = null;
		try {
			const r = await buscarNovidades({ buscaId: buscaSelecionada.id, dias: 7 });
			novidades = r;
		} catch (e) {
			erroNovidades = e.message;
			novidades = null;
		} finally {
			carregandoNovidades = false;
		}
	}

	function irParaPublicar(c) {
		const dados = encodeURIComponent(JSON.stringify(c));
		goto(`/publicar?dados=${dados}`);
	}
</script>

<svelte:head>
	<title>Lojas — Garimpei</title>
</svelte:head>

<section class="lojas-page">
	<h1>🏪 Lojas Monitoradas</h1>
	<p class="subtitulo">
		Acompanhe os produtos das lojas que você monitora. Veja novidades, variações de preço
		e publique ofertas diretamente.
	</p>

	{#if !$usuario}
		<div class="aviso">Faça login para ver as lojas monitoradas.</div>
	{:else}
		<!-- Formulário para adicionar loja -->
		<div class="form-loja">
			<h2>Adicionar loja</h2>
			<form onsubmit={(e) => { e.preventDefault(); handleAdicionarLoja(); }}>
				<div class="form-row">
					<input
						type="text"
						bind:value={inputLoja}
						placeholder="Cole a URL da loja (shopee.com.br/shop/123) ou ID numérico"
						disabled={adicionando}
						class="input-loja"
					/>
					<button type="submit" disabled={adicionando || !inputLoja.trim()} class="btn-adicionar">
						{adicionando ? '⏳' : '➕'} Adicionar
					</button>
				</div>
				{#if erroAdicionar}
					<p class="msg-erro-inline">{erroAdicionar}</p>
				{/if}
				{#if sucessoAdicionar}
					<p class="msg-sucesso">{sucessoAdicionar}</p>
				{/if}
			</form>
		</div>

		<!-- Configuração de alertas -->
		<div class="painel-alertas">
			<button class="btn-toggle-alertas" onclick={() => mostraAlertas = !mostraAlertas}>
				🔔 Alertas Telegram
				{#if alertasConfig?.ativo}
					<span class="badge-ativo">Ativo</span>
				{:else}
					<span class="badge-inativo">Inativo</span>
				{/if}
			</button>

			{#if mostraAlertas}
				<div class="alertas-config">
					<div class="campo-alerta">
						<label for="chat-id">Chat ID do grupo Telegram</label>
						<input id="chat-id" type="text" bind:value={alertaChatId}
							placeholder={alertasConfig?.chat_id || 'Ex: -1001234567890'} />
						<span class="hint">ID do grupo onde os alertas serão enviados. Use @BotFather para criar o bot.</span>
					</div>
					<div class="campo-alerta">
						<label for="threshold">Threshold de variação (%)</label>
						<input id="threshold" type="number" bind:value={alertaThreshold} min="5" max="50" step="5" />
						<span class="hint">Alerta se preço variar mais que {alertaThreshold}%.</span>
					</div>
					<div class="campo-alerta checkbox">
						<label>
							<input type="checkbox" bind:checked={alertaApenasQuedas} />
							Alertar apenas quedas de preço (oportunidades)
						</label>
					</div>
					<div class="alertas-acoes">
						<button onclick={handleSalvarAlertas} disabled={salvandoAlerta} class="btn-salvar">
							{salvandoAlerta ? '⏳' : '💾'} Salvar
						</button>
						<button onclick={handleTestarAlertas} disabled={testandoAlerta || !alertasConfig?.ativo} class="btn-testar">
							{testandoAlerta ? '⏳' : '📨'} Testar
						</button>
					</div>
					{#if msgAlerta}
						<p class={msgAlerta.tipo === 'erro' ? 'msg-erro-inline' : 'msg-sucesso'}>
							{msgAlerta.texto}
						</p>
					{/if}
				</div>
			{/if}
		</div>

		{#if buscasComLojas.length === 0}
			<div class="vazio">
				<p>Nenhuma loja monitorada ainda.</p>
				<p class="dica">Use o formulário acima para adicionar uma loja Shopee.</p>
			</div>
		{:else}
			<!-- Lista de buscas com lojas -->
			<div class="lojas-lista">
				{#each buscasComLojas as b (b.id)}
					<div class="loja-card-wrapper">
						<button
							class="loja-card"
							class:ativa={buscaSelecionada?.id === b.id}
							onclick={() => selecionarBusca(b)}
						>
							<strong>{b.id}</strong>
							<span class="loja-meta">
								🏪 {b.shop_ids.length} {b.shop_ids.length === 1 ? 'loja' : 'lojas'}
								{#if b.keywords?.length > 0}
									· 🔑 {b.keywords.join(', ')}
								{/if}
							</span>
						</button>
						<button
							class="btn-remover"
							onclick={() => handleRemoverLoja(b)}
							title="Remover monitoramento"
						>✕</button>
					</div>
				{/each}
			</div>

			{#if buscaSelecionada}
				<!-- Abas -->
				<nav class="abas">
					<button class:ativa={aba === 'produtos'} onclick={() => (aba = 'produtos')}>
						Produtos {#if produtos.length > 0}<span class="badge-n">{produtos.length}</span>{/if}
					</button>
					<button class:ativa={aba === 'novidades'} onclick={() => (aba = 'novidades')}>
						🆕 Novidades {#if novidades?.produtos_novos?.length}<span class="badge-n alerta">{novidades.produtos_novos.length}</span>{/if}
					</button>
					<button class:ativa={aba === 'precos'} onclick={() => (aba = 'precos')}>
						📉 Preços {#if novidades?.variacoes?.length}<span class="badge-n">{novidades.variacoes.length}</span>{/if}
					</button>
				</nav>

				{#if aba === 'produtos'}
					{#if carregandoProdutos}
						<p class="loading">Buscando produtos da loja…</p>
					{:else if erroProdutos}
						<div class="msg-erro">{erroProdutos}</div>
					{:else if produtos.length === 0}
						<p class="vazio-tab">Nenhum produto encontrado. A coleta periódica pode ainda não ter rodado.</p>
					{:else}
						<div class="grade-produtos">
							{#each produtos as p, i (p.id)}
								<div class="card-produto-loja">
									{#if p.imagem}
										<img src={p.imagem} alt={p.nome} class="prod-thumb" />
									{/if}
									<div class="prod-info">
										<h4>{p.nome}</h4>
										<div class="prod-dados">
											<span class="prod-preco">{brl(p.preco)}</span>
											<span class="prod-comissao">{pct(p.comissao)}</span>
											<span class="prod-vendas">{p.vendas} vendas</span>
											<span class="prod-nota">★ {p.avaliacao?.toFixed(1)}</span>
										</div>
										<div class="prod-score">teor: {p.score?.toFixed(3)}</div>
									</div>
									<button class="btn-pub-mini" onclick={() => irParaPublicar(p)} title="Publicar este produto">
										📤
									</button>
								</div>
							{/each}
						</div>
					{/if}

				{:else if aba === 'novidades'}
					{#if carregandoNovidades}
						<p class="loading">Analisando novidades…</p>
					{:else if erroNovidades}
						<div class="msg-erro">{erroNovidades}</div>
					{:else if !novidades || novidades.produtos_novos?.length === 0}
						<p class="vazio-tab">Nenhum produto novo detectado nos últimos {novidades?.dias_janela ?? 7} dias.</p>
					{:else}
						<p class="info-novidades">
							<strong>{novidades.produtos_novos.length}</strong> produto(s) novo(s) detectado(s)
							nos últimos {novidades.dias_janela} dias.
						</p>
						<div class="grade-novidades">
							{#each novidades.produtos_novos as p (p.produto_id)}
								<div class="card-novidade">
									<div class="novidade-badge">🆕</div>
									<div class="novidade-info">
										<strong>{p.nome}</strong>
										<div class="novidade-dados">
											<span>{brl(p.preco)}</span>
											<span>{pct(p.comissao)} comissão</span>
											<span>{p.vendas} vendas</span>
										</div>
										<span class="novidade-data">Detectado: {p.detectado_em?.split('T')[0]}</span>
									</div>
								</div>
							{/each}
						</div>
					{/if}

				{:else if aba === 'precos'}
					{#if carregandoNovidades}
						<p class="loading">Analisando variações…</p>
					{:else if !novidades || novidades.variacoes?.length === 0}
						<p class="vazio-tab">Nenhuma variação de preço detectada nos últimos {novidades?.dias_janela ?? 7} dias.</p>
					{:else}
						<p class="info-novidades">
							<strong>{novidades.variacoes.length}</strong> variação(ões) de preço detectada(s).
						</p>
						<div class="tabela-variacoes">
							<table>
								<thead>
									<tr>
										<th>Produto</th>
										<th>Antes</th>
										<th>Agora</th>
										<th>Variação</th>
										<th>Detectado</th>
										<th></th>
									</tr>
								</thead>
								<tbody>
									{#each novidades.variacoes as v (v.produto_id)}
										<tr class:baixou={v.variacao_pct < 0} class:subiu={v.variacao_pct > 0}>
											<td class="nome-col">{v.nome}</td>
											<td>{brl(v.preco_anterior)}</td>
											<td class="preco-atual">{brl(v.preco_atual)}</td>
											<td class="variacao">
												<span class="badge-variacao" class:badge-baixou={v.variacao_pct < 0} class:badge-subiu={v.variacao_pct > 0}>
													{v.variacao_pct < 0 ? '↓' : '↑'}
													{Math.abs(v.variacao_pct * 100).toFixed(1)}%
												</span>
											</td>
											<td class="data">{v.detectado_em?.split('T')[0]}</td>
											<td>
												<button class="btn-pub-mini" onclick={() => irParaPublicar({
													id: v.produto_id,
													nome: v.nome,
													preco: v.preco_atual
												})} title="Publicar esta oferta">📤</button>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				{/if}
			{/if}
		{/if}
	{/if}
</section>

<style>
	.lojas-page { max-width: 900px; }
	h1 { font-size: 1.5rem; margin-bottom: 0.25rem; }
	h2 { font-size: 1.1rem; margin-bottom: 0.5rem; color: var(--tinta); }
	.subtitulo { color: var(--tinta-suave); font-size: 0.9rem; margin-bottom: var(--r6); }

	.aviso, .vazio { background: var(--porcelana); padding: var(--r4); border-radius: var(--raio-sm); color: var(--tinta-suave); }
	.vazio a { color: var(--ouro); text-decoration: underline; }
	.dica { font-size: 0.85rem; margin-top: 4px; }
	.vazio-tab { color: var(--tinta-suave); font-size: 0.9rem; font-style: italic; }
	.msg-erro { background: var(--erro-fundo); color: var(--erro-texto); padding: var(--r3) var(--r4); border-radius: 8px; margin-bottom: var(--r4); }
	.msg-erro-inline { color: var(--erro-texto); font-size: 0.85rem; margin-top: 6px; }
	.msg-sucesso { color: var(--sucesso-texto); font-size: 0.85rem; margin-top: 6px; }
	.loading { color: var(--tinta-suave); font-style: italic; }

	/* Formulário de adicionar loja */
	.form-loja {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r4);
		margin-bottom: var(--r5);
	}
	.form-row {
		display: flex;
		gap: var(--r3);
		align-items: stretch;
	}
	.input-loja {
		flex: 1;
		padding: 10px 14px;
		border: 1px solid var(--linha);
		border-radius: 8px;
		font-size: 0.9rem;
		background: var(--branco);
	}
	.input-loja:focus {
		outline: none;
		border-color: var(--ouro);
		box-shadow: 0 0 0 2px var(--ouro-fundo);
	}
	.btn-adicionar {
		padding: 10px 18px;
		background: var(--ouro);
		color: white;
		border: none;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		white-space: nowrap;
	}
	.btn-adicionar:hover:not(:disabled) { opacity: 0.9; }
	.btn-adicionar:disabled { opacity: 0.5; cursor: not-allowed; }

	/* Lista de buscas com lojas */
	.lojas-lista { display: flex; flex-wrap: wrap; gap: var(--r3); margin-bottom: var(--r5); }
	.loja-card-wrapper { position: relative; }
	.loja-card {
		border: 1px solid var(--linha); background: var(--nevoa);
		border-radius: var(--raio-sm); padding: var(--r3) var(--r4);
		padding-right: 32px;
		cursor: pointer; text-align: left; display: flex; flex-direction: column; gap: 2px;
	}
	.loja-card:hover { border-color: var(--ouro); }
	.loja-card.ativa { border-color: var(--ouro); background: var(--ouro-fundo); }
	.loja-card strong { font-size: 0.9rem; }
	.loja-meta { font-size: 0.78rem; color: var(--tinta-suave); }
	.btn-remover {
		position: absolute;
		top: 4px;
		right: 4px;
		width: 22px;
		height: 22px;
		border-radius: 50%;
		border: none;
		background: transparent;
		color: var(--tinta-suave);
		font-size: 0.75rem;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.btn-remover:hover { background: var(--erro-fundo); color: var(--erro-texto); }

	/* Abas */
	.abas { display: flex; gap: 2px; margin-bottom: var(--r5); border-bottom: 2px solid var(--linha); }
	.abas button {
		padding: 8px 16px; border: none; background: transparent;
		font-weight: 600; font-size: 0.85rem; color: var(--tinta-suave);
		cursor: pointer; border-bottom: 2px solid transparent; margin-bottom: -2px;
		display: flex; align-items: center; gap: 6px;
	}
	.abas button.ativa { color: var(--tinta); border-bottom-color: var(--ouro); }
	.badge-n {
		font-size: 0.7rem; background: var(--ouro-fundo); color: var(--ouro-escuro);
		padding: 1px 6px; border-radius: var(--raio-full); font-weight: 700;
	}
	.badge-n.alerta { background: var(--erro-fundo); color: var(--erro-texto); }

	/* Grade de produtos */
	.grade-produtos { display: flex; flex-direction: column; gap: var(--r3); }
	.card-produto-loja {
		display: flex; gap: var(--r3); padding: var(--r3) var(--r4);
		border: 1px solid var(--linha); border-radius: var(--raio-sm); background: var(--branco);
		align-items: center;
	}
	.prod-thumb { width: 56px; height: 56px; border-radius: 8px; object-fit: cover; flex-shrink: 0; }
	.prod-info { flex: 1; min-width: 0; }
	.prod-info h4 { font-size: 0.9rem; margin: 0 0 4px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
	.prod-dados { display: flex; flex-wrap: wrap; gap: var(--r2); font-size: 0.78rem; color: var(--tinta-suave); }
	.prod-preco { font-weight: 700; color: var(--ouro); }
	.prod-comissao { font-weight: 600; }
	.prod-score { font-size: 0.72rem; color: var(--tinta-suave); margin-top: 2px; }
	.btn-pub-mini {
		border: 1px solid var(--linha); background: var(--porcelana);
		border-radius: 8px; width: 36px; height: 36px;
		display: flex; align-items: center; justify-content: center;
		cursor: pointer; font-size: 1rem; flex-shrink: 0;
	}
	.btn-pub-mini:hover { border-color: var(--rosa); background: color-mix(in srgb, var(--rosa) 8%, white); }

	/* Novidades */
	.info-novidades { font-size: 0.88rem; margin-bottom: var(--r4); }
	.grade-novidades { display: flex; flex-direction: column; gap: var(--r3); }
	.card-novidade {
		display: flex; gap: var(--r3); padding: var(--r3) var(--r4);
		border: 1px solid var(--sucesso-borda); border-left: 3px solid var(--sucesso-texto);
		border-radius: var(--raio-sm); background: var(--sucesso-fundo);
	}
	.novidade-badge { font-size: 1.2rem; }
	.novidade-info { flex: 1; }
	.novidade-info strong { font-size: 0.9rem; }
	.novidade-dados { display: flex; gap: var(--r3); font-size: 0.78rem; color: var(--tinta-suave); margin-top: 2px; }
	.novidade-data { font-size: 0.72rem; color: var(--tinta-suave); }

	/* Variações de preço */
	.tabela-variacoes { overflow-x: auto; }
	table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
	th { text-align: left; font-weight: 600; padding: 8px 10px; border-bottom: 2px solid var(--linha); font-size: 0.78rem; text-transform: uppercase; color: var(--tinta-suave); }
	td { padding: 8px 10px; border-bottom: 1px solid var(--linha); }
	.nome-col { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.preco-atual { font-weight: 700; }
	.variacao { font-weight: 700; }
	.badge-variacao {
		display: inline-block;
		padding: 2px 8px;
		border-radius: var(--raio-full);
		font-size: 0.78rem;
		font-weight: 700;
	}
	.badge-baixou { background: var(--sucesso-fundo); color: var(--sucesso-texto); }
	.badge-subiu { background: var(--erro-fundo); color: var(--erro-texto); }
	tr.baixou .preco-atual { color: var(--sucesso-texto); }
	tr.subiu .preco-atual { color: var(--erro-texto); }
	.data { font-size: 0.78rem; color: var(--tinta-suave); }

	/* Responsivo mobile */
	@media (max-width: 600px) {
		.form-row { flex-direction: column; }
		.btn-adicionar { width: 100%; }
		.abas { overflow-x: auto; }
		.abas button { padding: 8px 12px; font-size: 0.8rem; }
		.alertas-acoes { flex-direction: column; }
	}

	/* Painel de alertas */
	.painel-alertas {
		margin-bottom: var(--r5);
	}
	.btn-toggle-alertas {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		background: var(--nevoa);
		cursor: pointer;
		font-size: 0.9rem;
		font-weight: 600;
		width: 100%;
		text-align: left;
	}
	.btn-toggle-alertas:hover { border-color: var(--ouro); }
	.badge-ativo {
		font-size: 0.7rem;
		background: var(--sucesso-fundo);
		color: var(--sucesso-texto);
		padding: 2px 8px;
		border-radius: var(--raio-full);
		font-weight: 700;
	}
	.badge-inativo {
		font-size: 0.7rem;
		background: var(--erro-fundo);
		color: var(--erro-texto);
		padding: 2px 8px;
		border-radius: var(--raio-full);
		font-weight: 700;
	}
	.alertas-config {
		border: 1px solid var(--linha);
		border-top: none;
		border-radius: 0 0 10px 10px;
		padding: var(--r4);
		background: var(--branco);
	}
	.campo-alerta {
		margin-bottom: var(--r3);
	}
	.campo-alerta label {
		display: block;
		font-size: 0.82rem;
		font-weight: 600;
		margin-bottom: 4px;
		color: var(--tinta);
	}
	.campo-alerta input[type="text"],
	.campo-alerta input[type="number"] {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid var(--linha);
		border-radius: 8px;
		font-size: 0.88rem;
	}
	.campo-alerta input:focus {
		outline: none;
		border-color: var(--ouro);
	}
	.campo-alerta.checkbox label {
		display: flex;
		align-items: center;
		gap: 8px;
		font-weight: normal;
		cursor: pointer;
	}
	.hint {
		font-size: 0.72rem;
		color: var(--tinta-suave);
		margin-top: 2px;
		display: block;
	}
	.alertas-acoes {
		display: flex;
		gap: var(--r3);
		margin-top: var(--r4);
	}
	.btn-salvar, .btn-testar {
		padding: 8px 16px;
		border-radius: 8px;
		border: 1px solid var(--linha);
		font-weight: 600;
		font-size: 0.85rem;
		cursor: pointer;
	}
	.btn-salvar { background: var(--ouro); color: white; border-color: var(--ouro); }
	.btn-salvar:hover:not(:disabled) { opacity: 0.9; }
	.btn-testar { background: var(--branco); }
	.btn-testar:hover:not(:disabled) { border-color: var(--ouro); }
	.btn-salvar:disabled, .btn-testar:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
