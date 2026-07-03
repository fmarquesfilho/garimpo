<script>
	/**
	 * StepCanal — formulário de configuração de canal (Telegram/WhatsApp).
	 * Extraído de configurar/+page.svelte para respeitar limite de 400 linhas.
	 */
	import { Button } from '$lib/components/ui';

	let {
		canalEscolhido = $bindable('telegram'),
		telegramToken = $bindable(''),
		telegramChatId = $bindable(''),
		whatsappPhoneId = $bindable(''),
		whatsappToken = $bindable(''),
		salvandoCanal = false,
		onsalvar,
		onpular
	} = $props();
</script>

<div class="bg-nevoa border border-border rounded-md p-5 mb-5">
	<h2 class="text-lg mb-3">3. Canal de Publicação</h2>
	<p class="text-sm text-tinta-suave mb-3">Configure pelo menos um canal para publicar ofertas.</p>

	<div class="flex gap-2 mb-4">
		<button
			class="py-2 px-4 border rounded-lg text-sm font-semibold cursor-pointer {canalEscolhido === 'telegram'
				? 'bg-ouro-fundo border-ouro text-ouro-escuro'
				: 'bg-[var(--branco)] border-border text-tinta-suave hover:border-tinta-suave'}"
			onclick={() => (canalEscolhido = 'telegram')}
			type="button"
		>
			✈️ Telegram
		</button>
		<button
			class="py-2 px-4 border rounded-lg text-sm font-semibold cursor-pointer {canalEscolhido === 'whatsapp'
				? 'bg-ouro-fundo border-ouro text-ouro-escuro'
				: 'bg-[var(--branco)] border-border text-tinta-suave hover:border-tinta-suave'}"
			onclick={() => (canalEscolhido = 'whatsapp')}
			type="button"
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
					Envie <code class="bg-porcelana px-1 py-px rounded text-sm">/newbot</code> e siga as instruções
				</li>
				<li class="mb-1.5">Copie o <strong>Token</strong> fornecido</li>
				<li class="mb-1.5">Crie um grupo/canal, adicione o bot como admin</li>
				<li class="mb-1.5">
					Pegue o <strong>Chat ID</strong> (use
					<a href="https://t.me/getmyid_bot" target="_blank" rel="noopener" class="text-ouro underline">@getmyid_bot</a
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
				<li class="mb-1.5">Gere um <strong>Access Token</strong> permanente (System User)</li>
				<li class="mb-1.5">Registre o número de telefone e configure os templates</li>
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
		<Button onclick={onsalvar} disabled={salvandoCanal}>
			{salvandoCanal ? '⏳' : '💾'} Salvar {canalEscolhido === 'telegram' ? 'Telegram' : 'WhatsApp'}
		</Button>
		<Button variant="secondary" onclick={onpular} disabled={salvandoCanal}>Pular →</Button>
	</div>
</div>
