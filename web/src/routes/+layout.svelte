<script>
	import '../app.css';
	import { page } from '$app/stores';
	import { usuario, login, logout } from '$lib/firebase.js';
	let { children } = $props();

	let menuAberto = $state(false);

	const hoje = new Date().toLocaleDateString('pt-BR', {
		weekday: 'long',
		day: 'numeric',
		month: 'long'
	});

	function fecharMenu() {
		menuAberto = false;
	}
</script>

<header class="topo">
	<div class="casca barra">
		<a class="marca" href="/" onclick={fecharMenu}>
			Garimpo<span class="pingo">.</span>
		</a>
		<!-- Desktop: links principais condensados -->
		<nav class="nav-desktop">
			<a href="/" class:atual={$page.url.pathname === '/'}>Curadoria</a>
			<a href="/lojas" class:atual={$page.url.pathname === '/lojas'}>Lojas</a>
			<a href="/publicacoes" class:atual={$page.url.pathname === '/publicacoes'}>Publicações</a>
			<a href="/estatisticas" class:atual={$page.url.pathname === '/estatisticas'}>Dados</a>
		</nav>
		<div class="acoes-desktop">
			{#if $usuario}
				<span class="usuario-nome">{$usuario.nome ?? $usuario.email}</span>
				<button class="btn-auth" onclick={logout}>Sair</button>
			{:else}
				<button class="btn-auth" onclick={login}>Entrar</button>
			{/if}
			<button
				class="hamburguer-desktop"
				onclick={() => (menuAberto = !menuAberto)}
				aria-label={menuAberto ? 'Fechar menu' : 'Abrir menu'}
				aria-expanded={menuAberto}
			>☰</button>
		</div>
		<!-- Mobile: só hamburger -->
		<button
			class="hamburguer-mobile"
			onclick={() => (menuAberto = !menuAberto)}
			aria-label={menuAberto ? 'Fechar menu' : 'Abrir menu'}
			aria-expanded={menuAberto}
		>
			{#if menuAberto}✕{:else}☰{/if}
		</button>
	</div>
	{#if menuAberto}
		<nav class="nav-menu" onclick={fecharMenu}>
			<div class="menu-secao">
				<span class="menu-titulo">Principal</span>
				<a href="/" class:atual={$page.url.pathname === '/'}>🔍 Curadoria</a>
				<a href="/lojas" class:atual={$page.url.pathname === '/lojas'}>🏪 Lojas</a>
				<a href="/quadro" class:atual={$page.url.pathname === '/quadro'}>📋 Quadro</a>
			</div>
			<div class="menu-secao">
				<span class="menu-titulo">Publicar</span>
				<a href="/publicacoes" class:atual={$page.url.pathname === '/publicacoes'}>📤 Publicações</a>
				<a href="/canais" class:atual={$page.url.pathname === '/canais'}>📡 Destinos & Conversões</a>
			</div>
			<div class="menu-secao">
				<span class="menu-titulo">Monitoramento</span>
				<a href="/coletas" class:atual={$page.url.pathname === '/coletas'}>⏱ Coletas</a>
				<a href="/estatisticas" class:atual={$page.url.pathname === '/estatisticas'}>📊 Estatísticas</a>
			</div>
			{#if $usuario}
				<div class="menu-secao menu-auth">
					<span class="usuario-menu">{$usuario.nome ?? $usuario.email}</span>
					<button class="btn-auth" onclick={logout}>Sair</button>
				</div>
			{:else}
				<div class="menu-secao menu-auth">
					<button class="btn-auth" onclick={login}>Entrar com Google</button>
				</div>
			{/if}
		</nav>
	{/if}
</header>

<main class="casca">
	{@render children()}
</main>

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
		height: 64px;
	}
	.marca {
		font-family: var(--display);
		font-weight: 700;
		font-size: 1.5rem;
		text-decoration: none;
		letter-spacing: -0.02em;
	}
	.pingo { color: var(--ouro); }

	/* ── Desktop nav (links condensados) ─────────────────────────────── */
	.nav-desktop {
		display: flex;
		gap: var(--r5);
	}
	.nav-desktop a {
		text-decoration: none;
		font-weight: 600;
		font-size: 0.88rem;
		color: var(--tinta-suave);
		padding-bottom: 2px;
		border-bottom: 2px solid transparent;
	}
	.nav-desktop a:hover { color: var(--tinta); }
	.nav-desktop a.atual { color: var(--tinta); border-color: var(--ouro); }

	.acoes-desktop {
		display: flex;
		align-items: center;
		gap: var(--r3);
	}
	.usuario-nome {
		font-size: 0.82rem; color: var(--tinta-suave); font-weight: 500;
	}
	.hamburguer-desktop {
		border: 1px solid var(--linha);
		background: var(--porcelana);
		font-size: 1.1rem;
		color: var(--tinta-suave);
		cursor: pointer;
		padding: 4px 10px;
		border-radius: 8px;
		line-height: 1;
	}
	.hamburguer-desktop:hover { border-color: var(--ouro); color: var(--ouro); }

	/* ── Mobile hamburger ─────────────────────────────────────────────── */
	.hamburguer-mobile {
		display: none;
		border: none;
		background: transparent;
		font-size: 1.5rem;
		color: var(--tinta);
		cursor: pointer;
		padding: 4px 8px;
		line-height: 1;
	}

	/* ── Menu expansível (compartilhado desktop + mobile) ─────────────── */
	.nav-menu {
		display: flex;
		flex-direction: column;
		gap: 0;
		padding: var(--r4) var(--r6) var(--r6);
		border-top: 1px solid var(--linha);
		background: color-mix(in srgb, var(--porcelana) 95%, white);
		max-width: 320px;
		margin-left: auto;
	}
	.menu-secao {
		display: flex;
		flex-direction: column;
		gap: var(--r2);
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
		padding-bottom: 2px;
	}
	.nav-menu a {
		text-decoration: none;
		font-weight: 600;
		font-size: 0.95rem;
		color: var(--tinta-suave);
		padding: 6px 0;
	}
	.nav-menu a:hover { color: var(--tinta); }
	.nav-menu a.atual { color: var(--tinta); }
	.menu-auth {
		display: flex;
		flex-direction: column;
		gap: var(--r2);
		padding-top: var(--r3);
	}
	.usuario-menu {
		font-size: 0.85rem; color: var(--tinta-suave); font-weight: 500;
	}

	/* ── Auth button ──────────────────────────────────────────────────── */
	.btn-auth {
		border: 1px solid var(--linha); background: var(--porcelana);
		color: var(--tinta); font-size: 0.8rem; font-weight: 600;
		padding: 6px 14px; border-radius: 999px; cursor: pointer;
		width: fit-content;
	}
	.btn-auth:hover { border-color: var(--ouro); color: var(--ouro); }

	/* ── Breakpoints ──────────────────────────────────────────────────── */
	@media (max-width: 680px) {
		.nav-desktop, .acoes-desktop {
			display: none;
		}
		.hamburguer-mobile {
			display: block;
		}
		.nav-menu {
			max-width: none;
		}
	}

	@media (min-width: 681px) {
		.hamburguer-mobile { display: none; }
	}

	/* ── Main & footer ────────────────────────────────────────────────── */
	main {
		min-height: 70vh;
		padding-top: var(--r8);
		padding-bottom: var(--r12);
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
</style>
