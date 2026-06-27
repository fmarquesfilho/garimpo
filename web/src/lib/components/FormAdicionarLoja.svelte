<script>
	/**
	 * Formulário para adicionar uma loja ao monitoramento.
	 * Aceita URL da Shopee ou ID numérico, com seleção de origem.
	 */
	import { adicionarLoja } from '$lib/api.js';
	import { buscasSalvas } from '$lib/buscas.js';

	let { onadicionada = null } = $props();

	let inputLoja = $state('');
	let origemPadrao = $state('');
	let adicionando = $state(false);
	let erroAdicionar = $state(null);
	let sucessoAdicionar = $state(null);

	async function handleSubmit() {
		const valor = inputLoja.trim();
		if (!valor) return;

		adicionando = true;
		erroAdicionar = null;
		sucessoAdicionar = null;

		try {
			const r = await adicionarLoja({ input: valor, origemPadrao: origemPadrao || undefined });
			sucessoAdicionar = `Loja ${r.shop_id} adicionada com sucesso!`;
			inputLoja = '';
			origemPadrao = '';
			await buscasSalvas.sincronizarDoServidor();
			if (onadicionada) onadicionada(r);
			setTimeout(() => { sucessoAdicionar = null; }, 2000);
		} catch (e) {
			erroAdicionar = e.message;
		} finally {
			adicionando = false;
		}
	}
</script>

<div class="form-loja">
	<h2>Adicionar loja</h2>
	<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
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
		<div class="form-row-origem">
			<label for="origem-padrao">Origem dos produtos:</label>
			<select id="origem-padrao" bind:value={origemPadrao} class="select-origem">
				<option value="">— sem origem definida —</option>
				<option value="Coreia">🇰🇷 Coreia</option>
				<option value="Japão">🇯🇵 Japão</option>
				<option value="China">🇨🇳 China</option>
			</select>
			<span class="hint">Se a loja vende só produtos de um país, marque aqui para badge automático.</span>
		</div>
		{#if erroAdicionar}
			<p class="msg-erro-inline">{erroAdicionar}</p>
		{/if}
		{#if sucessoAdicionar}
			<p class="msg-sucesso">{sucessoAdicionar}</p>
		{/if}
	</form>
</div>

<style>
	.form-loja {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r4);
		margin-bottom: var(--r5);
	}
	h2 { font-size: 1.1rem; margin-bottom: 0.5rem; color: var(--tinta); }
	.form-row { display: flex; gap: var(--r3); align-items: stretch; }
	.form-row-origem {
		display: flex; flex-wrap: wrap; align-items: center;
		gap: var(--r2); margin-top: var(--r3);
	}
	.form-row-origem label { font-size: 0.82rem; font-weight: 600; color: var(--tinta); }
	.select-origem {
		padding: 6px 10px; border: 1px solid var(--linha);
		border-radius: 8px; font-size: 0.85rem; background: var(--branco);
	}
	.select-origem:focus { outline: none; border-color: var(--ouro); }
	.input-loja {
		flex: 1; padding: 10px 14px; border: 1px solid var(--linha);
		border-radius: 8px; font-size: 0.9rem; background: var(--branco);
	}
	.input-loja:focus { outline: none; border-color: var(--ouro); box-shadow: 0 0 0 2px var(--ouro-fundo); }
	.btn-adicionar {
		padding: 10px 18px; background: var(--ouro); color: white;
		border: none; border-radius: 8px; font-weight: 600;
		font-size: 0.9rem; cursor: pointer; white-space: nowrap;
	}
	.btn-adicionar:hover:not(:disabled) { opacity: 0.9; }
	.btn-adicionar:disabled { opacity: 0.5; cursor: not-allowed; }
	.hint { font-size: 0.72rem; color: var(--tinta-suave); margin-top: 2px; display: block; }
	.msg-erro-inline { color: var(--erro-texto); font-size: 0.85rem; margin-top: 6px; }
	.msg-sucesso { color: var(--sucesso-texto); font-size: 0.85rem; margin-top: 6px; }

	@media (max-width: 600px) {
		.form-row { flex-direction: column; }
		.btn-adicionar { width: 100%; }
	}
</style>
