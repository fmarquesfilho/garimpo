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
		<nav class="nav-desktop">
			<a href="/" class:atual={$page.url.pathname === '/'}>Curadoria</a>
			<a href="/coletas" class:atual={$page.url.pathname === '/coletas'}>Coletas</a>
			<a href="/quadro" class:atual={$page.url.pathname === '/quadro'}>Quadro</a>
			<a href="/estatisticas" class:atual={$page.url.pathname === '/estatisticas'}>Estatísticas</a>
		</nav>
		<div class="auth-desktop">
			{#if $usuario}
				<span class="usuario-nome">{$usuario.nome ?? $usuario.email}</span>
				<button class="btn-auth" onclick={logout}>Sair</button>
			{:else}
				<button class="btn-auth" onclick={login}>Entrar</button>
			{/if}
		</div>
		<button
			class="hamburguer"
			onclick={() => (menuAberto = !menuAberto)}
			aria-label={menuAberto ? 'Fechar menu' : 'Abrir menu'}
			aria-expanded={menuAberto}
		>
			{#if menuAberto}✕{:else}☰{/if}
		</button>
	</div>
	{#if menuAberto}
		<nav class="nav-mobile">
			<a href="/" class:atual={$page.url.pathname === '/'} onclick={fecharMenu}>Curadoria</a>
			<a href="/coletas" class:atual={$page.url.pathname === '/coletas'} onclick={fecharMenu}>Coletas</a>
			<a href="/quadro" class:atual={$page.url.pathname === '/quadro'} onclick={fecharMenu}>Quadro</a>
			<a href="/estatisticas" class:atual={$page.url.pathname === '/estatisticas'} onclick={fecharMenu}>Estatísticas</a>
			{#if $usuario}
				<span class="usuario-mobile">{$usuario.nome ?? $usuario.email}</span>
				<button class="btn-auth" onclick={() => { logout(); fecharMenu(); }}>Sair</button>
			{:else}
				<button class="btn-auth" onclick={() => { login(); fecharMenu(); }}>Entrar com Google</button>
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
	.pingo {
		color: var(--ouro);
	}

	/* Desktop nav */
	.nav-desktop {
		display: flex;
		gap: var(--r6);
	}
	.nav-desktop a, .nav-mobile a {
		text-decoration: none;
		font-weight: 600;
		font-size: 0.92rem;
		color: var(--tinta-suave);
		padding-bottom: 2px;
		border-bottom: 2px solid transparent;
	}
	.nav-desktop a:hover, .nav-mobile a:hover {
		color: var(--tinta);
	}
	.nav-desktop a.atual, .nav-mobile a.atual {
		color: var(--tinta);
		border-color: var(--ouro);
	}

	/* Hamburger button - hidden on desktop */
	.hamburguer {
		display: none;
		border: none;
		background: transparent;
		font-size: 1.5rem;
		color: var(--tinta);
		cursor: pointer;
		padding: 4px 8px;
		line-height: 1;
	}

	/* Mobile nav */
	.nav-mobile {
		display: none;
		flex-direction: column;
		gap: var(--r4);
		padding: var(--r4) var(--r6) var(--r6);
		border-top: 1px solid var(--linha);
		background: color-mix(in srgb, var(--porcelana) 95%, white);
	}
	.nav-mobile a {
		font-size: 1.05rem;
		padding: var(--r2) 0;
	}

	/* Mobile breakpoint */
	@media (max-width: 540px) {
		.nav-desktop, .auth-desktop {
			display: none;
		}
		.hamburguer {
			display: block;
		}
		.nav-mobile {
			display: flex;
		}
	}

	/* Auth */
	.auth-desktop {
		display: flex; align-items: center; gap: var(--r3);
	}
	.usuario-nome {
		font-size: 0.82rem; color: var(--tinta-suave); font-weight: 500;
	}
	.usuario-mobile {
		font-size: 0.9rem; color: var(--tinta-suave); font-weight: 500;
		padding: var(--r2) 0;
	}
	.btn-auth {
		border: 1px solid var(--linha); background: var(--porcelana);
		color: var(--tinta); font-size: 0.8rem; font-weight: 600;
		padding: 6px 14px; border-radius: 999px; cursor: pointer;
	}
	.btn-auth:hover { border-color: var(--ouro); color: var(--ouro); }

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
