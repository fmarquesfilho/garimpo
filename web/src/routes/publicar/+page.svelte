<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { listarDestinos, listarTemplates, agendarPublicacao, previewTemplate, resolverLinkShopee } from '$lib/api.js';
	import RichEditor from '$lib/components/RichEditor.svelte';
	import ResolverLink from '$lib/components/ResolverLink.svelte';
	import PreviewPublicacao from '$lib/components/PreviewPublicacao.svelte';

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
	let atualizandoLegenda = $state(false);

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
		if (!produto) {
			produto = { id: '', nome: '', preco: 0, categoria: '', estrategia: 'nicho', link: '', imagem: '' };
		}

		// Se veio sem imagem mas com link, tenta resolver dados completos (best-effort, 10s max)
		if (produto && !produto.imagem && produto.link) {
			try {
				const resolver = resolverLinkShopee(produto.link);
				const timeout = new Promise((_, rej) => setTimeout(() => rej(new Error('timeout')), 10000));
				const r = await Promise.race([resolver, timeout]);
				if (r.imagem) produto = { ...produto, imagem: r.imagem };
				if (r.nome && !produto.nome) produto = { ...produto, nome: r.nome };
				if (r.preco && !produto.preco) produto = { ...produto, preco: r.preco };
				if (r.comissao && !produto.comissao) produto = { ...produto, comissao: r.comissao };
			} catch { /* falha silenciosa — continua sem imagem */ }
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
		if (legendaEditada) return;

		function legendaLocal() {
			let txt = '';
			if (produto.nome) txt += `✨ <b>${produto.nome}</b>\n`;
			if (produto.categoria) txt += `📂 <i>${produto.categoria}</i>\n`;
			if (produto.preco > 0) txt += `💸 <b>R$ ${produto.preco.toFixed(2)}</b>`;
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

		atualizandoLegenda = true;
		legenda = novaLegenda;
		setTimeout(() => { atualizandoLegenda = false; }, 100);
	}

	// Regenera legenda quando template muda
	let lastTemplateId = $state(templateId);
	$effect(() => {
		if (templateId !== lastTemplateId) {
			lastTemplateId = templateId;
			legendaEditada = false;
			gerarLegenda();
		}
	});

	function resetarLegenda() {
		legendaEditada = false;
		gerarLegenda();
	}

	function onEditorChange(html) {
		if (atualizandoLegenda) return;
		legendaEditada = true;
		legenda = html;
	}

	function handleLinkResolvido(dados) {
		produto = {
			...produto,
			link: dados.link || produto.link,
			nome: dados.nome || produto.nome || '',
			id: dados.id || produto.id || '',
			preco: dados.preco ?? produto.preco ?? 0,
			comissao: dados.comissao ?? produto.comissao ?? 0,
			imagem: dados.imagem || produto.imagem || '',
			categoria: produto.categoria || ''
		};
		gerarLegenda();
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
</script>

<svelte:head>
	<title>Publicar — Garimpei</title>
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
				<ResolverLink onresolvido={handleLinkResolvido} />

				<!-- Resumo do produto -->
				<div class="card-produto">
					{#if produto.imagem}
						<img src={produto.imagem} alt={produto.nome} class="thumb" />
					{:else}
						<div class="thumb-placeholder">📦</div>
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
							<option value="" disabled>Selecione um destino…</option>
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
					<button class="btn-enviar" onclick={enviarAgora} disabled={publicando || !destinoId}>
						{#if publicando}
							Enviando…
						{:else if agendamento}
							⏱ Agendar
						{:else}
							🚀 Enviar agora
						{/if}
					</button>
					{#if !destinoId && destinos.length > 0}
						<p class="dica">Selecione um destino acima para enviar.</p>
					{/if}
				</div>

				{#if resultado}
					{#if resultado.status === 'erro'}
						<div class="resultado falha">
							<p>✕ Erro ao publicar: {resultado.detalhe}</p>
						</div>
					{:else}
						<div class="resultado ok">
							<p>✓ {resultado.status === 'enviada' ? 'Publicado' : 'Agendado'} com sucesso</p>
							{#if resultado.detalhe}
								<p class="subid">Atribuição: <code>{resultado.detalhe}</code></p>
							{/if}
						</div>
					{/if}
				{/if}
				{#if erro && produto}
					<div class="resultado falha"><p>✕ {erro}</p></div>
				{/if}
			</div>

			<!-- Coluna direita: Preview -->
			<PreviewPublicacao
				{legenda}
				imagem={produto.imagem}
				link={produto.link}
				{previewFoto}
			/>
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

	.config { display: flex; flex-direction: column; gap: var(--r6); }

	.card-produto {
		display: flex; gap: var(--r4); padding: var(--r4);
		border: 1px solid var(--linha); border-radius: var(--raio); background: var(--nevoa);
	}
	.thumb { width: 80px; height: 80px; object-fit: cover; border-radius: 8px; }
	.thumb-placeholder { width: 80px; height: 80px; border-radius: 8px; background: var(--porcelana); display: flex; align-items: center; justify-content: center; font-size: 2rem; flex-shrink: 0; }
	.produto-info { flex: 1; display: flex; flex-direction: column; gap: var(--r2); }

	.campo-pub { display: flex; flex-direction: column; gap: 8px; }
	.campo-pub label { font-weight: 600; font-size: 0.88rem; }
	.campo-pub select, .campo-pub input[type="datetime-local"] {
		padding: 10px 14px; border: 1px solid var(--linha); border-radius: var(--raio-sm);
		font-size: 0.9rem; background: var(--porcelana);
	}
	.dica { font-size: 0.82rem; color: var(--tinta-suave); margin: 0; }
	.dica a { color: var(--ouro); text-decoration: underline; }

	/* Produto editável */
	.nome-edit {
		font-size: 1rem; font-weight: 700; border: 1px solid var(--linha);
		background: var(--branco); border-radius: 8px; width: 100%; padding: 8px 12px;
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
		font-weight: 700; font-size: 0.95rem; border: none; border-radius: var(--raio-sm);
		cursor: pointer;
	}
	.btn-enviar:hover { background: var(--rosa-hover); }
	.btn-enviar:disabled { opacity: 0.5; cursor: not-allowed; }

	.resultado { padding: var(--r3) var(--r4); border-radius: var(--raio-sm); font-size: 0.88rem; }
	.resultado.ok { background: var(--sucesso-fundo); border: 1px solid var(--sucesso-borda); color: var(--sucesso-texto); }
	.resultado.falha { background: var(--erro-fundo); border: 1px solid var(--erro-borda); color: var(--erro-texto); }
	.resultado p { margin: 0; }
	.subid { font-size: 0.78rem; margin-top: 4px !important; }
	.subid code { font-size: 0.75rem; background: var(--sucesso-fundo); padding: 2px 6px; border-radius: 4px; }

	.loading { color: var(--tinta-suave); }
	.aviso { background: var(--porcelana); padding: var(--r4); border-radius: var(--raio-sm); color: var(--tinta-suave); }
</style>
