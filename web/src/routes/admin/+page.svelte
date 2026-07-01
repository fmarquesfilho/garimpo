<script>
	import { onMount } from 'svelte';
	import { usuario, getIdToken } from '$lib/firebase.js';

	let health = $state(null);
	let carregando = $state(true);

	onMount(async () => {
		try {
			const resp = await fetch('/api/health');
			health = await resp.json();
		} catch (e) {
			health = { status: 'erro', detail: e.message };
		} finally {
			carregando = false;
		}
	});
</script>

<svelte:head>
	<title>Admin — Garimpei</title>
</svelte:head>

<main class="admin">
	<h1>Painel Admin</h1>

	{#if carregando}
		<p>Carregando...</p>
	{:else if health}
		<section class="status-card">
			<h2>Status do sistema</h2>
			<table>
				<tbody>
				<tr><td>Backend</td><td><strong>{health.backend ?? 'unknown'}</strong></td></tr>
				<tr><td>Status</td><td>{health.status}</td></tr>
				<tr><td>Store</td><td>{health.store ?? '-'}</td></tr>
				<tr><td>Fonte</td><td>{health.fonte ?? '-'}</td></tr>
				</tbody>
			</table>
		</section>

		<section class="links">
			<h2>Links úteis</h2>
			<ul>
				<li><a href="/docs/" target="_blank">Documentação (Starlight)</a></li>
				<li><a href="/docs/api-reference.html" target="_blank">API Reference (Scalar)</a></li>
				<li><a href="https://console.cloud.google.com/run?project=garimpo-500114" target="_blank">Cloud Run Console</a></li>
				<li><a href="https://console.neon.tech" target="_blank">Neon (PostgreSQL)</a></li>
				<li><a href="https://dash.cloudflare.com" target="_blank">Cloudflare Dashboard</a></li>
			</ul>
		</section>
	{/if}
</main>

<style>
	.admin { max-width: 600px; margin: 0 auto; padding: var(--r5); }
	h1 { margin-bottom: var(--r4); }
	.status-card { background: var(--nevoa); padding: var(--r4); border-radius: var(--raio); margin-bottom: var(--r4); }
	table { width: 100%; }
	td { padding: 6px 0; }
	td:first-child { font-weight: 500; color: var(--texto-secundario); }
	.links ul { list-style: none; padding: 0; }
	.links li { margin: 8px 0; }
	.links a { color: var(--ouro); text-decoration: none; }
	.links a:hover { text-decoration: underline; }
</style>
