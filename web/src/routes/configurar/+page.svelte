<script>
	import { onMount } from 'svelte';
	import { usuario } from '$lib/firebase.js';
	import {
		onboardingStatus,
		onboardingTermos,
		onboardingShopee,
		onboardingTelegram,
		onboardingWhatsapp,
		onboardingValidar,
		excluirConta
	} from '$lib/api.js';
	import { Button, Alert } from '$lib/components/ui';

	let step = $state(0);
	let carregando = $state(true);
	let erro = $state(null);
	let sucesso = $state(null);
	let configurado = $state(false);

	// Step 2 — Shopee
	let appId = $state('');
	let secret = $state('');
	let salvandoShopee = $state(false);

	// Step 3 — Canal de publicação (Telegram ou WhatsApp)
	let canalEscolhido = $state('telegram'); // 'telegram' | 'whatsapp'
	let telegramToken = $state('');
	let telegramChatId = $state('');
	let whatsappPhoneId = $state('');
	let whatsappToken = $state('');
	let salvandoCanal = $state(false);

	// Step 4 — Validação
	let validando = $state(false);

	// Exclusão
	let mostraExcluir = $state(false);
	let excluindo = $state(false);

	onMount(carregar);

	async function carregar() {
		carregando = true;
		erro = null;
		try {
			const r = await onboardingStatus();
			step = r.onboarding_step ?? r.step ?? 0;
			configurado = r.configurado ?? false;
			if (r.shopee_app_id) appId = r.shopee_app_id;
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}
	}

	async function aceitarTermos() {
		erro = null;
		try {
			const r = await onboardingTermos();
			step = r.step;
			sucesso = 'Termos aceitos!';
			setTimeout(() => (sucesso = null), 2000);
		} catch (e) {
			erro = e.message;
		}
	}

	async function salvarShopee() {
		if (!appId.trim() || !secret.trim()) {
			erro = 'Preencha AppID e Secret';
			return;
		}
		salvandoShopee = true;
		erro = null;
		try {
			const r = await onboardingShopee({ appId: appId.trim(), secret: secret.trim() });
			step = r.step;
			sucesso = 'Credenciais salvas!';
			secret = '';
			setTimeout(() => (sucesso = null), 2000);
		} catch (e) {
			erro = e.message;
		} finally {
			salvandoShopee = false;
		}
	}

	async function salvarCanal(pular = false) {
		salvandoCanal = true;
		erro = null;
		try {
			let r;
			if (pular) {
				r = await onboardingTelegram({ pular: true });
			} else if (canalEscolhido === 'telegram') {
				r = await onboardingTelegram({ token: telegramToken.trim(), chatId: telegramChatId.trim() });
			} else {
				r = await onboardingWhatsapp({ phoneNumberId: whatsappPhoneId.trim(), accessToken: whatsappToken.trim() });
			}
			step = r.step;
			sucesso = pular
				? 'Canais pulados (pode configurar depois).'
				: `${canalEscolhido === 'telegram' ? 'Telegram' : 'WhatsApp'} configurado!`;
			setTimeout(() => (sucesso = null), 2000);
		} catch (e) {
			erro = e.message;
		} finally {
			salvandoCanal = false;
		}
	}

	async function validar() {
		validando = true;
		erro = null;
		try {
			const r = await onboardingValidar();
			step = r.step;
			configurado = true;
			sucesso = '✅ Tudo pronto! Suas credenciais foram validadas.';
		} catch (e) {
			erro = e.message;
		} finally {
			validando = false;
		}
	}

	async function handleExcluir() {
		excluindo = true;
		try {
			await excluirConta();
			sucesso = 'Conta excluída. Seus dados foram removidos.';
			step = 0;
			configurado = false;
			mostraExcluir = false;
		} catch (e) {
			erro = e.message;
		} finally {
			excluindo = false;
		}
	}
</script>

<svelte:head>
	<title>Configurar Conta — Garimpei</title>
</svelte:head>

<section class="config-page">
	<h1>⚙️ Configurar Conta</h1>
	<p class="subtitulo">Configure suas credenciais para usar o Garimpei com sua conta de afiliado.</p>

	{#if !$usuario}
		<div class="aviso">Faça login para configurar sua conta.</div>
	{:else if carregando}
		<p class="loading">Carregando...</p>
	{:else}
		{#if erro}
			<Alert variant="error">{erro}</Alert>
		{/if}
		{#if sucesso}
			<Alert variant="success">{sucesso}</Alert>
		{/if}

		<!-- Progress bar -->
		<div class="progress">
			{#each [1, 2, 3, 4] as s}
				<div class="progress-step" class:done={step >= s} class:current={step === s - 1}>
					<span class="step-num">{s}</span>
					<span class="step-label">
						{#if s === 1}Termos{:else if s === 2}Shopee{:else if s === 3}Telegram{:else}Validar{/if}
					</span>
				</div>
			{/each}
		</div>

		{#if configurado}
			<div class="painel-ok">
				<h2>✅ Conta configurada!</h2>
				<p>Suas credenciais estão ativas. O sistema usará seus tokens pessoais para coletas e publicações.</p>
				<p class="meta-info">AppID: <code>{appId}</code></p>
			</div>
		{/if}

		<!-- Step 1: Termos -->
		{#if step < 1}
			<div class="step-card">
				<h2>1. Termos de Uso e Privacidade</h2>
				<div class="termos-box">
					<p><strong>Ao usar o Garimpei, você concorda que:</strong></p>
					<ul>
						<li>Seus dados pessoais (email, UID) são armazenados para funcionamento do serviço</li>
						<li>Suas credenciais de API são criptografadas e nunca compartilhadas</li>
						<li>Você pode solicitar exclusão completa dos seus dados a qualquer momento</li>
						<li>O serviço está em fase beta — sem garantias de disponibilidade</li>
					</ul>
				</div>
				<Button onclick={aceitarTermos}>Aceitar e Continuar</Button>
			</div>
		{/if}

		<!-- Step 2: Shopee -->
		{#if step >= 1 && step < 2}
			<div class="step-card">
				<h2>2. Credenciais Shopee Affiliate API</h2>
				<div class="instrucoes">
					<p>Você precisa de um App ID e Secret do painel de afiliados da Shopee:</p>
					<ol>
						<li>
							Acesse <a href="https://affiliate.shopee.com.br/open_api" target="_blank" rel="noopener"
								>affiliate.shopee.com.br/open_api</a
							>
						</li>
						<li>Faça login com sua conta de afiliado</li>
						<li>Na seção "Gerenciamento de API", copie o <strong>App ID</strong> e <strong>Secret</strong></li>
						<li>Se não tiver acesso, solicite via Central de Ajuda (demora ~2 semanas)</li>
					</ol>
				</div>
				<div class="form-campos">
					<label>
						App ID
						<input type="text" bind:value={appId} placeholder="Ex: 18332030606" />
					</label>
					<label>
						Secret
						<input type="password" bind:value={secret} placeholder="Ex: MJS67QHU7HMCRX5..." />
					</label>
				</div>
				<Button onclick={salvarShopee} disabled={salvandoShopee}>
					{salvandoShopee ? '⏳ Salvando...' : '💾 Salvar Credenciais'}
				</Button>
			</div>
		{/if}

		<!-- Step 3: Canal de publicação (Telegram ou WhatsApp) -->
		{#if step >= 2 && step < 3}
			<div class="step-card">
				<h2>3. Canal de Publicação</h2>
				<p class="instrucoes-intro">Configure pelo menos um canal para publicar ofertas.</p>

				<!-- Seletor de canal -->
				<div class="canal-tabs">
					<button class="tab" class:ativo={canalEscolhido === 'telegram'} onclick={() => (canalEscolhido = 'telegram')}>
						✈️ Telegram
					</button>
					<button class="tab" class:ativo={canalEscolhido === 'whatsapp'} onclick={() => (canalEscolhido = 'whatsapp')}>
						💬 WhatsApp
					</button>
				</div>

				{#if canalEscolhido === 'telegram'}
					<div class="instrucoes">
						<p>Configure um bot para publicar ofertas no Telegram:</p>
						<ol>
							<li>
								Abra o Telegram e converse com <a href="https://t.me/BotFather" target="_blank" rel="noopener"
									>@BotFather</a
								>
							</li>
							<li>Envie <code>/newbot</code> e siga as instruções para criar um bot</li>
							<li>Copie o <strong>Token</strong> fornecido</li>
							<li>Crie um grupo/canal, adicione o bot como admin</li>
							<li>
								Pegue o <strong>Chat ID</strong> (use
								<a href="https://t.me/getmyid_bot" target="_blank" rel="noopener">@getmyid_bot</a>)
							</li>
						</ol>
					</div>
					<div class="form-campos">
						<label>
							Token do Bot
							<input type="password" bind:value={telegramToken} placeholder="Ex: 123456:ABC-DEF..." />
						</label>
						<label>
							Chat ID do grupo/canal
							<input type="text" bind:value={telegramChatId} placeholder="Ex: -1001234567890" />
						</label>
					</div>
				{:else}
					<div class="instrucoes">
						<p>Configure o WhatsApp Business via Meta Cloud API:</p>
						<ol>
							<li>
								Acesse <a href="https://developers.facebook.com/apps" target="_blank" rel="noopener"
									>Meta for Developers</a
								>
							</li>
							<li>Crie um app do tipo "Business" → selecione "WhatsApp"</li>
							<li>Em WhatsApp → Configuração da API, copie o <strong>Phone Number ID</strong></li>
							<li>Gere um <strong>Access Token</strong> permanente (System User no Business Settings)</li>
							<li>Registre o número de telefone e configure os templates de mensagem</li>
						</ol>
						<p class="dica">
							💡 O token temporário (24h) serve para teste. Para produção, use um System User token permanente.
						</p>
					</div>
					<div class="form-campos">
						<label>
							Phone Number ID
							<input type="text" bind:value={whatsappPhoneId} placeholder="Ex: 1234567890123456" />
						</label>
						<label>
							Access Token (Meta)
							<input type="password" bind:value={whatsappToken} placeholder="Ex: EAAG..." />
						</label>
					</div>
				{/if}

				<div class="acoes-dupla">
					<Button onclick={() => salvarCanal(false)} disabled={salvandoCanal}>
						{salvandoCanal ? '⏳' : '💾'} Salvar {canalEscolhido === 'telegram' ? 'Telegram' : 'WhatsApp'}
					</Button>
					<Button variant="secondary" onclick={() => salvarCanal(true)} disabled={salvandoCanal}>Pular →</Button>
				</div>
			</div>
		{/if}

		<!-- Step 4: Validar -->
		{#if step >= 3 && !configurado}
			<div class="step-card">
				<h2>4. Validar Credenciais</h2>
				<p>Vamos testar suas credenciais fazendo uma chamada real à API da Shopee.</p>
				<Button onclick={validar} disabled={validando}>
					{validando ? '⏳ Validando...' : '🔍 Testar Credenciais'}
				</Button>
			</div>
		{/if}

		<!-- Exclusão de conta -->
		<div class="zona-perigo">
			<Button variant="danger" size="sm" onclick={() => (mostraExcluir = !mostraExcluir)}>
				🗑️ Excluir minha conta e dados
			</Button>
			{#if mostraExcluir}
				<div class="confirmar-exclusao">
					<p>
						⚠️ Esta ação é <strong>irreversível</strong>. Todos os seus dados (buscas, publicações, credenciais) serão
						removidos permanentemente.
					</p>
					<Button variant="danger" onclick={handleExcluir} disabled={excluindo}>
						{excluindo ? '⏳' : '🗑️'} Confirmar Exclusão
					</Button>
				</div>
			{/if}
		</div>
	{/if}
</section>

<style>
	.config-page {
		max-width: 700px;
	}
	h1 {
		font-size: 1.5rem;
		margin-bottom: 0.25rem;
	}
	h2 {
		font-size: 1.1rem;
		margin-bottom: 0.75rem;
	}
	.subtitulo {
		color: var(--tinta-suave);
		font-size: 0.9rem;
		margin-bottom: var(--r6);
	}
	.aviso {
		background: var(--porcelana);
		padding: var(--r4);
		border-radius: var(--raio-sm);
		color: var(--tinta-suave);
	}
	.loading {
		color: var(--tinta-suave);
		font-style: italic;
	}

	/* Progress bar */
	.progress {
		display: flex;
		gap: var(--r3);
		margin-bottom: var(--r6);
	}
	.progress-step {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 12px;
		border-radius: var(--raio-full);
		font-size: 0.82rem;
		font-weight: 600;
		background: var(--porcelana);
		color: var(--tinta-suave);
	}
	.progress-step.done {
		background: var(--sucesso-fundo);
		color: var(--sucesso-texto);
	}
	.progress-step.current {
		background: var(--ouro-fundo);
		color: var(--ouro-escuro);
	}
	.step-num {
		width: 20px;
		height: 20px;
		border-radius: 50%;
		background: currentcolor;
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.7rem;
	}
	.progress-step.done .step-num {
		background: var(--sucesso-texto);
	}
	.progress-step.current .step-num {
		background: var(--ouro);
	}

	/* Step cards */
	.step-card {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r5);
		margin-bottom: var(--r5);
	}
	.instrucoes {
		margin-bottom: var(--r4);
		font-size: 0.88rem;
		line-height: 1.6;
	}
	.instrucoes ol {
		padding-left: 1.2em;
	}
	.instrucoes li {
		margin-bottom: 6px;
	}
	.instrucoes a {
		color: var(--ouro);
		text-decoration: underline;
	}
	.instrucoes code {
		background: var(--porcelana);
		padding: 1px 5px;
		border-radius: 4px;
		font-size: 0.82rem;
	}

	.form-campos {
		display: flex;
		flex-direction: column;
		gap: var(--r3);
		margin-bottom: var(--r4);
	}
	.form-campos label {
		font-size: 0.82rem;
		font-weight: 600;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.form-campos input {
		padding: 10px 14px;
		border: 1px solid var(--linha);
		border-radius: 8px;
		font-size: 0.9rem;
	}
	.form-campos input:focus {
		outline: none;
		border-color: var(--ouro);
		box-shadow: 0 0 0 2px var(--ouro-fundo);
	}

	.acoes-dupla {
		display: flex;
		gap: var(--r3);
	}

	/* Canal tabs */
	.canal-tabs {
		display: flex;
		gap: var(--r2);
		margin-bottom: var(--r4);
	}
	.tab {
		padding: 8px 16px;
		border: 1px solid var(--linha);
		border-radius: 8px;
		background: var(--branco);
		font-size: 0.88rem;
		font-weight: 600;
		cursor: pointer;
		color: var(--tinta-suave);
	}
	.tab.ativo {
		background: var(--ouro-fundo);
		border-color: var(--ouro);
		color: var(--ouro-escuro);
	}
	.tab:hover:not(.ativo) {
		border-color: var(--tinta-suave);
	}
	.instrucoes-intro {
		font-size: 0.9rem;
		color: var(--tinta-suave);
		margin-bottom: var(--r3);
	}
	.dica {
		font-size: 0.82rem;
		color: var(--tinta-suave);
		background: var(--porcelana);
		padding: var(--r2) var(--r3);
		border-radius: 6px;
		margin-top: var(--r2);
	}

	/* OK panel */
	.painel-ok {
		background: var(--sucesso-fundo);
		border: 1px solid var(--sucesso-borda, var(--sucesso-texto));
		border-radius: var(--raio);
		padding: var(--r5);
		margin-bottom: var(--r5);
	}
	.painel-ok h2 {
		color: var(--sucesso-texto);
		margin-bottom: 4px;
	}
	.meta-info {
		font-size: 0.82rem;
		color: var(--tinta-suave);
		margin-top: 8px;
	}
	.meta-info code {
		background: var(--porcelana);
		padding: 1px 5px;
		border-radius: 4px;
	}

	/* Termos */
	.termos-box {
		background: var(--branco);
		border: 1px solid var(--linha);
		border-radius: 8px;
		padding: var(--r4);
		margin-bottom: var(--r4);
		font-size: 0.88rem;
		line-height: 1.6;
	}
	.termos-box ul {
		padding-left: 1.2em;
	}
	.termos-box li {
		margin-bottom: 6px;
	}

	/* Zona de perigo */
	.zona-perigo {
		margin-top: var(--r8);
		padding-top: var(--r5);
		border-top: 1px solid var(--linha);
	}
	.confirmar-exclusao {
		margin-top: var(--r3);
		background: var(--erro-fundo);
		border-radius: var(--raio-sm);
		padding: var(--r4);
	}
	.confirmar-exclusao p {
		font-size: var(--text-base);
		margin-bottom: var(--r3);
	}

	@media (max-width: 600px) {
		.progress {
			flex-wrap: wrap;
		}
		.acoes-dupla {
			flex-direction: column;
		}
	}
</style>
