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
	let alertaThreshold = $state(15);
	let alertaApenasQuedas = $state(true);
	let salvandoAlerta = $state(false);
	let testandoAlerta = $state(false);
	let msgAlerta = $state(null);

	onMount(async () => {
		try {
			alertasConfig = await buscarAlertasConfig();
			if (alertasConfig) {
				alertaThreshold = Math.round((alertasConfig.threshold ?? 0.15) * 100);
				alertaApenasQuedas = alertasConfig.apenas_quedas ?? true;
			}
		} catch { /* sem alertas configurados */ }
	});

	async function handleSalvar() {
		salvandoAlerta = true;
		msgAlerta = null;
		try {
			await configurarAlertas({
				chatId: alertaChatId || undefined,
				threshold: alertaThreshold / 100,
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

<div class="painel-alertas">
	<button class="btn-toggle-alertas" onclick={() => mostraAlertas = !mostraAlertas}>
		🔔 Alertas Telegram
		{#if alertasConfig?.ativo}
			<Badge variant="green">Ativo</Badge>
		{:else}
			<Badge variant="red">Inativo</Badge>
		{/if}
	</button>

	{#if mostraAlertas}
		<div class="alertas-config">
			<Input
				bind:value={alertaChatId}
				label="Chat ID do grupo Telegram"
				placeholder={alertasConfig?.chat_id || 'Ex: -1001234567890'}
			/>
			<span class="hint">ID do grupo onde os alertas serão enviados. Use @BotFather para criar o bot.</span>

			<div class="campo-alerta">
				<Input
					bind:value={alertaThreshold}
					type="number"
					label="Threshold de variação (%)"
					variant="mono"
					size="sm"
				/>
				<span class="hint">Alerta se preço variar mais que {alertaThreshold}%.</span>
			</div>

			<div class="campo-alerta checkbox">
				<label>
					<input type="checkbox" bind:checked={alertaApenasQuedas} />
					Alertar apenas quedas de preço (oportunidades)
				</label>
			</div>

			<div class="alertas-acoes">
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

<style>
	.painel-alertas {
		margin-bottom: var(--r5);
	}
	.btn-toggle-alertas {
		display: flex;
		align-items: center;
		gap: var(--r2);
		padding: var(--r3) var(--r4);
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		background: var(--nevoa);
		cursor: pointer;
		font-family: var(--ui);
		font-size: var(--text-base);
		font-weight: var(--font-semi);
		width: 100%;
		text-align: left;
		transition: border-color 0.15s ease;
	}
	.btn-toggle-alertas:hover {
		border-color: var(--ouro);
	}
	.alertas-config {
		border: 1px solid var(--linha);
		border-top: none;
		border-radius: 0 0 var(--raio-sm) var(--raio-sm);
		padding: var(--r4);
		background: var(--branco);
		display: flex;
		flex-direction: column;
		gap: var(--r3);
	}
	.campo-alerta {
		display: flex;
		flex-direction: column;
		gap: var(--r1);
	}
	.campo-alerta.checkbox label {
		display: flex;
		align-items: center;
		gap: var(--r2);
		font-size: var(--text-sm);
		cursor: pointer;
	}
	.campo-alerta.checkbox input {
		width: 16px;
		height: 16px;
		accent-color: var(--ouro);
		cursor: pointer;
	}
	.hint {
		font-size: var(--text-xs);
		color: var(--tinta-suave);
	}
	.alertas-acoes {
		display: flex;
		gap: var(--r3);
		margin-top: var(--r2);
	}

	@media (max-width: 600px) {
		.alertas-acoes { flex-direction: column; }
	}
	@media (prefers-reduced-motion: reduce) {
		.btn-toggle-alertas { transition-duration: 0ms; }
	}
</style>
