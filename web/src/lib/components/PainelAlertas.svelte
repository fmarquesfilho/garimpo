<script>
	/**
	 * Painel de configuração de alertas Telegram para variação de preço.
	 * Toggle expansível com config de chat_id, threshold e filtro.
	 */
	import { buscarAlertasConfig, testarAlertas, configurarAlertas } from '$lib/api.js';
	import { onMount } from 'svelte';

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
			<span class="badge-ativo">Ativo</span>
		{:else}
			<span class="badge-inativo">Inativo</span>
		{/if}
	</button>

	{#if mostraAlertas}
		<div class="alertas-config">
			<div class="campo-alerta">
				<label for="chat-id">Chat ID do grupo Telegram</label>
				<input id="chat-id" type="text" bind:value={alertaChatId}
					placeholder={alertasConfig?.chat_id || 'Ex: -1001234567890'} />
				<span class="hint">ID do grupo onde os alertas serão enviados. Use @BotFather para criar o bot.</span>
			</div>
			<div class="campo-alerta">
				<label for="threshold">Threshold de variação (%)</label>
				<input id="threshold" type="number" bind:value={alertaThreshold} min="5" max="50" step="5" />
				<span class="hint">Alerta se preço variar mais que {alertaThreshold}%.</span>
			</div>
			<div class="campo-alerta checkbox">
				<label>
					<input type="checkbox" bind:checked={alertaApenasQuedas} />
					Alertar apenas quedas de preço (oportunidades)
				</label>
			</div>
			<div class="alertas-acoes">
				<button onclick={handleSalvar} disabled={salvandoAlerta} class="btn-salvar">
					{salvandoAlerta ? '⏳' : '💾'} Salvar
				</button>
				<button onclick={handleTestar} disabled={testandoAlerta || !alertasConfig?.ativo} class="btn-testar">
					{testandoAlerta ? '⏳' : '📨'} Testar
				</button>
			</div>
			{#if msgAlerta}
				<p class={msgAlerta.tipo === 'erro' ? 'msg-erro-inline' : 'msg-sucesso'}>
					{msgAlerta.texto}
				</p>
			{/if}
		</div>
	{/if}
</div>

<style>
	.painel-alertas { margin-bottom: var(--r5); }
	.btn-toggle-alertas {
		display: flex; align-items: center; gap: 8px;
		padding: 10px 16px; border: 1px solid var(--linha);
		border-radius: var(--raio-sm); background: var(--nevoa);
		cursor: pointer; font-size: 0.9rem; font-weight: 600;
		width: 100%; text-align: left;
	}
	.btn-toggle-alertas:hover { border-color: var(--ouro); }
	.badge-ativo {
		font-size: 0.7rem; background: var(--sucesso-fundo);
		color: var(--sucesso-texto); padding: 2px 8px;
		border-radius: var(--raio-full); font-weight: 700;
	}
	.badge-inativo {
		font-size: 0.7rem; background: var(--erro-fundo);
		color: var(--erro-texto); padding: 2px 8px;
		border-radius: var(--raio-full); font-weight: 700;
	}
	.alertas-config {
		border: 1px solid var(--linha); border-top: none;
		border-radius: 0 0 10px 10px; padding: var(--r4); background: var(--branco);
	}
	.campo-alerta { margin-bottom: var(--r3); }
	.campo-alerta label {
		display: block; font-size: 0.82rem; font-weight: 600;
		margin-bottom: 4px; color: var(--tinta);
	}
	.campo-alerta input[type="text"],
	.campo-alerta input[type="number"] {
		width: 100%; padding: 8px 12px; border: 1px solid var(--linha);
		border-radius: 8px; font-size: 0.88rem;
	}
	.campo-alerta input:focus { outline: none; border-color: var(--ouro); }
	.campo-alerta.checkbox label {
		display: flex; align-items: center; gap: 8px;
		font-weight: normal; cursor: pointer;
	}
	.hint { font-size: 0.72rem; color: var(--tinta-suave); margin-top: 2px; display: block; }
	.alertas-acoes { display: flex; gap: var(--r3); margin-top: var(--r4); }
	.btn-salvar, .btn-testar {
		padding: 8px 16px; border-radius: 8px; border: 1px solid var(--linha);
		font-weight: 600; font-size: 0.85rem; cursor: pointer;
	}
	.btn-salvar { background: var(--ouro); color: white; border-color: var(--ouro); }
	.btn-salvar:hover:not(:disabled) { opacity: 0.9; }
	.btn-testar { background: var(--branco); }
	.btn-testar:hover:not(:disabled) { border-color: var(--ouro); }
	.btn-salvar:disabled, .btn-testar:disabled { opacity: 0.5; cursor: not-allowed; }
	.msg-erro-inline { color: var(--erro-texto); font-size: 0.85rem; margin-top: 6px; }
	.msg-sucesso { color: var(--sucesso-texto); font-size: 0.85rem; margin-top: 6px; }

	@media (max-width: 600px) {
		.alertas-acoes { flex-direction: column; }
	}
</style>
