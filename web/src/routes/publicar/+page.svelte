<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { listarDestinos, listarTemplates, agendarPublicacao, previewTemplate, resolverLinkShopee } from '$lib/api.js';
	import RichEditor from '$lib/components/RichEditor.svelte';
	import ResolverLink from '$lib/components/ResolverLink.svelte';
	import HeroProduto from '$lib/components/HeroProduto.svelte';
	import PublicarPreview from '$lib/components/PublicarPreview.svelte';
	import { Select } from '$lib/components/ui';

	let produto = $state(null);
	let destinos = $state([]);
	let templates = $state([]);
	let carregando = $state(true);

	let destinoId = $state('');
	let templateId = $state('padrao');
	let agendamento = $state('');

	let legenda = $state('');
	let legendaEditada = $state(false);
	let atualizandoLegenda = $state(false);

	let publicando = $state(false);
	let resultado = $state(null);
	let erro = $state(null);

	onMount(async () => {
		const dados = $page.url.searchParams.get('dados');
		if (dados) { try { produto = JSON.parse(decodeURIComponent(dados)); } catch { /* */ } }
		if (!produto) produto = { id: '', nome: '', preco: 0, categoria: '', link: '', imagem: '' };

		if (produto && !produto.imagem && produto.link) {
			try {
				const r = await Promise.race([resolverLinkShopee(produto.link), new Promise((_, rej) => setTimeout(() => rej(), 10000))]);
				if (r.imagem) produto = { ...produto, imagem: r.imagem };
				if (r.nome && !produto.nome) produto = { ...produto, nome: r.nome };
				if (r.preco && !produto.preco) produto = { ...produto, preco: r.preco };
			} catch { /* */ }
		}

		try {
			const [rd, rt] = await Promise.all([
				listarDestinos().catch(() => ({ destinos: [] })),
				listarTemplates().catch(() => ({ templates: [] }))
			]);
			destinos = rd?.destinos ?? [];
			templates = rt?.templates ?? [];
			if (templates.length > 0 && !templates.find(t => t.id === templateId)) templateId = templates[0].id;
		} catch (e) { erro = e.message; }
		finally { carregando = false; }
		gerarLegenda();
	});

	async function gerarLegenda() {
		if (!produto || legendaEditada) return;
		const local = () => {
			let t = '';
			if (produto.nome) t += `✨ <b>${produto.nome}</b>\n`;
			if (produto.categoria) t += `📂 <i>${produto.categoria}</i>\n`;
			if (produto.preco > 0) t += `💸 <b>R$ ${produto.preco.toFixed(2)}</b>`;
			return t.trimEnd();
		};
		try {
			const r = await previewTemplate({ template_id: templateId || undefined, nome: produto.nome, preco: produto.preco, categoria: produto.categoria, link: produto.link, imagem: produto.imagem });
			atualizandoLegenda = true;
			legenda = r.preview || local();
			setTimeout(() => { atualizandoLegenda = false; }, 100);
		} catch {
			atualizandoLegenda = true;
			legenda = local();
			setTimeout(() => { atualizandoLegenda = false; }, 100);
		}
	}

	let lastTemplateId = $state(templateId);
	$effect(() => { if (templateId !== lastTemplateId) { lastTemplateId = templateId; legendaEditada = false; gerarLegenda(); } });

	function onEditorChange(html) { if (!atualizandoLegenda) { legendaEditada = true; legenda = html; } }
	function resetarLegenda() { legendaEditada = false; gerarLegenda(); }

	function handleLinkResolvido(dados) {
		produto = { ...produto, link: dados.link || produto.link, nome: dados.nome || produto.nome || '', id: dados.id || produto.id || '', preco: dados.preco ?? produto.preco ?? 0, comissao: dados.comissao ?? produto.comissao ?? 0, imagem: dados.imagem || produto.imagem || '' };
		gerarLegenda();
	}

	async function enviarAgora() {
		publicando = true; resultado = null; erro = null;
		try {
			const r = await agendarPublicacao({ ...produto, produto_id: produto.id, destino_id: destinoId || undefined, template_id: templateId || undefined, agendada_em: agendamento ? new Date(agendamento).toISOString() : '', legenda_custom: legenda || undefined });
			resultado = r.publicacao;
		} catch (e) { erro = e.message; }
		finally { publicando = false; }
	}
</script>

<svelte:head><title>Publicar — Garimpei</title></svelte:head>

<section class="pub-page">
	<button class="voltar" onclick={() => goto('/')}>← Voltar</button>

	{#if carregando}
		<p class="loading">Carregando…</p>
	{:else if !produto}
		<div class="aviso">{erro ?? 'Cole um link ou volte à curadoria para selecionar um produto.'}</div>
	{:else}
		<!-- Produto (hero) -->
		<HeroProduto bind:produto={produto} />

		<!-- Link (colar novo) -->
		<ResolverLink onresolvido={handleLinkResolvido} />

		<!-- Configuração -->
		<div class="config-grid">
			<div class="campo">
				<label for="destino-sel">📡 Destino</label>
				{#if destinos.length === 0}
					<p class="dica">Nenhum destino. <a href="/canais">Adicione</a>.</p>
				{:else}
					<Select
						bind:value={destinoId}
						options={destinos.map(d => ({ value: d.id, label: `${d.nome} (${d.tipo})` }))}
						placeholder="Selecione…"
					/>
				{/if}
			</div>
			<div class="campo">
				<label for="template-sel">🎨 Template</label>
				{#if templates.length === 0}
					<p class="dica">Formatação padrão.</p>
				{:else}
					<Select
						bind:value={templateId}
						options={templates.map(t => ({ value: t.id, label: `${t.nome} ${t.com_foto ? '📷' : ''}` }))}
					/>
				{/if}
			</div>
			<div class="campo">
				<label for="agendar">⏱ Agendar (opcional)</label>
				<input id="agendar" type="datetime-local" bind:value={agendamento} />
			</div>
		</div>

		<!-- Legenda -->
		<div class="legenda-section">
			<div class="legenda-header">
				<label>✏️ Legenda</label>
				{#if legendaEditada}
					<button class="btn-reset" onclick={resetarLegenda} type="button">↺ Resetar</button>
				{/if}
			</div>
			<RichEditor bind:content={legenda} placeholder="Legenda da publicação…" onchange={onEditorChange} />
		</div>

		<!-- Preview -->
		<PublicarPreview imagem={produto.imagem} {legenda} link={produto.link} />

		<!-- Ação -->
		<div class="acao">
			<button class="btn-enviar" onclick={enviarAgora} disabled={publicando || !destinoId}>
				{#if publicando}Enviando…{:else if agendamento}⏱ Agendar{:else}🚀 Enviar agora{/if}
			</button>
			{#if !destinoId && destinos.length > 0}
				<p class="dica">Selecione um destino acima.</p>
			{/if}
		</div>

		{#if resultado}
			<div class="resultado" class:ok={resultado.status !== 'erro'} class:falha={resultado.status === 'erro'}>
				{#if resultado.status === 'erro'}
					<p>✕ {resultado.detalhe}</p>
				{:else}
					<p>✓ {resultado.status === 'enviada' ? 'Publicado' : 'Agendado'} com sucesso</p>
					{#if resultado.detalhe}<p class="subid"><code>{resultado.detalhe}</code></p>{/if}
				{/if}
			</div>
		{/if}
		{#if erro && produto}<div class="resultado falha"><p>✕ {erro}</p></div>{/if}
	{/if}
</section>

<style>
	.pub-page { max-width: 600px; }
	.voltar { border: 1px solid var(--linha); background: var(--porcelana); padding: 6px 14px; border-radius: 8px; font-size: 0.85rem; font-weight: 600; cursor: pointer; color: var(--tinta-suave); margin-bottom: var(--r5); }
	.voltar:hover { color: var(--tinta); border-color: var(--tinta-suave); }

	/* Config grid */
	.config-grid { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: var(--r4); margin: var(--r5) 0; }
	@media (max-width: 500px) { .config-grid { grid-template-columns: 1fr; } }
	.campo { display: flex; flex-direction: column; gap: 6px; }
	.campo label { font-weight: 600; font-size: 0.82rem; }
	.campo select, .campo input { padding: 9px 12px; border: 1px solid var(--linha); border-radius: var(--raio-sm); font-size: 0.88rem; background: var(--porcelana); }
	.dica { font-size: 0.78rem; color: var(--tinta-suave); margin: 0; }
	.dica a { color: var(--ouro); }

	/* Legenda */
	.legenda-section { margin-bottom: var(--r5); }
	.legenda-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
	.legenda-header label { font-weight: 600; font-size: 0.88rem; }
	.btn-reset { border: none; background: transparent; color: var(--tinta-suave); font-size: 0.75rem; font-weight: 600; cursor: pointer; }
	.btn-reset:hover { color: var(--ouro); }

	/* Ação */
	.acao { margin-bottom: var(--r5); }
	.btn-enviar { padding: 14px 32px; background: var(--rosa); color: white; font-weight: 700; font-size: 1rem; border: none; border-radius: var(--raio-sm); cursor: pointer; width: 100%; }
	.btn-enviar:hover { background: var(--rosa-hover); }
	.btn-enviar:disabled { opacity: 0.5; cursor: not-allowed; }

	/* Resultado */
	.resultado { padding: var(--r3) var(--r4); border-radius: var(--raio-sm); font-size: 0.88rem; }
	.resultado.ok { background: var(--sucesso-fundo); border: 1px solid var(--sucesso-borda); color: var(--sucesso-texto); }
	.resultado.falha { background: var(--erro-fundo); border: 1px solid var(--erro-borda); color: var(--erro-texto); }
	.resultado p { margin: 0; }
	.subid { font-size: 0.75rem; margin-top: 4px !important; }
	.subid code { font-size: 0.72rem; background: var(--sucesso-fundo); padding: 2px 6px; border-radius: 4px; }
	.loading { color: var(--tinta-suave); }
	.aviso { background: var(--porcelana); padding: var(--r4); border-radius: var(--raio-sm); color: var(--tinta-suave); }
</style>
