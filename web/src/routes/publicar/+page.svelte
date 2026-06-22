<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { listarDestinos, listarTemplates, publicar, previewTemplate } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';

	// Produto vem via query params (serializado pelo card)
	let produto = $state(null);
	let destinos = $state([]);
	let templates = $state([]);
	let carregando = $state(true);

	// Seleções
	let destinoId = $state('');
	let templateId = $state('padrao');
	let preview = $state('');
	let previewFoto = $state(false);

	// Status
	let publicando = $state(false);
	let resultado = $state(null);
	let erro = $state(null);

	onMount(async () => {
		// Desserializa produto da URL
		const params = $page.url.searchParams;
		const dados = params.get('dados');
		if (dados) {
			try {
				produto = JSON.parse(decodeURIComponent(dados));
			} catch { /* */ }
		}
		if (!produto) {
			erro = 'Nenhum produto selecionado. Volte à curadoria e clique em Publicar.';
			carregando = false;
			return;
		}

		// Carrega destinos e templates
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

		atualizarPreview();
	});

	async function atualizarPreview() {
		if (!produto) return;
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
			preview = r.preview ?? '';
			previewFoto = r.com_foto && produto.imagem;
		} catch {
			preview = `✨ ${produto.nome}\n💸 R$ ${produto.preco?.toFixed(2)}`;
		}
	}

	// Recarrega preview quando template muda
	$effect(() => {
		templateId;
		atualizarPreview();
	});

	async function enviarAgora() {
		publicando = true;
		resultado = null;
		erro = null;
		try {
			const r = await publicar(produto, { destinoId: destinoId || undefined, templateId: templateId || undefined });
			resultado = r;
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
		<div class="aviso">{erro ?? 'Nenhum produto selecionado.'}</div>
	{:else}
		<div class="layout">
			<!-- Coluna esquerda: Produto + Configuração -->
			<div class="config">
				<!-- Resumo do produto -->
				<div class="card-produto">
					{#if produto.imagem}
						<img src={produto.imagem} alt={produto.nome} class="thumb" />
					{/if}
					<div class="produto-info">
						<h3>{produto.nome}</h3>
						<p class="meta">
							<span>{produto.categoria}</span> · <span class="preco">{brl(produto.preco)}</span>
						</p>
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

				<!-- Ações -->
				<div class="acoes">
					<button class="btn-enviar" onclick={enviarAgora} disabled={publicando}>
						{publicando ? 'Enviando…' : '🚀 Enviar agora'}
					</button>
				</div>

				{#if resultado}
					<div class="resultado ok">
						<p>✓ Publicado em <strong>{resultado.canal}</strong></p>
						{#if resultado.sub_id}
							<p class="subid">Atribuição: <code>{resultado.sub_id}</code></p>
						{/if}
					</div>
				{/if}

				{#if erro && produto}
					<div class="resultado falha">
						<p>✕ {erro}</p>
					</div>
				{/if}
			</div>

			<!-- Coluna direita: Preview -->
			<div class="preview-col">
				<h2>Preview</h2>
				<div class="preview-card">
					{#if previewFoto && produto.imagem}
						<img src={produto.imagem} alt="preview" class="preview-img" />
					{/if}
					<div class="preview-corpo">{@html preview.replace(/\n/g, '<br>')}</div>
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
	.produto-info h3 { font-size: 1rem; margin: 0 0 4px; }
	.produto-info .meta { font-size: 0.85rem; color: var(--tinta-suave); margin: 0; }
	.preco { font-weight: 700; color: var(--ouro); }

	.campo-pub { display: flex; flex-direction: column; gap: 6px; }
	.campo-pub label { font-weight: 600; font-size: 0.88rem; }
	.campo-pub select {
		padding: 10px 14px; border: 1px solid var(--linha); border-radius: 10px;
		font-size: 0.9rem; background: var(--porcelana);
	}
	.dica { font-size: 0.82rem; color: var(--tinta-suave); margin: 0; }
	.dica a { color: var(--ouro); text-decoration: underline; }

	.acoes { display: flex; gap: var(--r3); flex-wrap: wrap; }
	.btn-enviar {
		padding: 12px 28px; background: var(--rosa); color: white;
		font-weight: 700; font-size: 0.95rem; border: none; border-radius: 10px;
		cursor: pointer;
	}
	.btn-enviar:hover { background: #8f4c62; }
	.btn-enviar:disabled { opacity: 0.5; cursor: not-allowed; }

	.resultado {
		padding: var(--r3) var(--r4); border-radius: 10px; font-size: 0.88rem;
	}
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
	.preview-corpo {
		padding: var(--r4); font-size: 0.92rem; line-height: 1.5;
	}
	.preview-botao {
		padding: 0 var(--r4) var(--r4);
	}
	.btn-fake {
		display: inline-block; padding: 8px 18px; background: #0088cc;
		color: white; border-radius: 8px; font-size: 0.85rem; font-weight: 600;
	}
	.preview-nota {
		font-size: 0.75rem; color: var(--tinta-suave); margin-top: var(--r2);
		font-style: italic;
	}

	.loading { color: var(--tinta-suave); }
	.aviso { background: var(--porcelana); padding: var(--r4); border-radius: 10px; color: var(--tinta-suave); }
</style>
