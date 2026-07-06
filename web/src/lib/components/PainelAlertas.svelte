<script>
	/**
	 * Painel de configuração de alertas Telegram para variação de preço.
	 * Toggle expansível com config de chat_id, threshold e filtro.
	 */
	import { buscarAlertasConfig, testarAlertas, configurarAlertas } from '$lib/api.js';
	import { onMount } from 'svelte';
	import { Button, Badge, Alert, Input } from '$lib/components/ui';

	let { buscaSelecionada = null } = $props();

	let alertasConfig = $state(null);
	let mostraAlertas = $state(false);
	let alertaChatId = $state('');
	let alertaThreshold = $state('15');
	let alertaApenasQuedas = $state(true);
	let salvandoAlerta = $state(false);
	let testandoAlerta = $state(false);
	let msgAlerta = $state(null);

	onMount(async () => {
		try {
			alertasConfig = await buscarAlertasConfig();
			if (alertasConfig) {
				alertaThreshold = String(Math.round((alertasConfig.threshold ?? 0.15) * 100));
				alertaApenasQuedas = alertasConfig.apenas_quedas ?? true;
			}
		} catch {
			/* sem alertas configurados */
		}
	});

	async function handleSalvar() {
		salvandoAlerta = true;
		msgAlerta = null;
		try {
			await configurarAlertas({
				chatId: alertaChatId || undefined,
				threshold: Number(alertaThreshold) / 100,
				apenasQuedas: alertaApenasQuedas
			});
			alertasConfig = await buscarAlertasConfig();
			msgAlerta = { tipo: 'sucesso', texto: 'Configuração salva!' };
		} catch (e) {
			msgAlerta = { tipo: 'erro', texto: e.message };
		} finally {
			salvandoAlerta = false;
		}
	}

	async function handleTestar() {
		testandoAlerta = true;
		msgAlerta = null;
		try {
			const r = await testarAlertas({ buscaId: buscaSelecionada?.id });
			msgAlerta = { tipo: 'sucesso', texto: r.status };
		} catch (e) {
			msgAlerta = { tipo: 'erro', texto: e.message };
		} finally {
			testandoAlerta = false;
		}
	}
</script>

<div class="mb-5">
	<button
		class="flex w-full items-center gap-2 rounded-sm border border-border bg-card px-4 py-3 text-left font-[var(--ui)] font-semibold transition-[border-color] duration-150 ease-linear hover:border-primary motion-reduce:transition-none"
		onclick={() => (mostraAlertas = !mostraAlertas)}
	>
		🔔 Alertas Telegram
		{#if alertasConfig?.ativo}
			<Badge variant="success">Ativo</Badge>
		{:else}
			<Badge variant="error">Inativo</Badge>
		{/if}
	</button>

	{#if mostraAlertas}
		<div class="flex flex-col gap-3 rounded-b-sm border border-t-0 border-border bg-background p-4">
			<Input
				bind:value={alertaChatId}
				label="Chat ID do grupo Telegram"
				placeholder={alertasConfig?.chat_id || 'Ex: -1001234567890'}
			/>
			<span class="text-xs text-muted-foreground"
				>ID do grupo onde os alertas serão enviados. Use @BotFather para criar o bot.</span
			>

			<div class="flex flex-col gap-1">
				<Input bind:value={alertaThreshold} type="number" label="Threshold de variação (%)" variant="mono" size="sm" />
				<span class="text-xs text-muted-foreground">Alerta se preço variar mais que {alertaThreshold}%.</span>
			</div>

			<div class="flex flex-col gap-1">
				<label class="flex cursor-pointer items-center gap-2 text-sm">
					<input type="checkbox" class="h-4 w-4 cursor-pointer accent-primary" bind:checked={alertaApenasQuedas} />
					Alertar apenas quedas de preço (oportunidades)
				</label>
			</div>

			<div class="mt-2 flex flex-col gap-3 sm:flex-row">
				<Button onclick={handleSalvar} disabled={salvandoAlerta} size="sm">
					{salvandoAlerta ? '⏳' : '💾'} Salvar
				</Button>
				<Button variant="secondary" onclick={handleTestar} disabled={testandoAlerta || !alertasConfig?.ativo} size="sm">
					{testandoAlerta ? '⏳' : '📨'} Testar
				</Button>
			</div>

			{#if msgAlerta}
				<Alert variant={msgAlerta.tipo === 'erro' ? 'error' : 'success'} inline>
					{msgAlerta.texto}
				</Alert>
			{/if}
		</div>
	{/if}
</div>
