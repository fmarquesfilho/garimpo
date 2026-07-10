<script>
	/**
	 * Admin Logs — Visualizador de logs estruturados.
	 * Filtra por severidade, serviço, keyword e trace_id.
	 * Em produção: busca do Cloud Logging via /api/admin/logs.
	 * Em dev local: mostra mock (Cloud Logging não disponível).
	 */
	import { onMount } from 'svelte';
	import JsonTree from '$lib/components/JsonTree.svelte';
	import { getIdToken } from '$lib/firebase.js';
	import { Button, Select } from '$lib/components/ui';

	const SEVERITIES = [
		{ value: '', label: 'Todas' },
		{ value: 'DEBUG', label: '🔵 DEBUG' },
		{ value: 'INFO', label: '🟢 INFO' },
		{ value: 'WARNING', label: '🟡 WARNING' },
		{ value: 'ERROR', label: '🔴 ERROR' },
		{ value: 'CRITICAL', label: '⚫ CRITICAL' }
	];

	const SERVICES = [
		{ value: '', label: 'Todos' },
		{ value: 'garimpei-api', label: 'garimpei-api (C#)' },
		{ value: 'garimpei-web', label: 'garimpei-web (Browser)' },
		{ value: 'collector', label: 'collector (Go)' },
		{ value: 'publisher', label: 'publisher (Go)' },
		{ value: 'scheduler', label: 'scheduler (Go)' },
		{ value: 'analyzer', label: 'analyzer (Python)' }
	];

	const WINDOWS = [
		{ value: '15', label: '15 min' },
		{ value: '60', label: '1 hora' },
		{ value: '360', label: '6 horas' },
		{ value: '1440', label: '24 horas' }
	];

	let severity = $state('');
	let service = $state('');
	let keyword = $state('');
	let traceId = $state('');
	let minutes = $state('60');
	let limit = $state('50');

	let entries = $state([]);
	let loading = $state(false);
	let error = $state('');
	let source = $state('');
	let filter = $state('');
	let expandedIdx = $state(-1);

	async function fetchLogs() {
		loading = true;
		error = '';
		try {
			const token = await getIdToken();
			const params = new URLSearchParams();
			if (severity) params.set('severity', severity);
			if (service) params.set('service', service);
			if (keyword.trim()) params.set('keyword', keyword.trim());
			if (traceId.trim()) params.set('traceId', traceId.trim());
			params.set('minutes', minutes);
			params.set('limit', limit);

			const resp = await fetch(`/api/admin/logs?${params}`, {
				headers: { Authorization: `Bearer ${token}` }
			});

			if (!resp.ok) {
				error = `HTTP ${resp.status}`;
				return;
			}

			const data = await resp.json();
			entries = data.entries ?? [];
			source = data.source ?? 'unknown';
			filter = data.filter ?? '';
			if (data.error) error = data.error;
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	onMount(fetchLogs);

	function severityColor(sev) {
		switch (sev?.toUpperCase()) {
			case 'DEBUG':
				return 'text-blue-500';
			case 'INFO':
				return 'text-green-600';
			case 'WARNING':
				return 'text-yellow-600';
			case 'ERROR':
				return 'text-red-600';
			case 'CRITICAL':
				return 'text-red-800 font-bold';
			default:
				return 'text-muted-foreground';
		}
	}

	function severityBadge(sev) {
		switch (sev?.toUpperCase()) {
			case 'DEBUG':
				return '🔵';
			case 'INFO':
				return '🟢';
			case 'WARNING':
				return '🟡';
			case 'ERROR':
				return '🔴';
			case 'CRITICAL':
				return '⚫';
			default:
				return '⚪';
		}
	}

	function formatTime(ts) {
		if (!ts) return '';
		try {
			const d = new Date(ts);
			return d.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
		} catch {
			return ts;
		}
	}
</script>

<svelte:head>
	<title>Logs — Admin — Garimpei</title>
</svelte:head>

<main class="mx-auto max-w-[1100px] space-y-4 p-4">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold">📋 Logs</h1>
		<a href="/admin" class="text-sm text-muted-foreground hover:text-foreground">← Admin</a>
	</div>

	<!-- Filtros -->
	<div class="flex flex-wrap items-end gap-3 rounded-md border border-border bg-card p-3">
		<Select label="Severidade" bind:value={severity} options={SEVERITIES} size="sm" class="w-36" />
		<Select label="Serviço" bind:value={service} options={SERVICES} size="sm" class="w-44" />
		<Select label="Janela" bind:value={minutes} options={WINDOWS} size="sm" class="w-28" />
		<div class="flex flex-col gap-1">
			<span class="text-xs font-semibold text-muted-foreground">Keyword</span>
			<input
				type="text"
				bind:value={keyword}
				placeholder="buscar no log..."
				class="rounded-sm border border-input bg-background px-2 py-1.5 text-sm"
				onkeydown={(e) => e.key === 'Enter' && fetchLogs()}
			/>
		</div>
		<div class="flex flex-col gap-1">
			<span class="text-xs font-semibold text-muted-foreground">Trace ID</span>
			<input
				type="text"
				bind:value={traceId}
				placeholder="trace_id (hex 32 chars)"
				class="w-64 rounded-sm border border-input bg-background px-2 py-1.5 text-sm font-mono"
				onkeydown={(e) => e.key === 'Enter' && fetchLogs()}
			/>
		</div>
		<Button size="sm" onclick={fetchLogs}>
			{loading ? '⏳' : '🔍'} Buscar
		</Button>
	</div>

	<!-- Info -->
	{#if source}
		<div class="flex items-center gap-3 text-xs text-muted-foreground">
			<span>Fonte: <strong>{source}</strong></span>
			<span>{entries.length} entries</span>
			{#if error}
				<span class="text-destructive">⚠️ {error}</span>
			{/if}
		</div>
	{/if}

	<!-- Entries -->
	{#if loading}
		<div class="py-8 text-center text-muted-foreground">⏳ Carregando logs...</div>
	{:else if entries.length === 0}
		<div class="py-8 text-center text-muted-foreground">
			<p class="text-lg">Nenhum log encontrado</p>
			<p class="text-sm">Ajuste os filtros ou amplie a janela de tempo.</p>
		</div>
	{:else}
		<div class="overflow-hidden rounded-md border border-border">
			<table class="w-full text-sm">
				<thead class="bg-muted text-left text-xs text-muted-foreground">
					<tr>
						<th class="px-3 py-2 w-16">Hora</th>
						<th class="px-3 py-2 w-10">Sev</th>
						<th class="px-3 py-2 w-28">Serviço</th>
						<th class="px-3 py-2">Mensagem</th>
					</tr>
				</thead>
				<tbody>
					{#each entries as entry, i (i)}
						<tr
							class="border-t border-border hover:bg-accent/50 cursor-pointer transition-colors {expandedIdx === i
								? 'bg-accent/30'
								: ''}"
							onclick={() => (expandedIdx = expandedIdx === i ? -1 : i)}
						>
							<td class="px-3 py-1.5 font-mono text-xs text-muted-foreground">{formatTime(entry.timestamp)}</td>
							<td class="px-3 py-1.5">{severityBadge(entry.severity)}</td>
							<td class="px-3 py-1.5 font-mono text-xs">{entry.service || '—'}</td>
							<td class="px-3 py-1.5 truncate max-w-[500px] {severityColor(entry.severity)}">
								{entry.message || '(sem mensagem)'}
							</td>
						</tr>
						{#if expandedIdx === i}
							<tr class="border-t border-border bg-muted/50">
								<td colspan="4" class="px-4 py-3">
									<div class="space-y-2 text-xs">
										{#if entry.trace}
											<div>
												<span class="font-semibold">Trace:</span> <code class="text-primary">{entry.trace}</code>
											</div>
										{/if}
										{#if entry.spanId}
											<div><span class="font-semibold">Span:</span> <code>{entry.spanId}</code></div>
										{/if}
										{#if entry.jsonPayload}
											<div class="mt-2 rounded border border-border bg-background p-2 overflow-auto max-h-[300px]">
												<JsonTree data={entry.jsonPayload} />
											</div>
										{/if}
									</div>
								</td>
							</tr>
						{/if}
					{/each}
				</tbody>
			</table>
		</div>
	{/if}

	<!-- Cloud Logging query (debug) -->
	{#if filter}
		<details class="text-xs text-muted-foreground">
			<summary class="cursor-pointer hover:text-foreground">Ver filtro Cloud Logging</summary>
			<pre class="mt-2 rounded bg-muted p-2 overflow-auto">{filter}</pre>
		</details>
	{/if}
</main>
