<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { listarDestinos, listarTemplates, agendarPublicacao, previewTemplate, resolverLinkShopee } from '$lib/api.js';
	import { recuperarProduto } from '$lib/publicar-store.js';
	import RichEditor from '$lib/components/RichEditor.svelte';
	import ResolverLink from '$lib/components/ResolverLink.svelte';
	import HeroProduto from '$lib/components/HeroProduto.svelte';
	import PublicarPreview from '$lib/components/PublicarPreview.svelte';
	import { Select, Button, Alert, Card } from '$lib/components/ui';

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

	async function resolverDadosProduto() {
		if (!produto || produto.imagem || !produto.link) return;
		try {
			const r = await Promise.race([
				resolverLinkShopee(produto.link),
				new Promise((_, rej) => setTimeout(() => rej(), 10000))
			]);
			if (r.imagem) produto = { ...produto, imagem: r.imagem };
			if (r.nome && !produto.nome) produto = { ...produto, nome: r.nome };
			if (r.preco && !produto.preco) produto = { ...produto, preco: r.preco };
		} catch {
			/* timeout ou erro de rede */
		}
	}

	async function carregarDestinosETemplates() {
		try {
			const timeout = (ms) => new Promise((_, rej) => setTimeout(() => rej(new Error('timeout')), ms));
			const buscar = async () => {
				const [rd, rt] = await Promise.all([
					Promise.race([listarDestinos(), timeout(15000)]).catch(() => null),
					Promise.race([listarTemplates(), timeout(15000)]).catch(() => null)
				]);
				return { rd, rt };
			};
			let { rd, rt } = await buscar();
			if (!rd && !rt) {
				await new Promise((r) => setTimeout(r, 1500));
				({ rd, rt } = await buscar());
			}
			destinos = rd?.destinos ?? [];
			templates = rt?.templates ?? [];
			if (templates.length > 0 && !templates.find((t) => t.id === templateId)) templateId = templates[0].id;
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}
	}

	onMount(async () => {
		const safety = setTimeout(() => {
			if (carregando) carregando = false;
		}, 20000);
		produto = recuperarProduto();
		if (!produto) produto = { id: '', nome: '', preco: 0, categoria: '', link: '', imagem: '' };
		await resolverDadosProduto();
		await carregarDestinosETemplates();
		clearTimeout(safety);
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
			const r = await previewTemplate({
				template_id: templateId || undefined,
				nome: produto.nome,
				preco: produto.preco,
				categoria: produto.categoria,
				link: produto.link,
				imagem: produto.imagem
			});
			atualizandoLegenda = true;
			legenda = r.preview || local();
			setTimeout(() => {
				atualizandoLegenda = false;
			}, 100);
		} catch {
			atualizandoLegenda = true;
			legenda = local();
			setTimeout(() => {
				atualizandoLegenda = false;
			}, 100);
		}
	}

	let lastTemplateId = $state('padrao');
	$effect(() => {
		if (templateId !== lastTemplateId) {
			lastTemplateId = templateId;
			legendaEditada = false;
			gerarLegenda();
		}
	});
	function onEditorChange(html) {
		if (!atualizandoLegenda) {
			legendaEditada = true;
			legenda = html;
		}
	}
	function resetarLegenda() {
		legendaEditada = false;
		gerarLegenda();
	}

	function handleLinkResolvido(dados) {
		produto = {
			...produto,
			link: dados.link || produto.link,
			nome: dados.nome || produto.nome || '',
			id: dados.id || produto.id || '',
			preco: dados.preco ?? produto.preco ?? 0,
			comissao: dados.comissao ?? produto.comissao ?? 0,
			imagem: dados.imagem || produto.imagem || ''
		};
		gerarLegenda();
	}

	async function enviarAgora() {
		publicando = true;
		resultado = null;
		erro = null;
		try {
			const payload = {
				...produto,
				produto_id: produto.id,
				destino_id: destinoId || null,
				template_id: templateId || null,
				legenda_custom: legenda || null
			};
			// Só inclui agendada_em se tem valor (evita enviar string vazia ao backend)
			if (agendamento) payload.agendada_em = new Date(agendamento).toISOString();
			const r = await agendarPublicacao(payload);
			resultado = r.publicacao;
		} catch (e) {
			erro = e.message;
		} finally {
			publicando = false;
		}
	}
</script>

<svelte:head><title>Publicar — Garimpei</title></svelte:head>

<section class="max-w-xl space-y-8">
	<Button variant="ghost" size="sm" onclick={() => goto('/')}>← Voltar</Button>

	{#if carregando}
		<p class="text-muted-foreground italic">Carregando…</p>
	{:else if !produto}
		<Card class="p-4">
			<p class="text-muted-foreground">{erro ?? 'Cole um link ou volte à curadoria para selecionar um produto.'}</p>
		</Card>
	{:else}
		<!-- Produto (hero) -->
		<HeroProduto bind:produto />

		<!-- Link (colar novo) -->
		<ResolverLink onresolvido={handleLinkResolvido} />

		<!-- Configuração -->
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
			<div class="flex flex-col gap-1.5">
				<span class="text-xs font-semibold text-muted-foreground">📡 Destino</span>
				{#if destinos.length === 0}
					<p class="text-xs text-muted-foreground">
						Nenhum destino. <a href="/canais" class="text-primary underline">Adicione</a>.
					</p>
				{:else}
					<Select
						bind:value={destinoId}
						options={destinos.map((d) => ({ value: d.id, label: `${d.nome} (${d.tipo})` }))}
						placeholder="Selecione…"
					/>
				{/if}
			</div>
			<div class="flex flex-col gap-1.5">
				<span class="text-xs font-semibold text-muted-foreground">🎨 Template</span>
				{#if templates.length === 0}
					<p class="text-xs text-muted-foreground">Formatação padrão.</p>
				{:else}
					<Select
						bind:value={templateId}
						options={templates.map((t) => ({ value: t.id, label: `${t.nome} ${t.com_foto ? '📷' : ''}` }))}
					/>
				{/if}
			</div>
			<div class="flex flex-col gap-1.5">
				<span class="text-xs font-semibold text-muted-foreground">⏱ Agendar</span>
				<input
					type="datetime-local"
					bind:value={agendamento}
					class="h-9 rounded-sm border border-input bg-background px-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
				/>
			</div>
		</div>

		<!-- Legenda -->
		<div>
			<div class="flex items-center justify-between mb-2">
				<span class="text-sm font-semibold">✏️ Legenda</span>
				{#if legendaEditada}
					<button
						class="text-xs font-semibold text-muted-foreground hover:text-primary transition-colors"
						onclick={resetarLegenda}
						type="button">↺ Resetar</button
					>
				{/if}
			</div>
			<RichEditor bind:content={legenda} placeholder="Legenda da publicação…" onchange={onEditorChange} />
		</div>

		<!-- Preview -->
		<PublicarPreview imagem={produto.imagem} {legenda} link={produto.link} />

		<!-- Ação -->
		<div>
			<Button
				variant="danger"
				size="lg"
				class="w-full text-base"
				onclick={enviarAgora}
				disabled={publicando || !destinoId}
			>
				{#if publicando}Enviando…{:else if agendamento}⏱ Agendar{:else}🚀 Enviar agora{/if}
			</Button>
			{#if !destinoId && destinos.length > 0}
				<p class="text-xs text-muted-foreground mt-2">Selecione um destino acima.</p>
			{/if}
		</div>

		{#if resultado}
			<Alert variant={resultado.status === 'erro' ? 'error' : 'success'}>
				{#if resultado.status === 'erro'}
					✕ {resultado.detalhe}
				{:else}
					✓ {resultado.status === 'enviada' ? 'Publicado' : 'Agendado'} com sucesso
					{#if resultado.detalhe}
						<code class="block mt-1 text-xs opacity-70">{resultado.detalhe}</code>
					{/if}
				{/if}
			</Alert>
		{/if}
		{#if erro && produto}
			<Alert variant="error">✕ {erro}</Alert>
		{/if}
	{/if}
</section>
