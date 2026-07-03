<script>
	/**
	 * Formulário para adicionar uma loja ao monitoramento.
	 * Aceita URL da Shopee ou ID numérico, com seleção de origem.
	 */
	import { adicionarLoja } from '$lib/api.js';
	import { buscasSalvas } from '$lib/buscas.js';
	import { Card, Button, Input, Alert, Select } from '$lib/components/ui';

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
			setTimeout(() => {
				sucessoAdicionar = null;
			}, 2000);
		} catch (e) {
			erroAdicionar = e.message;
		} finally {
			adicionando = false;
		}
	}
</script>

<Card padding="md">
	<h2 class="mb-3 text-lg text-foreground">Adicionar loja</h2>
	<form
		onsubmit={(e) => {
			e.preventDefault();
			handleSubmit();
		}}
	>
		<div class="flex flex-col items-end gap-3 sm:flex-row">
			<Input
				bind:value={inputLoja}
				placeholder="Cole a URL da loja (shopee.com.br/shop/123) ou ID numérico"
				disabled={adicionando}
			/>
			<Button type="submit" disabled={adicionando || !inputLoja.trim()} size="md">
				{adicionando ? '⏳' : '➕'} Adicionar
			</Button>
		</div>
		<div class="mt-3 flex flex-wrap items-center gap-2">
			<span class="text-sm font-semibold text-foreground">Origem dos produtos:</span>
			<Select
				bind:value={origemPadrao}
				options={[
					{ value: '', label: '— sem origem definida —' },
					{ value: 'Coreia', label: '🇰🇷 Coreia' },
					{ value: 'Japão', label: '🇯🇵 Japão' },
					{ value: 'China', label: '🇨🇳 China' }
				]}
				size="sm"
				placeholder="Selecione origem"
			/>
			<span class="mt-0.5 block text-xs text-tinta-suave"
				>Se a loja vende só produtos de um país, marque aqui para badge automático.</span
			>
		</div>
		{#if erroAdicionar}
			<Alert variant="error" inline>{erroAdicionar}</Alert>
		{/if}
		{#if sucessoAdicionar}
			<Alert variant="success" inline>{sucessoAdicionar}</Alert>
		{/if}
	</form>
</Card>
