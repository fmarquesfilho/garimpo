<script>
	/**
	 * Landing page hero section (unauthenticated state).
	 */
	import { login } from '$lib/firebase.js';
</script>

<section class="max-w-[520px] text-center">
	<h1 class="m-0 mb-4 font-display text-[clamp(2.5rem,8vw,4rem)] font-bold tracking-tight">
		Garimpei<span class="text-ouro">.</span>
	</h1>
	<p class="m-0 mb-8 text-lg leading-relaxed text-tinta-suave">
		Curadoria inteligente de afiliados Shopee.<br />
		Encontre os melhores produtos, monitore lojas e publique com um clique.
	</p>
	<div class="mx-auto mb-8 flex max-w-[380px] flex-col gap-3 text-left">
		<div class="flex items-center gap-3 text-[0.92rem] text-tinta-suave">
			<span class="shrink-0 text-xl">🔍</span>
			<span>Busca e ranking por comissão, vendas e nota</span>
		</div>
		<div class="flex items-center gap-3 text-[0.92rem] text-tinta-suave">
			<span class="shrink-0 text-xl">🏪</span>
			<span>Monitoramento de lojas com alertas de novidades</span>
		</div>
		<div class="flex items-center gap-3 text-[0.92rem] text-tinta-suave">
			<span class="shrink-0 text-xl">📤</span>
			<span>Publicação com templates, fotos e agendamento</span>
		</div>
		<div class="flex items-center gap-3 text-[0.92rem] text-tinta-suave">
			<span class="shrink-0 text-xl">📊</span>
			<span>Rastreamento de conversões por destino</span>
		</div>
	</div>
	<button
		class="cursor-pointer rounded-md border-none bg-[var(--ouro)] px-9 py-3.5 text-base font-bold text-white shadow-[0_4px_12px_rgba(184,142,58,0.3)] hover:bg-[var(--ouro-hover)] hover:shadow-[0_6px_20px_rgba(184,142,58,0.4)]"
		onclick={login}>Entrar com Google</button
	>
</section>
