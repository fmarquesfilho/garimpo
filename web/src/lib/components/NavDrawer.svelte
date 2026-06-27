<script>
	/**
	 * Drawer lateral de navegação. Slide-in mobile-first.
	 */
	import { page } from '$app/stores';
	import { logout } from '$lib/firebase.js';

	let { aberto = false, usuario = null, isAdmin = false, onfechar = null } = $props();
</script>

{#if aberto && usuario}
	<!-- Backdrop -->
	<div class="backdrop" onclick={onfechar} aria-hidden="true"></div>
	<!-- Menu slide -->
	<nav class="drawer" aria-label="Menu de navegação">
		<div class="drawer-header">
			<span class="drawer-titulo">Menu</span>
			<button class="drawer-fechar" onclick={onfechar} aria-label="Fechar menu">✕</button>
		</div>
		<div class="drawer-conteudo">
			<div class="menu-secao">
				<span class="menu-titulo">Principal</span>
				<a href="/" class:atual={$page.url.pathname === '/'}>🔍 Descobrir</a>
			</div>
			<div class="menu-secao">
				<span class="menu-titulo">Publicar</span>
				<a href="/publicar" class:atual={$page.url.pathname === '/publicar'}>🔗 Link</a>
			</div>
			<div class="menu-secao">
				<span class="menu-titulo">Monitoramento</span>
				<a href="/estatisticas" class:atual={$page.url.pathname === '/estatisticas'}>📊 Estatísticas</a>
				<a href="/publicacoes" class:atual={$page.url.pathname === '/publicacoes'}>📤 Publicações</a>
				{#if isAdmin}
					<a href="/coletas" class:atual={$page.url.pathname === '/coletas'}>⏱ Coletas</a>
					<a href="/admin" class:atual={$page.url.pathname === '/admin'}>🛠 Admin</a>
				{/if}
			</div>
			<div class="menu-secao">
				<span class="menu-titulo">Configurações</span>
				<a href="/lojas" class:atual={$page.url.pathname === '/lojas'}>🏪 Lojas</a>
				<a href="/canais" class:atual={$page.url.pathname === '/canais'}>📡 Destinos</a>
				<a href="/configurar" class:atual={$page.url.pathname === '/configurar'}>⚙️ Conta</a>
			</div>
		</div>
		<div class="drawer-footer">
			<span class="drawer-user">{usuario.nome ?? usuario.email}</span>
			<button class="btn-auth" onclick={logout}>Sair</button>
		</div>
	</nav>
{/if}

<style>
	.backdrop {
		position: fixed; inset: 0;
		background: rgba(43, 29, 46, 0.4);
		z-index: 90; animation: fadeIn 0.2s ease;
	}
	.drawer {
		position: fixed; top: 0; right: 0; bottom: 0;
		width: 280px; max-width: 85vw;
		background: var(--nevoa); z-index: 100;
		display: flex; flex-direction: column;
		box-shadow: -4px 0 24px rgba(43, 29, 46, 0.15);
		animation: slideIn 0.25s ease; overflow: hidden;
	}
	.drawer-header {
		display: flex; align-items: center; justify-content: space-between;
		padding: var(--r4) var(--r5); border-bottom: 1px solid var(--linha); flex-shrink: 0;
	}
	.drawer-titulo { font-family: var(--display); font-weight: 600; font-size: 1.1rem; }
	.drawer-fechar {
		border: none; background: transparent; font-size: 1.2rem;
		color: var(--tinta-suave); cursor: pointer; padding: 4px 8px; border-radius: 6px;
	}
	.drawer-fechar:hover { background: var(--porcelana); color: var(--tinta); }
	.drawer-conteudo {
		flex: 1; overflow-y: auto; padding: var(--r3) var(--r5);
		-webkit-overflow-scrolling: touch;
	}
	.drawer-footer {
		border-top: 1px solid var(--linha); padding: var(--r4) var(--r5);
		display: flex; align-items: center; justify-content: space-between; flex-shrink: 0;
	}
	.drawer-user {
		font-size: 0.78rem; color: var(--tinta-suave);
		overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 150px;
	}
	.menu-secao {
		display: flex; flex-direction: column; gap: 2px;
		padding: var(--r3) 0; border-bottom: 1px solid var(--linha);
	}
	.menu-secao:last-child { border-bottom: none; }
	.menu-titulo {
		font-size: 0.68rem; font-weight: 700; text-transform: uppercase;
		letter-spacing: 0.06em; color: var(--tinta-suave); opacity: 0.7; margin-bottom: var(--r2);
	}
	.drawer a {
		text-decoration: none; font-weight: 600; font-size: 0.92rem;
		color: var(--tinta-suave); padding: 10px 12px; border-radius: 8px; display: block;
	}
	.drawer a:hover { color: var(--tinta); background: var(--porcelana); }
	.drawer a.atual { color: var(--tinta); background: var(--ouro-fundo); }
	.btn-auth {
		border: 1px solid var(--linha); background: var(--porcelana);
		color: var(--tinta); font-size: 0.8rem; font-weight: 600;
		padding: 6px 14px; border-radius: var(--raio-full); cursor: pointer; width: fit-content;
	}
	.btn-auth:hover { border-color: var(--ouro); color: var(--ouro); }

	@keyframes slideIn { from { transform: translateX(100%); } to { transform: translateX(0); } }
	@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }

	@media (min-width: 768px) { .drawer { width: 320px; } }
	@media (prefers-reduced-motion: reduce) { .drawer, .backdrop { animation: none; } }
</style>
