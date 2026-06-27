<script>
	/**
	 * Landing page hero section (unauthenticated state).
	 */
	import { login } from '$lib/firebase.js';
</script>

<section class="hero">
	<h1>Garimpei<span class="pingo">.</span></h1>
	<p class="hero-sub">
		Curadoria inteligente de afiliados Shopee.<br>
		Encontre os melhores produtos, monitore lojas e publique com um clique.
	</p>
	<div class="hero-features">
		<div class="feature">
			<span class="feat-icon">🔍</span>
			<span>Busca e ranking por comissão, vendas e nota</span>
		</div>
		<div class="feature">
			<span class="feat-icon">🏪</span>
			<span>Monitoramento de lojas com alertas de novidades</span>
		</div>
		<div class="feature">
			<span class="feat-icon">📤</span>
			<span>Publicação com templates, fotos e agendamento</span>
		</div>
		<div class="feature">
			<span class="feat-icon">📊</span>
			<span>Rastreamento de conversões por destino</span>
		</div>
	</div>
	<button class="btn-entrar" onclick={login}>Entrar com Google</button>
</section>

<style>
	.hero { text-align: center; max-width: 520px; }
	.hero h1 {
		font-family: var(--display); font-size: clamp(2.5rem, 8vw, 4rem);
		font-weight: 700; letter-spacing: -0.03em; margin: 0 0 var(--r4);
	}
	.pingo { color: var(--ouro); }
	.hero-sub {
		font-size: 1.1rem; color: var(--tinta-suave); line-height: 1.6; margin: 0 0 var(--r8);
	}
	.hero-features {
		display: flex; flex-direction: column; gap: var(--r3);
		text-align: left; margin: 0 auto var(--r8); max-width: 380px;
	}
	.feature { display: flex; align-items: center; gap: var(--r3); font-size: 0.92rem; color: var(--tinta-suave); }
	.feat-icon { font-size: 1.3rem; flex-shrink: 0; }
	.btn-entrar {
		padding: 14px 36px; background: var(--ouro); color: white;
		font-weight: 700; font-size: 1rem; border: none;
		border-radius: var(--raio); cursor: pointer;
		box-shadow: 0 4px 12px rgba(184, 142, 58, 0.3);
	}
	.btn-entrar:hover {
		background: var(--ouro-hover);
		box-shadow: 0 6px 20px rgba(184, 142, 58, 0.4);
	}
</style>
