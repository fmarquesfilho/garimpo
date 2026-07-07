<script>
	/**
	 * Formulário para adicionar uma loja ao monitoramento.
	 * Aceita URL da Shopee ou ID numérico, com seleção de origem, palavras-chave
	 * (filtro opcional das coletas agendadas) e agendamento da coleta periódica.
	 */
	import { adicionarLoja } from '$lib/api.js';
	import { buscasSalvas } from '$lib/buscas.js';
	import { Card, Button, Input, Alert, Select } from '$lib/components/ui';
	import TagInput from '$lib/components/TagInput.svelte';
	import AgendadorBusca from '$lib/components/AgendadorBusca.svelte';

	let { onadicionada = null } = $props();

	let inputLoja = $state('');
	let origemPadrao = $state('');
	let keywords = $state([]);
	// Loja monitorada sempre coleta periodicamente — padrão a cada 8h, editável.
	let cron = $state('0 */8 * * *');
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
			const r = await adicionarLoja({
				input: valor,
				origemPadrao: origemPadrao || undefined,
				keywords,
				cron: cron || undefined
			});
			sucessoAdicionar = `Loja "${r.keyword ?? valor}" adicionada com sucesso!`;
			inputLoja = '';
			origemPadrao = '';
			keywords = [];
			cron = '0 */8 * * *';
			await buscasSalvas.sincronizarDoServidor();
			if (onadicionada) onadicionada(r);
			setTimeout(() => {
				sucessoAdicionar = null;
			}, 2500);
		} catch (e) {
			erroAdicionar = e.message;
		} finally {
			adicionando = false;
		}
	}
</script>

<Card padding="md">
	<h2 class="mb-1 text-lg text-foreground">Adicionar loja</h2>
	<p class="mb-3 text-sm text-muted-foreground">
		Monitore uma loja Shopee com coleta agendada. Sem palavras-chave, acompanha todos os produtos; com palavras-chave,
		filtra o que é coletado para monitorar preços e novidades.
	</p>
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
			<span class="mt-0.5 block text-xs text-muted-foreground"
				>Se a loja vende só produtos de um país, marque aqui para badge automático.</span
			>
		</div>

		<div class="mt-4">
			<TagInput
				bind:tags={keywords}
				label="Palavras-chave (opcional)"
				placeholder="ex.: sérum, protetor solar… (deixe vazio para monitorar tudo)"
			/>
		</div>

		<div class="mt-4">
			<AgendadorBusca bind:value={cron} permitirNunca={false} />
		</div>

		{#if erroAdicionar}
			<Alert variant="error" inline>{erroAdicionar}</Alert>
		{/if}
		{#if sucessoAdicionar}
			<Alert variant="success" inline>{sucessoAdicionar}</Alert>
		{/if}
	</form>
</Card>
