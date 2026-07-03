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
	<div class="backdrop" onclick={onfechar} aria-hidden="true"></div>
	<!-- Menu slide -->
	<nav class="drawer" aria-label="Menu de navegação">
		<div class="drawer-header">
			<span class="drawer-titulo">Menu</span>
			<Button variant="ghost" size="sm" onclick={onfechar} aria-label="Fechar menu">✕</Button>
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
					<a href="/docs/" target="_blank">📚 Docs</a>
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
			<Button variant="secondary" size="sm" onclick={logout}>Sair</Button>
		</div>
	</nav>
{/if}

<style>
	.backdrop {
		position: fixed;
		inset: 0;
		background: rgba(46, 34, 38, 0.4);
		z-index: 90;
		animation: fadeIn 0.2s ease;
	}
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
		box-shadow: var(--sombra);
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
		font-weight: var(--font-semi);
		font-size: var(--text-lg);
	}
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
		font-size: var(--text-xs);
		color: var(--tinta-suave);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		max-width: 150px;
	}
	.menu-secao {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: var(--r3) 0;
		border-bottom: 1px solid var(--linha);
	}
	.menu-secao:last-child { border-bottom: none; }
	.menu-titulo {
		font-size: var(--text-xs);
		font-weight: var(--font-bold);
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--tinta-suave);
		opacity: 0.7;
		margin-bottom: var(--r2);
	}
	.drawer a {
		text-decoration: none;
		font-weight: var(--font-semi);
		font-size: var(--text-base);
		color: var(--tinta-suave);
		padding: var(--r3) var(--r3);
		border-radius: var(--raio-sm);
		display: block;
	}
	.drawer a:hover {
		color: var(--tinta);
		background: var(--porcelana);
	}
	.drawer a.atual {
		color: var(--tinta);
		background: var(--ouro-fundo);
	}

	@keyframes slideIn { from { transform: translateX(100%); } to { transform: translateX(0); } }
	@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }

	@media (min-width: 768px) { .drawer { width: 320px; } }
	@media (prefers-reduced-motion: reduce) { .drawer, .backdrop { animation: none; } }
</style>
