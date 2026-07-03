<script>
	/**
	 * Drawer lateral de navegação. Slide-in mobile-first.
	 */
	import { page } from '$app/stores';
	import { logout } from '$lib/firebase.js';
	import { Button } from '$lib/components/ui';

	let { aberto = false, usuario = null, isAdmin = false, onfechar = null } = $props();
</script>

{#if aberto && usuario}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 z-90 animate-[fadeIn_0.2s_ease] bg-[rgba(0,0,0,0.5)] motion-reduce:animate-none"
		onclick={onfechar}
		aria-hidden="true"
	></div>
	<!-- Menu slide -->
	<nav
		class="fixed top-0 right-0 bottom-0 z-100 flex w-[280px] max-w-[85vw] animate-[slideIn_0.25s_ease] flex-col overflow-hidden bg-card shadow-sm motion-reduce:animate-none md:w-[320px]"
		aria-label="Menu de navegação"
	>
		<div class="flex shrink-0 items-center justify-between border-b border-border px-5 py-4">
			<span class="font-display text-lg font-semibold">Menu</span>
			<Button variant="ghost" size="sm" onclick={onfechar} aria-label="Fechar menu">✕</Button>
		</div>
		<div class="flex-1 overflow-y-auto px-5 py-3 [-webkit-overflow-scrolling:touch]">
			<div class="flex flex-col gap-0.5 border-b border-border py-3 last:border-b-0">
				<span class="mb-2 text-xs font-bold uppercase tracking-wide text-tinta-suave opacity-70">Principal</span>
				<a
					href="/"
					class="block rounded-sm px-3 py-3 font-semibold text-tinta-suave no-underline hover:bg-porcelana hover:text-foreground"
					class:!bg-ouro-fundo={$page.url.pathname === '/'}
					class:!text-foreground={$page.url.pathname === '/'}>🔍 Descobrir</a
				>
			</div>
			<div class="flex flex-col gap-0.5 border-b border-border py-3 last:border-b-0">
				<span class="mb-2 text-xs font-bold uppercase tracking-wide text-tinta-suave opacity-70">Publicar</span>
				<a
					href="/publicar"
					class="block rounded-sm px-3 py-3 font-semibold text-tinta-suave no-underline hover:bg-porcelana hover:text-foreground"
					class:!bg-ouro-fundo={$page.url.pathname === '/publicar'}
					class:!text-foreground={$page.url.pathname === '/publicar'}>🔗 Link</a
				>
			</div>
			<div class="flex flex-col gap-0.5 border-b border-border py-3 last:border-b-0">
				<span class="mb-2 text-xs font-bold uppercase tracking-wide text-tinta-suave opacity-70">Monitoramento</span>
				<a
					href="/estatisticas"
					class="block rounded-sm px-3 py-3 font-semibold text-tinta-suave no-underline hover:bg-porcelana hover:text-foreground"
					class:!bg-ouro-fundo={$page.url.pathname === '/estatisticas'}
					class:!text-foreground={$page.url.pathname === '/estatisticas'}>📊 Estatísticas</a
				>
				<a
					href="/publicacoes"
					class="block rounded-sm px-3 py-3 font-semibold text-tinta-suave no-underline hover:bg-porcelana hover:text-foreground"
					class:!bg-ouro-fundo={$page.url.pathname === '/publicacoes'}
					class:!text-foreground={$page.url.pathname === '/publicacoes'}>📤 Publicações</a
				>
				{#if isAdmin}
					<a
						href="/coletas"
						class="block rounded-sm px-3 py-3 font-semibold text-tinta-suave no-underline hover:bg-porcelana hover:text-foreground"
						class:!bg-ouro-fundo={$page.url.pathname === '/coletas'}
						class:!text-foreground={$page.url.pathname === '/coletas'}>⏱ Coletas</a
					>
					<a
						href="/admin"
						class="block rounded-sm px-3 py-3 font-semibold text-tinta-suave no-underline hover:bg-porcelana hover:text-foreground"
						class:!bg-ouro-fundo={$page.url.pathname === '/admin'}
						class:!text-foreground={$page.url.pathname === '/admin'}>🛠 Admin</a
					>
					<a
						href="/docs/"
						target="_blank"
						class="block rounded-sm px-3 py-3 font-semibold text-tinta-suave no-underline hover:bg-porcelana hover:text-foreground"
						>📚 Docs</a
					>
				{/if}
			</div>
			<div class="flex flex-col gap-0.5 border-b border-border py-3 last:border-b-0">
				<span class="mb-2 text-xs font-bold uppercase tracking-wide text-tinta-suave opacity-70">Configurações</span>
				<a
					href="/lojas"
					class="block rounded-sm px-3 py-3 font-semibold text-tinta-suave no-underline hover:bg-porcelana hover:text-foreground"
					class:!bg-ouro-fundo={$page.url.pathname === '/lojas'}
					class:!text-foreground={$page.url.pathname === '/lojas'}>🏪 Lojas</a
				>
				<a
					href="/canais"
					class="block rounded-sm px-3 py-3 font-semibold text-tinta-suave no-underline hover:bg-porcelana hover:text-foreground"
					class:!bg-ouro-fundo={$page.url.pathname === '/canais'}
					class:!text-foreground={$page.url.pathname === '/canais'}>📡 Destinos</a
				>
				<a
					href="/configurar"
					class="block rounded-sm px-3 py-3 font-semibold text-tinta-suave no-underline hover:bg-porcelana hover:text-foreground"
					class:!bg-ouro-fundo={$page.url.pathname === '/configurar'}
					class:!text-foreground={$page.url.pathname === '/configurar'}>⚙️ Conta</a
				>
			</div>
		</div>
		<div class="flex shrink-0 items-center justify-between border-t border-border px-5 py-4">
			<span class="max-w-[150px] truncate text-xs text-tinta-suave">{usuario.nome ?? usuario.email}</span>
			<Button variant="secondary" size="sm" onclick={logout}>Sair</Button>
		</div>
	</nav>
{/if}

<style>
	@keyframes slideIn {
		from {
			transform: translateX(100%);
		}
		to {
			transform: translateX(0);
		}
	}
	@keyframes fadeIn {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}
</style>
