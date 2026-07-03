<script>
	/**
	 * Campo de link do produto com resolução automática de links curtos da Shopee.
	 * Emite onresolvido({ nome, preco, link, imagem, comissao, id }) quando resolve.
	 */
	import { resolverLinkShopee } from '$lib/api.js';

	let { onresolvido = null } = $props();

	let linkColado = $state('');
	let resolvendoLink = $state(false);
	let linkAplicado = $state(false);

	async function aplicarLink() {
		const url = linkColado.trim();
		if (!url) return;

		linkAplicado = false;
		const isShortLink = /s\.shopee|shope\.ee/i.test(url) && !url.includes('-i.');

		let resultado = { link: url };

		if (isShortLink) {
			resolvendoLink = true;
			try {
				const r = await resolverLinkShopee(url);
				resultado = {
					link: r.link_afiliado || url,
					nome: r.nome || '',
					id: r.item_id || '',
					preco: r.preco ?? 0,
					comissao: r.comissao ?? 0,
					imagem: r.imagem || ''
				};
			} catch {
				resultado = { link: url };
			} finally {
				resolvendoLink = false;
			}
		} else {
			// Tenta extrair nome da URL longa
			const match = url.match(/\/([^/]+?)(?:-i\.\d+\.\d+)?(?:\?|$)/);
			if (match && match[1].length > 3) {
				resultado.nome = decodeURIComponent(match[1]).replace(/-/g, ' ');
			}
		}

		linkColado = '';
		linkAplicado = true;
		setTimeout(() => {
			linkAplicado = false;
		}, 4000);
		if (onresolvido) onresolvido(resultado);
	}

	async function colarDoClipboard() {
		try {
			const texto = await navigator.clipboard.readText();
			if (texto?.trim()) linkColado = texto.trim();
		} catch {
			/* permissão negada */
		}
		if (linkColado.trim()) aplicarLink();
	}
</script>

<div class="flex flex-col gap-2">
	<label for="link-produto" class="text-sm font-semibold">🔗 Link do produto</label>
	<div class="flex flex-wrap gap-2">
		<input
			id="link-produto"
			type="url"
			bind:value={linkColado}
			placeholder="Cole o link da Shopee aqui…"
			onkeydown={(e) => e.key === 'Enter' && aplicarLink()}
			class="min-w-[200px] flex-1 rounded-sm border border-border bg-porcelana px-3.5 py-2.5 text-[0.9rem]"
		/>
		<button
			type="button"
			class="whitespace-nowrap rounded-sm border border-ouro bg-[var(--ouro)] px-4 py-2.5 text-sm font-semibold text-white hover:bg-ouro-escuro disabled:cursor-not-allowed disabled:opacity-50"
			onclick={colarDoClipboard}
			disabled={resolvendoLink}
		>
			{resolvendoLink ? '⏳ Resolvendo…' : '📋 Colar e aplicar'}
		</button>
	</div>
	{#if resolvendoLink}
		<p class="m-0 text-sm text-ouro">Buscando dados do produto…</p>
	{:else if linkAplicado}
		<p class="m-0 text-sm text-[var(--sucesso-texto)]">✓ Link aplicado — edite os campos abaixo se necessário.</p>
	{/if}
</div>
