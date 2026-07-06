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
	<header class="sticky top-0 z-10 border-b border-border bg-porcelana/80 backdrop-blur-md">
		<div class="casca flex h-14 items-center justify-between">
			<a class="font-display text-2xl font-bold tracking-tight no-underline" href="/" onclick={fecharMenu}>
				Garimpei<span class="text-ouro">.</span>
			</a>
			{#if $usuario}
				<div class="flex items-center gap-3">
					<span class="hidden text-xs font-medium text-tinta-suave sm:inline">{$usuario.nome ?? $usuario.email}</span>
					<ThemeToggle />
					<Button variant="secondary" size="sm" onclick={logout}>Sair</Button>
					<button
						class="rounded-sm border border-border bg-porcelana px-3 py-1.5 text-sm leading-none text-tinta-suave hover:border-ouro hover:text-ouro"
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
		<main class="casca flex flex-col gap-8 min-h-[70vh] pt-8 pb-12">
			{@render children()}
		</main>
	{:else}
		<main class="casca flex min-h-[75vh] items-center justify-center pt-0">
			<LandingHero />
		</main>
	{/if}

	<footer class="casca flex items-center justify-between border-t border-border pt-6 pb-8 text-xs text-tinta-suave">
		<span>{hoje}</span>
		<span class="dado">teor = grau de ouro da pepita</span>
	</footer>
</Tooltip.Provider>
