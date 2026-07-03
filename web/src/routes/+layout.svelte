<script>
	import '../app.css';
	import { page } from '$app/stores';
	import { usuario, logout } from '$lib/firebase.js';
	import { verificarAdmin } from '$lib/api.js';
	import { initTheme } from '$lib/theme.js';
	import NavDrawer from '$lib/components/NavDrawer.svelte';
	import LandingHero from '$lib/components/LandingHero.svelte';
	import { Button, ThemeToggle } from '$lib/components/ui';
	import { Tooltip } from 'bits-ui';
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

	// Inicializa theme engine e remove classe no-transitions após primeiro paint
	onMount(() => {
		const cleanup = initTheme();
		requestAnimationFrame(() => {
			document.documentElement.classList.remove('no-transitions');
		});
		return cleanup;
	});

	// Fecha menu ao navegar
	$effect(() => {
		$page.url.pathname;
		menuAberto = false;
	});

	// Verifica role quando o usuário loga
	$effect(() => {
		if ($usuario) {
			verificarAdmin()
				.then((r) => {
					isAdmin = r?.admin ?? false;
				})
				.catch(() => {
					isAdmin = false;
				});
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

<Tooltip.Provider>
	<header class="topo">
		<div class="casca barra">
			<a class="marca" href="/" onclick={fecharMenu}>
				Garimpei<span class="pingo">.</span>
			</a>
			{#if $usuario}
				<div class="acoes-barra">
					<span class="usuario-nome">{$usuario.nome ?? $usuario.email}</span>
					<ThemeToggle />
					<Button variant="secondary" size="sm" onclick={logout}>Sair</Button>
					<button
						class="hamburguer"
						onclick={toggleMenu}
						aria-label={menuAberto ? 'Fechar menu' : 'Abrir menu'}
						aria-expanded={menuAberto}
						>{#if menuAberto}✕{:else}☰{/if}</button
					>
				</div>
			{/if}
		</div>
	</header>

	<NavDrawer aberto={menuAberto} usuario={$usuario} {isAdmin} onfechar={fecharMenu} />

	{#if $usuario}
		<main class="casca">
			{@render children()}
		</main>
	{:else}
		<main class="casca landing">
			<LandingHero />
		</main>
	{/if}

	<footer class="rodape casca">
		<span>{hoje}</span>
		<span class="dado">teor = grau de ouro da pepita</span>
	</footer>
</Tooltip.Provider>

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
	.pingo {
		color: var(--ouro);
	}
	.acoes-barra {
		display: flex;
		align-items: center;
		gap: var(--r3);
	}
	.usuario-nome {
		font-size: 0.82rem;
		color: var(--tinta-suave);
		font-weight: 500;
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
	.hamburguer:hover {
		border-color: var(--ouro);
		color: var(--ouro);
	}

	@media (max-width: 480px) {
		.usuario-nome {
			display: none;
		}
	}

	main {
		min-height: 70vh;
		padding-top: var(--r8);
		padding-bottom: var(--r12);
	}
	.landing {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 75vh;
		padding-top: 0;
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
