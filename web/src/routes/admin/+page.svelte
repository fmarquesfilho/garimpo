<script>
	import { onMount } from 'svelte';

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

<main class="mx-auto max-w-[600px] p-5">
	<h1 class="mb-4">Painel Admin</h1>

	{#if carregando}
		<p>Carregando...</p>
	{:else if health}
		<section class="mb-4 rounded-md bg-card p-4">
			<h2>Status do sistema</h2>
			<table class="w-full">
				<tbody>
					<tr
						><td class="py-1.5 font-medium text-tinta-suave">Backend</td><td class="py-1.5"
							><strong>{health.backend ?? 'unknown'}</strong></td
						></tr
					>
					<tr><td class="py-1.5 font-medium text-tinta-suave">Status</td><td class="py-1.5">{health.status}</td></tr>
					<tr
						><td class="py-1.5 font-medium text-tinta-suave">Store</td><td class="py-1.5">{health.store ?? '-'}</td></tr
					>
					<tr
						><td class="py-1.5 font-medium text-tinta-suave">Fonte</td><td class="py-1.5">{health.fonte ?? '-'}</td></tr
					>
				</tbody>
			</table>
		</section>

		{#if health.quality}
			<section class="mb-4 rounded-md bg-card p-4">
				<h2>Qualidade de código</h2>
				<table class="w-full">
					<tbody>
						<tr
							><td class="py-1.5 font-medium text-tinta-suave">Go lint</td><td class="py-1.5"
								>{health.quality.lint_go}</td
							></tr
						>
						<tr
							><td class="py-1.5 font-medium text-tinta-suave">Python lint</td><td class="py-1.5"
								>{health.quality.lint_python}</td
							></tr
						>
						<tr
							><td class="py-1.5 font-medium text-tinta-suave">C# lint</td><td class="py-1.5"
								>{health.quality.lint_csharp}</td
							></tr
						>
						<tr
							><td class="py-1.5 font-medium text-tinta-suave">Testes C#</td><td class="py-1.5"
								>{health.quality.tests_csharp} testes</td
							></tr
						>
						<tr
							><td class="py-1.5 font-medium text-tinta-suave">Pre-push checks</td><td class="py-1.5"
								>{health.quality.pre_push_checks} checks</td
							></tr
						>
					</tbody>
				</table>
			</section>
		{/if}

		<section>
			<h2>Links úteis</h2>
			<ul class="list-none p-0">
				<li class="my-2">
					<a class="text-ouro no-underline hover:underline" href="/admin/logs">📋 Logs (Cloud Logging)</a>
				</li>
				<li class="my-2">
					<a class="text-ouro no-underline hover:underline" href="/docs/" target="_blank">📚 Documentação (Starlight)</a
					>
				</li>
				<li class="my-2">
					<a class="text-ouro no-underline hover:underline" href="/docs/api-reference.html" target="_blank"
						>📋 API Reference (Scalar)</a
					>
				</li>
				<li class="my-2">
					<a class="text-ouro no-underline hover:underline" href="https://app.codacy.com" target="_blank"
						>🔍 Codacy (Code Quality)</a
					>
				</li>
				<li class="my-2">
					<a
						class="text-ouro no-underline hover:underline"
						href="https://github.com/fmarquesfilho/garimpo/actions"
						target="_blank">⚙️ GitHub Actions (CI)</a
					>
				</li>
				<li class="my-2">
					<a
						class="text-ouro no-underline hover:underline"
						href="https://github.com/fmarquesfilho/garimpo/pulls"
						target="_blank">🔀 Pull Requests</a
					>
				</li>
				<li class="my-2">
					<a
						class="text-ouro no-underline hover:underline"
						href="https://console.cloud.google.com/run?project=garimpo-500114"
						target="_blank">☁️ Cloud Run Console</a
					>
				</li>
				<li class="my-2">
					<a class="text-ouro no-underline hover:underline" href="https://console.neon.tech" target="_blank"
						>🐘 Neon (PostgreSQL)</a
					>
				</li>
				<li class="my-2">
					<a class="text-ouro no-underline hover:underline" href="https://dash.cloudflare.com" target="_blank"
						>🌐 Cloudflare Dashboard</a
					>
				</li>
			</ul>
		</section>
	{/if}
</main>
