<script>
	/**
	 * Formulário para adicionar uma loja ao monitoramento.
	 * Aceita URL da Shopee ou ID numérico, com seleção de origem.
	 */
	import { adicionarLoja } from '$lib/api.js';
	import { buscasSalvas } from '$lib/buscas.js';
	import { Card, Button, Input, Alert } from '$lib/components/ui';

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

<Card padding="md">
	<h2 class="form-titulo">Adicionar loja</h2>
	<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
		<div class="form-row">
			<Input
				bind:value={inputLoja}
				placeholder="Cole a URL da loja (shopee.com.br/shop/123) ou ID numérico"
				disabled={adicionando}
			/>
			<Button
				type="submit"
				disabled={adicionando || !inputLoja.trim()}
				size="md"
			>
				{adicionando ? '⏳' : '➕'} Adicionar
			</Button>
		</div>
		<div class="form-row-origem">
			<label for="origem-padrao" class="label-origem">Origem dos produtos:</label>
			<select id="origem-padrao" bind:value={origemPadrao} class="select-origem">
				<option value="">— sem origem definida —</option>
				<option value="Coreia">🇰🇷 Coreia</option>
				<option value="Japão">🇯🇵 Japão</option>
				<option value="China">🇨🇳 China</option>
			</select>
			<span class="hint">Se a loja vende só produtos de um país, marque aqui para badge automático.</span>
		</div>
		{#if erroAdicionar}
			<Alert variant="error" inline>{erroAdicionar}</Alert>
		{/if}
		{#if sucessoAdicionar}
			<Alert variant="success" inline>{sucessoAdicionar}</Alert>
		{/if}
	</form>
</Card>

<style>
	.form-titulo {
		font-size: var(--text-lg);
		margin-bottom: var(--r3);
		color: var(--tinta);
	}
	.form-row {
		display: flex;
		gap: var(--r3);
		align-items: flex-end;
	}
	.form-row-origem {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--r2);
		margin-top: var(--r3);
	}
	.label-origem {
		font-size: var(--text-sm);
		font-weight: var(--font-semi);
		color: var(--tinta);
	}
	.select-origem {
		padding: var(--r2) var(--r3);
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		font-size: var(--text-sm);
		background: var(--branco);
		font-family: var(--ui);
	}
	.select-origem:focus {
		outline: none;
		border-color: var(--ouro);
	}
	.hint {
		font-size: var(--text-xs);
		color: var(--tinta-suave);
		margin-top: 2px;
		display: block;
	}

	@media (max-width: 600px) {
		.form-row {
			flex-direction: column;
		}
	}
</style>
