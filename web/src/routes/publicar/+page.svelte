<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { listarDestinos, listarTemplates, agendarPublicacao, previewTemplate, resolverLinkShopee } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';
	import RichEditor from '$lib/components/RichEditor.svelte';

	// Produto vem via query params
	let produto = $state(null);
	let destinos = $state([]);
	let templates = $state([]);
	let carregando = $state(true);

	// Seleções
	let destinoId = $state('');
	let templateId = $state('padrao');
	let agendamento = $state('');

	// Legenda editável
	let legenda = $state('');
	let legendaEditada = $state(false);
	let previewFoto = $state(false);
	let atualizandoLegenda = $state(false); // flag para ignorar onEditorChange programático

	// Colar link do produto
	let linkColado = $state('');
	let resolvendoLink = $state(false);
	let linkAplicado = $state(false); // feedback visual

	// Status
	let publicando = $state(false);
	let resultado = $state(null);
	let erro = $state(null);

	onMount(async () => {
		const params = $page.url.searchParams;
		const dados = params.get('dados');
		if (dados) {
			try { produto = JSON.parse(decodeURIComponent(dados)); } catch { /* */ }
		}
		// Se não veio produto via query, permite preencher manualmente
		if (!produto) {
			produto = { id: '', nome: '', preco: 0, categoria: '', estrategia: 'nicho', link: '', imagem: '' };
		}

		try {
			const [rd, rt] = await Promise.all([
				listarDestinos().catch(() => ({ destinos: [] })),
				listarTemplates().catch(() => ({ templates: [] }))
			]);
			destinos = rd?.destinos ?? [];
			templates = rt?.templates ?? [];
			if (templates.length > 0 && !templates.find(t => t.id === templateId)) {
				templateId = templates[0].id;
			}
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}

		gerarLegenda();
	});

	async function gerarLegenda() {
		if (!produto) return;
		// Não sobrescreve se o user editou manualmente
		if (legendaEditada) return;

		function legendaLocal() {
			let txt = '';
			if (produto.nome) txt += `✨ <b>${produto.nome}</b>\n`;
			if (produto.categoria) txt += `📂 <i>${produto.categoria}</i>\n`;
			if (produto.preco > 0) txt += `💸 <b>R$ ${produto.preco.toFixed(2)}</b>\n`;
			if (produto.estrategia) txt += `🎯 ${produto.estrategia}`;
			return txt.trimEnd();
		}

		let novaLegenda = '';
		try {
			const r = await previewTemplate({
				template_id: templateId || undefined,
				nome: produto.nome,
				preco: produto.preco,
				categoria: produto.categoria,
				estrategia: produto.estrategia,
				link: produto.link,
				imagem: produto.imagem
			});
			novaLegenda = r.preview || legendaLocal();
			previewFoto = r.com_foto && !!produto.imagem;
		} catch {
			novaLegenda = legendaLocal();
			previewFoto = false;
		}

		// Atualiza a legenda marcando que é programático (não do user)
		atualizandoLegenda = true;
		legenda = novaLegenda;
		// O $effect do RichEditor vai sincronizar, resetamos o flag após um tick
		setTimeout(() => { atualizandoLegenda = false; }, 100);
	}

	// Regenera legenda quando template muda (se user não editou)
	let lastTemplateId = $state(templateId);
	$effect(() => {
		if (templateId !== lastTemplateId) {
			lastTemplateId = templateId;
			legendaEditada = false;
			gerarLegenda();
		}
	});

	function onLegendaInput() {
		legendaEditada = true;
	}

	function resetarLegenda() {
		legendaEditada = false;
		gerarLegenda();
	}

	function onEditorChange(html) {
		// Ignora se a mudança veio de gerarLegenda (programática)
		if (atualizandoLegenda) return;
		legendaEditada = true;
		legenda = html;
	}

	async function aplicarLink() {
		const url = linkColado.trim();
		if (!url) return;

		linkAplicado = false;

		// Detecta link curto
		const isShortLink = /s\.shopee|shope\.ee/i.test(url) && !url.includes('-i.');

		if (isShortLink) {
			resolvendoLink = true;
			try {
				const r = await resolverLinkShopee(url);
				produto = {
					...produto,
					link: url, // mantém o link curto original (mais limpo no Telegram)
					nome: r.nome || produto.nome || '',
					id: r.item_id || '',
					preco: r.preco ?? produto.preco ?? 0,
					comissao: r.comissao ?? produto.comissao ?? 0,
					imagem: r.imagem || produto.imagem || '',
					categoria: produto.categoria || ''
				};
				// Se a API retornou link de afiliado, usa esse (tem tracking)
				if (r.link_afiliado) {
					produto = { ...produto, link: r.link_afiliado };
				}
			} catch {
				produto = { ...produto, link: url };
			} finally {
				resolvendoLink = false;
			}
		} else {
			produto = { ...produto, link: url };
			if (!produto.nome) {
				const match = url.match(/\/([^\/\?]+?)(?:-i\.\d+\.\d+)?(?:\?|$)/);
				if (match && match[1].length > 3) {
					produto = { ...produto, nome: decodeURIComponent(match[1]).replace(/-/g, ' ') };
				}
			}
		}

		linkColado = '';
		linkAplicado = true;
		setTimeout(() => { linkAplicado = false; }, 4000);
		gerarLegenda();
	}

	async function colarDoClipboard() {
		try {
			const texto = await navigator.clipboard.readText();
			if (texto?.trim()) {
				linkColado = texto.trim();
				// Se parece ser link da Shopee, aplica automaticamente
				if (/shopee|shope\.ee/i.test(linkColado)) {
					aplicarLink();
				}
			}
		} catch {
			// Permissão negada — o user cola manualmente no campo
		}
	}

	async function enviarAgora() {
		publicando = true;
		resultado = null;
		erro = null;
		try {
			const r = await agendarPublicacao({
				...produto,
				produto_id: produto.id,
				destino_id: destinoId || undefined,
				template_id: templateId || undefined,
				agendada_em: agendamento ? new Date(agendamento).toISOString() : '',
				legenda_custom: legenda || undefined
			});
			resultado = r.publicacao;
		} catch (e) {
			erro = e.message;
		} finally {
			publicando = false;
		}
	}

	const brl = (v) => v?.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' }) ?? '';
</script>

<svelte:head>
	<title>Publicar — Garimpo</title>
</svelte:head>

<section class="publicar-page">
	<div class="cabecalho">
		<button class="voltar" onclick={() => goto('/')}>← Voltar</button>
		<h1>📤 Publicar oferta</h1>
	</div>

	{#if carregando}
		<p class="loading">Carregando…</p>
	{:else if !produto}
		<div class="aviso">{erro ?? 'Cole um link ou volte à curadoria para selecionar um produto.'}</div>
	{:else}
		<div class="layout">
			<!-- Coluna esquerda: Configuração -->
			<div class="config">
				<!-- Colar link do produto -->
				<div class="campo-pub">
					<label>🔗 Link do produto</label>
					<div class="link-input">
						<input
							type="url"
							bind:value={linkColado}
							placeholder="Cole o link da Shopee aqui…"
							onkeydown={(e) => e.key === 'Enter' && aplicarLink()}
						/>
						<button type="button" class="btn-colar" onclick={colarDoClipboard}>📋 Colar</button>
						<button type="button" class="btn-link" onclick={aplicarLink} disabled={!linkColado.trim() || resolvendoLink}>
							{resolvendoLink ? '⏳' : 'Aplicar'}
						</button>
					</div>
					{#if resolvendoLink}
						<p class="dica loading-msg">⏳ Resolvendo link curto…</p>
					{:else if linkAplicado}
						<p class="dica sucesso-msg">✓ Link aplicado — edite os campos abaixo se necessário.</p>
					{/if}
				</div>

				<!-- Resumo do produto -->
				<div class="card-produto">
					{#if produto.imagem}
						<img src={produto.imagem} alt={produto.nome} class="thumb" />
					{/if}
					<div class="produto-info">
						<input class="nome-edit" bind:value={produto.nome} placeholder="Nome do produto" />
						<div class="meta-edit">
							<input class="campo-mini" bind:value={produto.categoria} placeholder="Categoria" />
							<input class="campo-mini preco-edit" type="number" step="0.01" bind:value={produto.preco} placeholder="Preço" />
						</div>
						{#if produto.link}
							<a class="link-preview" href={produto.link} target="_blank" rel="noopener">{produto.link.substring(0, 50)}…</a>
						{/if}
					</div>
				</div>

				<!-- Destino -->
				<div class="campo-pub">
					<label>📡 Destino</label>
					{#if destinos.length === 0}
						<p class="dica">Nenhum destino cadastrado. <a href="/canais">Adicione um destino</a> primeiro.</p>
					{:else}
						<select bind:value={destinoId}>
							<option value="">Canal padrão</option>
							{#each destinos as d (d.id)}
								<option value={d.id}>{d.nome} ({d.tipo})</option>
							{/each}
						</select>
					{/if}
				</div>

				<!-- Template -->
				<div class="campo-pub">
					<label>🎨 Template</label>
					{#if templates.length === 0}
						<p class="dica">Usando formatação padrão.</p>
					{:else}
						<select bind:value={templateId}>
							{#each templates as t (t.id)}
								<option value={t.id}>{t.nome} {t.com_foto ? '📷' : ''}</option>
							{/each}
						</select>
					{/if}
				</div>

				<!-- Legenda — editor WYSIWYG -->
				<div class="campo-pub">
					<div class="legenda-header">
						<label>✏️ Legenda</label>
						{#if legendaEditada}
							<button class="btn-reset" onclick={resetarLegenda} type="button">↺ Resetar do template</button>
						{/if}
					</div>
					<RichEditor bind:content={legenda} placeholder="Escreva a legenda da publicação…" onchange={onEditorChange} />
					{#if legendaEditada}
						<p class="dica-editada">Legenda editada. O preview à direita reflete o que será enviado.</p>
					{/if}
				</div>

				<!-- Agendamento -->
				<div class="campo-pub">
					<label>⏱ Agendar para (opcional)</label>
					<input type="datetime-local" bind:value={agendamento} />
				</div>

				<!-- Ações -->
				<div class="acoes">
					<button class="btn-enviar" onclick={enviarAgora} disabled={publicando}>
						{#if publicando}
							Enviando…
						{:else if agendamento}
							⏱ Agendar
						{:else}
							🚀 Enviar agora
						{/if}
					</button>
				</div>

				{#if resultado}
					<div class="resultado ok">
						<p>✓ {resultado.status === 'enviada' ? 'Publicado' : 'Agendado'} com sucesso</p>
						{#if resultado.detalhe}
							<p class="subid">Atribuição: <code>{resultado.detalhe}</code></p>
						{/if}
					</div>
				{/if}
				{#if erro && produto}
					<div class="resultado falha"><p>✕ {erro}</p></div>
				{/if}
			</div>

			<!-- Coluna direita: Preview -->
			<div class="preview-col">
				<h2>Preview</h2>
				<div class="preview-card">
					{#if previewFoto && produto.imagem}
						<img src={produto.imagem} alt="preview" class="preview-img" />
					{/if}
					<div class="preview-corpo">{@html legenda.replace(/\n/g, '<br>')}</div>
					{#if produto.link}
						<div class="preview-botao">
							<span class="btn-fake">🛒 Comprar</span>
						</div>
					{/if}
				</div>
				<p class="preview-nota">Como ficará no Telegram</p>
			</div>
		</div>
	{/if}
</section>

<style>
	.publicar-page { max-width: 900px; }
	.cabecalho { display: flex; align-items: center; gap: var(--r4); margin-bottom: var(--r6); }
	.cabecalho h1 { font-size: 1.4rem; margin: 0; }
	.voltar {
		border: 1px solid var(--linha); background: var(--porcelana);
		padding: 6px 14px; border-radius: 8px; font-size: 0.85rem;
		font-weight: 600; cursor: pointer; color: var(--tinta-suave);
	}
	.voltar:hover { color: var(--tinta); border-color: var(--tinta-suave); }

	.layout { display: grid; grid-template-columns: 1fr 1fr; gap: var(--r8); }
	@media (max-width: 700px) { .layout { grid-template-columns: 1fr; } }

	.config { display: flex; flex-direction: column; gap: var(--r5); }

	.card-produto {
		display: flex; gap: var(--r4); padding: var(--r4);
		border: 1px solid var(--linha); border-radius: 12px; background: var(--nevoa);
	}
	.thumb { width: 80px; height: 80px; object-fit: cover; border-radius: 8px; }
	.produto-info { flex: 1; display: flex; flex-direction: column; gap: var(--r2); }

	.campo-pub { display: flex; flex-direction: column; gap: 8px; }
	.campo-pub label { font-weight: 600; font-size: 0.88rem; }
	.campo-pub select, .campo-pub input[type="datetime-local"] {
		padding: 10px 14px; border: 1px solid var(--linha); border-radius: 10px;
		font-size: 0.9rem; background: var(--porcelana);
	}
	.dica { font-size: 0.82rem; color: var(--tinta-suave); margin: 0; }
	.dica a { color: var(--ouro); text-decoration: underline; }
	.sucesso-msg { color: #166534; }
	.loading-msg { color: var(--ouro); }

	/* Link input */
	.link-input { display: flex; gap: var(--r2); flex-wrap: wrap; }
	.link-input input {
		flex: 1; min-width: 200px; padding: 10px 14px; border: 1px solid var(--linha);
		border-radius: 10px; font-size: 0.9rem; background: var(--porcelana);
	}
	.btn-link {
		padding: 10px 16px; background: var(--ouro-fundo); border: 1px solid var(--ouro);
		color: #7a5a1e; font-weight: 600; font-size: 0.85rem;
		border-radius: 10px; cursor: pointer; white-space: nowrap;
	}
	.btn-link:disabled { opacity: 0.4; cursor: not-allowed; }
	.btn-colar {
		padding: 10px 14px; background: var(--porcelana); border: 1px solid var(--linha);
		color: var(--tinta-suave); font-weight: 600; font-size: 0.85rem;
		border-radius: 10px; cursor: pointer; white-space: nowrap;
	}
	.btn-colar:hover { border-color: var(--ouro); color: var(--ouro); }

	/* Produto editável */
	.nome-edit {
		font-size: 1rem; font-weight: 700;
		border: 1px solid var(--linha); background: white;
		border-radius: 8px;
		width: 100%; padding: 8px 12px;
	}
	.nome-edit::placeholder { color: var(--tinta-suave); opacity: 0.6; font-weight: 400; }
	.nome-edit:focus { outline: 2px solid var(--ouro); outline-offset: 1px; }
	.meta-edit { display: flex; gap: var(--r3); flex-wrap: wrap; }
	.campo-mini {
		font-size: 0.85rem; padding: 6px 10px; border: 1px solid var(--linha);
		border-radius: 8px; background: var(--porcelana); width: 120px;
	}
	.campo-mini:focus { outline: 2px solid var(--ouro); outline-offset: 1px; }
	.preco-edit { width: 90px; font-weight: 600; }
	.link-preview {
		font-size: 0.75rem; color: var(--tinta-suave); display: block;
		margin-top: var(--r2); text-decoration: none; overflow: hidden;
		text-overflow: ellipsis; white-space: nowrap;
	}
	.link-preview:hover { color: var(--ouro); }

	/* Legenda */
	.legenda-header { display: flex; align-items: center; justify-content: space-between; }
	.btn-reset {
		border: none; background: transparent; color: var(--tinta-suave);
		font-size: 0.78rem; font-weight: 600; cursor: pointer;
	}
	.btn-reset:hover { color: var(--ouro); }
	.dica-editada { font-size: 0.75rem; color: var(--ouro); margin: 0; font-style: italic; }

	.acoes { display: flex; gap: var(--r3); flex-wrap: wrap; }
	.btn-enviar {
		padding: 12px 28px; background: var(--rosa); color: white;
		font-weight: 700; font-size: 0.95rem; border: none; border-radius: 10px;
		cursor: pointer;
	}
	.btn-enviar:hover { background: #8f4c62; }
	.btn-enviar:disabled { opacity: 0.5; cursor: not-allowed; }

	.resultado { padding: var(--r3) var(--r4); border-radius: 10px; font-size: 0.88rem; }
	.resultado.ok { background: #f0fdf4; border: 1px solid #bbf7d0; color: #166534; }
	.resultado.falha { background: #fef2f2; border: 1px solid #fecaca; color: #b91c1c; }
	.resultado p { margin: 0; }
	.subid { font-size: 0.78rem; margin-top: 4px !important; }
	.subid code { font-size: 0.75rem; background: #dcfce7; padding: 2px 6px; border-radius: 4px; }

	/* Preview */
	.preview-col h2 { font-size: 1rem; margin: 0 0 var(--r3); color: var(--tinta-suave); }
	.preview-card {
		border: 1px solid var(--linha); border-radius: 12px; overflow: hidden;
		background: white; box-shadow: 0 2px 8px rgba(0,0,0,0.06);
	}
	.preview-img { width: 100%; max-height: 240px; object-fit: cover; }
	.preview-corpo { padding: var(--r4); font-size: 0.92rem; line-height: 1.5; }
	.preview-botao { padding: 0 var(--r4) var(--r4); }
	.btn-fake {
		display: inline-block; padding: 8px 18px; background: #0088cc;
		color: white; border-radius: 8px; font-size: 0.85rem; font-weight: 600;
	}
	.preview-nota { font-size: 0.75rem; color: var(--tinta-suave); margin-top: var(--r2); font-style: italic; }
	.loading { color: var(--tinta-suave); }
	.aviso { background: var(--porcelana); padding: var(--r4); border-radius: 10px; color: var(--tinta-suave); }
</style>
