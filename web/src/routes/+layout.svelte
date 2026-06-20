<script>
	import '../app.css';
	import { page } from '$app/stores';
	let { children } = $props();

	const hoje = new Date().toLocaleDateString('pt-BR', {
		weekday: 'long',
		day: 'numeric',
		month: 'long'
	});
</script>

<header class="topo">
	<div class="casca barra">
		<a class="marca" href="/">
			Garimpo<span class="pingo">.</span>
		</a>
		<nav>
			<a href="/" class:atual={$page.url.pathname === '/'}>Curadoria</a>
			<a href="/quadro" class:atual={$page.url.pathname === '/quadro'}>Quadro</a>
		</nav>
	</div>
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
	nav {
		display: flex;
		gap: var(--r6);
	}
	nav a {
		text-decoration: none;
		font-weight: 600;
		font-size: 0.92rem;
		color: var(--tinta-suave);
		padding-bottom: 2px;
		border-bottom: 2px solid transparent;
	}
	nav a:hover {
		color: var(--tinta);
	}
	nav a.atual {
		color: var(--tinta);
		border-color: var(--ouro);
	}
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
