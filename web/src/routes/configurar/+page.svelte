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

<section class="max-w-[700px]">
	<h1 class="text-2xl mb-1">⚙️ Configurar Conta</h1>
	<p class="text-tinta-suave text-sm mb-6">
		Configure suas credenciais para usar o Garimpei com sua conta de afiliado.
	</p>

	{#if !$usuario}
		<div class="bg-porcelana p-4 rounded-sm text-tinta-suave">Faça login para configurar sua conta.</div>
	{:else if carregando}
		<p class="text-tinta-suave italic">Carregando...</p>
	{:else}
		{#if erro}
			<Alert variant="error">{erro}</Alert>
		{/if}
		{#if sucesso}
			<Alert variant="success">{sucesso}</Alert>
		{/if}

		<!-- Progress bar -->
		<div class="flex gap-3 mb-6 max-sm:flex-wrap">
			{#each [1, 2, 3, 4] as s}
				<div
					class="flex items-center gap-1.5 py-1.5 px-3 rounded-full text-sm font-semibold {step >= s
						? 'bg-sucesso-fundo text-sucesso'
						: step === s - 1
							? 'bg-ouro-fundo text-ouro-escuro'
							: 'bg-porcelana text-tinta-suave'}"
				>
					<span
						class="w-5 h-5 rounded-full flex items-center justify-center text-[0.7rem] text-white {step >= s
							? 'bg-sucesso'
							: step === s - 1
								? 'bg-ouro'
								: 'bg-tinta-suave'}">{s}</span
					>
					<span>
						{#if s === 1}Termos{:else if s === 2}Shopee{:else if s === 3}Telegram{:else}Validar{/if}
					</span>
				</div>
			{/each}
		</div>

		{#if configurado}
			<div class="bg-sucesso-fundo border border-sucesso-borda rounded-md p-5 mb-5">
				<h2 class="text-lg mb-1 text-sucesso">✅ Conta configurada!</h2>
				<p>Suas credenciais estão ativas. O sistema usará seus tokens pessoais para coletas e publicações.</p>
				<p class="text-sm text-tinta-suave mt-2">AppID: <code class="bg-porcelana px-1 py-px rounded">{appId}</code></p>
			</div>
		{/if}

		<!-- Step 1: Termos -->
		{#if step < 1}
			<div class="bg-nevoa border border-border rounded-md p-5 mb-5">
				<h2 class="text-lg mb-3">1. Termos de Uso e Privacidade</h2>
				<div class="bg-[var(--branco)] border border-border rounded-lg p-4 mb-4 text-sm leading-relaxed">
					<p><strong>Ao usar o Garimpei, você concorda que:</strong></p>
					<ul class="pl-5">
						<li class="mb-1.5">Seus dados pessoais (email, UID) são armazenados para funcionamento do serviço</li>
						<li class="mb-1.5">Suas credenciais de API são criptografadas e nunca compartilhadas</li>
						<li class="mb-1.5">Você pode solicitar exclusão completa dos seus dados a qualquer momento</li>
						<li class="mb-1.5">O serviço está em fase beta — sem garantias de disponibilidade</li>
					</ul>
				</div>
				<Button onclick={aceitarTermos}>Aceitar e Continuar</Button>
			</div>
		{/if}

		<!-- Step 2: Shopee -->
		{#if step >= 1 && step < 2}
			<div class="bg-nevoa border border-border rounded-md p-5 mb-5">
				<h2 class="text-lg mb-3">2. Credenciais Shopee Affiliate API</h2>
				<div class="mb-4 text-sm leading-relaxed">
					<p>Você precisa de um App ID e Secret do painel de afiliados da Shopee:</p>
					<ol class="pl-5">
						<li class="mb-1.5">
							Acesse <a
								href="https://affiliate.shopee.com.br/open_api"
								target="_blank"
								rel="noopener"
								class="text-ouro underline">affiliate.shopee.com.br/open_api</a
							>
						</li>
						<li class="mb-1.5">Faça login com sua conta de afiliado</li>
						<li class="mb-1.5">
							Na seção "Gerenciamento de API", copie o <strong>App ID</strong> e <strong>Secret</strong>
						</li>
						<li class="mb-1.5">Se não tiver acesso, solicite via Central de Ajuda (demora ~2 semanas)</li>
					</ol>
				</div>
				<div class="flex flex-col gap-3 mb-4">
					<label class="text-sm font-semibold flex flex-col gap-1">
						App ID
						<input
							type="text"
							class="py-2.5 px-3.5 border border-border rounded-lg text-sm focus:outline-none focus:border-ouro focus:shadow-[0_0_0_2px_var(--ouro-fundo)]"
							bind:value={appId}
							placeholder="Ex: 18332030606"
						/>
					</label>
					<label class="text-sm font-semibold flex flex-col gap-1">
						Secret
						<input
							type="password"
							class="py-2.5 px-3.5 border border-border rounded-lg text-sm focus:outline-none focus:border-ouro focus:shadow-[0_0_0_2px_var(--ouro-fundo)]"
							bind:value={secret}
							placeholder="Ex: MJS67QHU7HMCRX5..."
						/>
					</label>
				</div>
				<Button onclick={salvarShopee} disabled={salvandoShopee}>
					{salvandoShopee ? '⏳ Salvando...' : '💾 Salvar Credenciais'}
				</Button>
			</div>
		{/if}

		<!-- Step 3: Canal de publicação (Telegram ou WhatsApp) -->
		{#if step >= 2 && step < 3}
			<div class="bg-nevoa border border-border rounded-md p-5 mb-5">
				<h2 class="text-lg mb-3">3. Canal de Publicação</h2>
				<p class="text-sm text-tinta-suave mb-3">Configure pelo menos um canal para publicar ofertas.</p>

				<!-- Seletor de canal -->
				<div class="flex gap-2 mb-4">
					<button
						class="py-2 px-4 border rounded-lg text-sm font-semibold cursor-pointer {canalEscolhido === 'telegram'
							? 'bg-ouro-fundo border-ouro text-ouro-escuro'
							: 'bg-[var(--branco)] border-border text-tinta-suave hover:border-tinta-suave'}"
						onclick={() => (canalEscolhido = 'telegram')}
					>
						✈️ Telegram
					</button>
					<button
						class="py-2 px-4 border rounded-lg text-sm font-semibold cursor-pointer {canalEscolhido === 'whatsapp'
							? 'bg-ouro-fundo border-ouro text-ouro-escuro'
							: 'bg-[var(--branco)] border-border text-tinta-suave hover:border-tinta-suave'}"
						onclick={() => (canalEscolhido = 'whatsapp')}
					>
						💬 WhatsApp
					</button>
				</div>

				{#if canalEscolhido === 'telegram'}
					<div class="mb-4 text-sm leading-relaxed">
						<p>Configure um bot para publicar ofertas no Telegram:</p>
						<ol class="pl-5">
							<li class="mb-1.5">
								Abra o Telegram e converse com <a
									href="https://t.me/BotFather"
									target="_blank"
									rel="noopener"
									class="text-ouro underline">@BotFather</a
								>
							</li>
							<li class="mb-1.5">
								Envie <code class="bg-porcelana px-1 py-px rounded text-sm">/newbot</code> e siga as instruções para criar
								um bot
							</li>
							<li class="mb-1.5">Copie o <strong>Token</strong> fornecido</li>
							<li class="mb-1.5">Crie um grupo/canal, adicione o bot como admin</li>
							<li class="mb-1.5">
								Pegue o <strong>Chat ID</strong> (use
								<a href="https://t.me/getmyid_bot" target="_blank" rel="noopener" class="text-ouro underline"
									>@getmyid_bot</a
								>)
							</li>
						</ol>
					</div>
					<div class="flex flex-col gap-3 mb-4">
						<label class="text-sm font-semibold flex flex-col gap-1">
							Token do Bot
							<input
								type="password"
								class="py-2.5 px-3.5 border border-border rounded-lg text-sm focus:outline-none focus:border-ouro focus:shadow-[0_0_0_2px_var(--ouro-fundo)]"
								bind:value={telegramToken}
								placeholder="Ex: 123456:ABC-DEF..."
							/>
						</label>
						<label class="text-sm font-semibold flex flex-col gap-1">
							Chat ID do grupo/canal
							<input
								type="text"
								class="py-2.5 px-3.5 border border-border rounded-lg text-sm focus:outline-none focus:border-ouro focus:shadow-[0_0_0_2px_var(--ouro-fundo)]"
								bind:value={telegramChatId}
								placeholder="Ex: -1001234567890"
							/>
						</label>
					</div>
				{:else}
					<div class="mb-4 text-sm leading-relaxed">
						<p>Configure o WhatsApp Business via Meta Cloud API:</p>
						<ol class="pl-5">
							<li class="mb-1.5">
								Acesse <a
									href="https://developers.facebook.com/apps"
									target="_blank"
									rel="noopener"
									class="text-ouro underline">Meta for Developers</a
								>
							</li>
							<li class="mb-1.5">Crie um app do tipo "Business" → selecione "WhatsApp"</li>
							<li class="mb-1.5">Em WhatsApp → Configuração da API, copie o <strong>Phone Number ID</strong></li>
							<li class="mb-1.5">
								Gere um <strong>Access Token</strong> permanente (System User no Business Settings)
							</li>
							<li class="mb-1.5">Registre o número de telefone e configure os templates de mensagem</li>
						</ol>
						<p class="text-sm text-tinta-suave bg-porcelana py-2 px-3 rounded-md mt-2">
							💡 O token temporário (24h) serve para teste. Para produção, use um System User token permanente.
						</p>
					</div>
					<div class="flex flex-col gap-3 mb-4">
						<label class="text-sm font-semibold flex flex-col gap-1">
							Phone Number ID
							<input
								type="text"
								class="py-2.5 px-3.5 border border-border rounded-lg text-sm focus:outline-none focus:border-ouro focus:shadow-[0_0_0_2px_var(--ouro-fundo)]"
								bind:value={whatsappPhoneId}
								placeholder="Ex: 1234567890123456"
							/>
						</label>
						<label class="text-sm font-semibold flex flex-col gap-1">
							Access Token (Meta)
							<input
								type="password"
								class="py-2.5 px-3.5 border border-border rounded-lg text-sm focus:outline-none focus:border-ouro focus:shadow-[0_0_0_2px_var(--ouro-fundo)]"
								bind:value={whatsappToken}
								placeholder="Ex: EAAG..."
							/>
						</label>
					</div>
				{/if}

				<div class="flex gap-3 max-sm:flex-col">
					<Button onclick={() => salvarCanal(false)} disabled={salvandoCanal}>
						{salvandoCanal ? '⏳' : '💾'} Salvar {canalEscolhido === 'telegram' ? 'Telegram' : 'WhatsApp'}
					</Button>
					<Button variant="secondary" onclick={() => salvarCanal(true)} disabled={salvandoCanal}>Pular →</Button>
				</div>
			</div>
		{/if}

		<!-- Step 4: Validar -->
		{#if step >= 3 && !configurado}
			<div class="bg-nevoa border border-border rounded-md p-5 mb-5">
				<h2 class="text-lg mb-3">4. Validar Credenciais</h2>
				<p>Vamos testar suas credenciais fazendo uma chamada real à API da Shopee.</p>
				<Button onclick={validar} disabled={validando}>
					{validando ? '⏳ Validando...' : '🔍 Testar Credenciais'}
				</Button>
			</div>
		{/if}

		<!-- Exclusão de conta -->
		<div class="mt-8 pt-5 border-t border-border">
			<Button variant="danger" size="sm" onclick={() => (mostraExcluir = !mostraExcluir)}>
				🗑️ Excluir minha conta e dados
			</Button>
			{#if mostraExcluir}
				<div class="mt-3 bg-erro-fundo rounded-sm p-4">
					<p class="text-base mb-3">
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
