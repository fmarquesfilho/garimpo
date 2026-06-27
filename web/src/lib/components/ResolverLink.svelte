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
		setTimeout(() => { linkAplicado = false; }, 4000);
		if (onresolvido) onresolvido(resultado);
	}

	async function colarDoClipboard() {
		try {
			const texto = await navigator.clipboard.readText();
			if (texto?.trim()) linkColado = texto.trim();
		} catch { /* permissão negada */ }
		if (linkColado.trim()) aplicarLink();
	}
</script>

<div class="campo-pub">
	<label>🔗 Link do produto</label>
	<div class="link-input">
		<input
			type="url"
			bind:value={linkColado}
			placeholder="Cole o link da Shopee aqui…"
			onkeydown={(e) => e.key === 'Enter' && aplicarLink()}
		/>
		<button type="button" class="btn-colar" onclick={colarDoClipboard} disabled={resolvendoLink}>
			{resolvendoLink ? '⏳ Resolvendo…' : '📋 Colar e aplicar'}
		</button>
	</div>
	{#if resolvendoLink}
		<p class="dica loading-msg">Buscando dados do produto…</p>
	{:else if linkAplicado}
		<p class="dica sucesso-msg">✓ Link aplicado — edite os campos abaixo se necessário.</p>
	{/if}
</div>

<style>
	.campo-pub { display: flex; flex-direction: column; gap: 8px; }
	.campo-pub label { font-weight: 600; font-size: 0.88rem; }
	.link-input { display: flex; gap: var(--r2); flex-wrap: wrap; }
	.link-input input {
		flex: 1; min-width: 200px; padding: 10px 14px; border: 1px solid var(--linha);
		border-radius: var(--raio-sm); font-size: 0.9rem; background: var(--porcelana);
	}
	.btn-colar {
		padding: 10px 18px; background: var(--ouro); border: 1px solid var(--ouro);
		color: white; font-weight: 600; font-size: 0.85rem;
		border-radius: var(--raio-sm); cursor: pointer; white-space: nowrap;
	}
	.btn-colar:hover:not(:disabled) { background: var(--ouro-escuro); }
	.btn-colar:disabled { opacity: 0.5; cursor: not-allowed; }
	.dica { font-size: 0.82rem; color: var(--tinta-suave); margin: 0; }
	.sucesso-msg { color: var(--sucesso-texto); }
	.loading-msg { color: var(--ouro); }
</style>
