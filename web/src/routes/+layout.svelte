<script>
	import '../app.css';
	import { page } from '$app/stores';
	import { usuario, login, logout } from '$lib/firebase.js';
	import { verificarAdmin } from '$lib/api.js';
	import { onMount } from 'svelte';
	let { children } = $props();

	let menuAberto = $state(false);
	let isAdmin = $state(false);

	const hoje = new Date().toLocaleDateString('pt-BR', {
		weekday: 'long',
		day: 'numeric',
		month: 'long'
	});

	function fecharMenu() {
		menuAberto = false;
	}

	function toggleMenu() {
		menuAberto = !menuAberto;
	}

	// Fecha menu ao navegar
	$effect(() => {
		$page.url.pathname;
		menuAberto = false;
	});

	// Verifica role quando o usuário loga
	$effect(() => {
		if ($usuario) {
			verificarAdmin().then(r => { isAdmin = r?.admin ?? false; }).catch(() => { isAdmin = false; });
		} else {
			isAdmin = false;
		}
	});

	// Trava scroll do body quando menu está aberto (mobile)
	$effect(() => {
		if (typeof document !== 'undefined') {
			document.body.style.overflow = menuAberto ? 'hidden' : '';
		}
	});
</script>

<header class="topo">
	<div class="casca barra">
		<a class="marca" href="/" onclick={fecharMenu}>
			Garimpei<span class="pingo">.</span>
		</a>
		{#if $usuario}
			<div class="acoes-barra">
				<span class="usuario-nome">{$usuario.nome ?? $usuario.email}</span>
				<button class="btn-auth" onclick={logout}>Sair</button>
				<button
					class="hamburguer"
					onclick={toggleMenu}
					aria-label={menuAberto ? 'Fechar menu' : 'Abrir menu'}
					aria-expanded={menuAberto}
				>{#if menuAberto}✕{:else}☰{/if}</button>
			</div>
		{/if}
	</div>
</header>

<!-- Drawer lateral -->
{#if menuAberto && $usuario}
	<!-- Backdrop -->
	<div class="backdrop" onclick={fecharMenu} aria-hidden="true"></div>
	<!-- Menu slide -->
	<nav class="drawer" aria-label="Menu de navegação">
		<div class="drawer-header">
			<span class="drawer-titulo">Menu</span>
			<button class="drawer-fechar" onclick={fecharMenu} aria-label="Fechar menu">✕</button>
		</div>
		<div class="drawer-conteudo">
			<div class="menu-secao">
				<span class="menu-titulo">Principal</span>
				<a href="/" class:atual={$page.url.pathname === '/'}>🔍 Buscar</a>
				<a href="/oportunidades" class:atual={$page.url.pathname === '/oportunidades'}>🎯 Oportunidades</a>
			</div>
			<div class="menu-secao">
				<span class="menu-titulo">Publicar</span>
				<a href="/publicar" class:atual={$page.url.pathname === '/publicar'}>🔗 Link</a>
				<a href="/publicacoes" class:atual={$page.url.pathname === '/publicacoes'}>📤 Publicações</a>
			</div>
			<div class="menu-secao">
				<span class="menu-titulo">Monitoramento</span>
				<a href="/coletas" class:atual={$page.url.pathname === '/coletas'}>⏱ Coletas</a>
				<a href="/estatisticas" class:atual={$page.url.pathname === '/estatisticas'}>📊 Estatísticas</a>
				{#if isAdmin}
					<a href="/admin" class:atual={$page.url.pathname === '/admin'}>🛠 Admin</a>
				{/if}
			</div>
			<div class="menu-secao">
				<span class="menu-titulo">Configurações</span>
				<a href="/lojas" class:atual={$page.url.pathname === '/lojas'}>🏪 Lojas</a>
				<a href="/canais" class:atual={$page.url.pathname === '/canais'}>📡 Destinos</a>
			</div>
		</div>
		<div class="drawer-footer">
			<span class="drawer-user">{$usuario.nome ?? $usuario.email}</span>
			<button class="btn-auth" onclick={logout}>Sair</button>
		</div>
	</nav>
{/if}

{#if $usuario}
	<main class="casca">
		{@render children()}
	</main>
{:else}
	<main class="casca landing">
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
	</main>
{/if}

<footer class="rodape casca">
	<span>{hoje}</span>
	<span class="dado">teor = grau de ouro da pepita</span>
</footer>

<style>
	.topo {
		border-bottom: 1px solid var(--linha);
		background: color-mix(in srgb, var(--porcelana) 80%, white);
		position: sticky;
		top: 0;
		z-index: 10;
		backdrop-filter: blur(8px);
	}
	.barra {
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: 56px;
	}
	.marca {
		font-family: var(--display);
		font-weight: 700;
		font-size: 1.5rem;
		text-decoration: none;
		letter-spacing: -0.02em;
	}
	.pingo { color: var(--ouro); }

	.acoes-barra {
		display: flex;
		align-items: center;
		gap: var(--r3);
	}
	.usuario-nome {
		font-size: 0.82rem; color: var(--tinta-suave); font-weight: 500;
	}
	.hamburguer {
		border: 1px solid var(--linha);
		background: var(--porcelana);
		font-size: 1.1rem;
		color: var(--tinta-suave);
		cursor: pointer;
		padding: 6px 12px;
		border-radius: 8px;
		line-height: 1;
	}
	.hamburguer:hover { border-color: var(--ouro); color: var(--ouro); }

	/* ── Backdrop ─────────────────────────────────────────────────────── */
	.backdrop {
		position: fixed;
		inset: 0;
		background: rgba(43, 29, 46, 0.4);
		z-index: 90;
		animation: fadeIn 0.2s ease;
	}

	/* ── Drawer (menu lateral) ────────────────────────────────────────── */
	.drawer {
		position: fixed;
		top: 0;
		right: 0;
		bottom: 0;
		width: 280px;
		max-width: 85vw;
		background: var(--nevoa);
		z-index: 100;
		display: flex;
		flex-direction: column;
		box-shadow: -4px 0 24px rgba(43, 29, 46, 0.15);
		animation: slideIn 0.25s ease;
		overflow: hidden;
	}

	.drawer-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--r4) var(--r5);
		border-bottom: 1px solid var(--linha);
		flex-shrink: 0;
	}
	.drawer-titulo {
		font-family: var(--display);
		font-weight: 600;
		font-size: 1.1rem;
	}
	.drawer-fechar {
		border: none;
		background: transparent;
		font-size: 1.2rem;
		color: var(--tinta-suave);
		cursor: pointer;
		padding: 4px 8px;
		border-radius: 6px;
	}
	.drawer-fechar:hover { background: var(--porcelana); color: var(--tinta); }

	.drawer-conteudo {
		flex: 1;
		overflow-y: auto;
		padding: var(--r3) var(--r5);
		-webkit-overflow-scrolling: touch;
	}

	.drawer-footer {
		border-top: 1px solid var(--linha);
		padding: var(--r4) var(--r5);
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-shrink: 0;
	}
	.drawer-user {
		font-size: 0.78rem;
		color: var(--tinta-suave);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		max-width: 150px;
	}

	/* ── Seções do menu ───────────────────────────────────────────────── */
	.menu-secao {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: var(--r3) 0;
		border-bottom: 1px solid var(--linha);
	}
	.menu-secao:last-child { border-bottom: none; }
	.menu-titulo {
		font-size: 0.68rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--tinta-suave);
		opacity: 0.7;
		margin-bottom: var(--r2);
	}
	.drawer a {
		text-decoration: none;
		font-weight: 600;
		font-size: 0.92rem;
		color: var(--tinta-suave);
		padding: 10px 12px;
		border-radius: 8px;
		display: block;
	}
	.drawer a:hover { color: var(--tinta); background: var(--porcelana); }
	.drawer a.atual { color: var(--tinta); background: var(--ouro-fundo); }

	/* ── Animações ────────────────────────────────────────────────────── */
	@keyframes slideIn {
		from { transform: translateX(100%); }
		to { transform: translateX(0); }
	}
	@keyframes fadeIn {
		from { opacity: 0; }
		to { opacity: 1; }
	}

	/* ── Auth button ──────────────────────────────────────────────────── */
	.btn-auth {
		border: 1px solid var(--linha); background: var(--porcelana);
		color: var(--tinta); font-size: 0.8rem; font-weight: 600;
		padding: 6px 14px; border-radius: var(--raio-full); cursor: pointer;
		width: fit-content;
	}
	.btn-auth:hover { border-color: var(--ouro); color: var(--ouro); }

	/* ── Mobile: esconde nome do usuário na barra ─────────────────────── */
	@media (max-width: 480px) {
		.usuario-nome { display: none; }
	}

	/* ── Desktop: drawer mais largo ───────────────────────────────────── */
	@media (min-width: 768px) {
		.drawer { width: 320px; }
	}

	/* ── Main & footer ────────────────────────────────────────────────── */
	main {
		min-height: 70vh;
		padding-top: var(--r8);
		padding-bottom: var(--r12);
	}

	/* ── Landing page ─────────────────────────────────────────────────── */
	.landing {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 75vh;
		padding-top: 0;
	}
	.hero {
		text-align: center;
		max-width: 520px;
	}
	.hero h1 {
		font-family: var(--display);
		font-size: clamp(2.5rem, 8vw, 4rem);
		font-weight: 700;
		letter-spacing: -0.03em;
		margin: 0 0 var(--r4);
	}
	.hero-sub {
		font-size: 1.1rem;
		color: var(--tinta-suave);
		line-height: 1.6;
		margin: 0 0 var(--r8);
	}
	.hero-features {
		display: flex;
		flex-direction: column;
		gap: var(--r3);
		text-align: left;
		margin: 0 auto var(--r8);
		max-width: 380px;
	}
	.feature {
		display: flex;
		align-items: center;
		gap: var(--r3);
		font-size: 0.92rem;
		color: var(--tinta-suave);
	}
	.feat-icon { font-size: 1.3rem; flex-shrink: 0; }
	.btn-entrar {
		padding: 14px 36px;
		background: var(--ouro);
		color: white;
		font-weight: 700;
		font-size: 1rem;
		border: none;
		border-radius: var(--raio);
		cursor: pointer;
		box-shadow: 0 4px 12px rgba(184, 142, 58, 0.3);
	}
	.btn-entrar:hover {
		background: var(--ouro-hover);
		box-shadow: 0 6px 20px rgba(184, 142, 58, 0.4);
	}

	.rodape {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-top: var(--r6);
		padding-bottom: var(--r8);
		border-top: 1px solid var(--linha);
		font-size: 0.8rem;
		color: var(--tinta-suave);
	}

	/* ── Reduced motion ───────────────────────────────────────────────── */
	@media (prefers-reduced-motion: reduce) {
		.drawer, .backdrop {
			animation: none;
		}
	}
</style>
