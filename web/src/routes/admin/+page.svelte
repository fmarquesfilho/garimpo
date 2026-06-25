<script>
	import { onMount } from 'svelte';
	import { buscarLogs, alterarNivelLog } from '$lib/api.js';
	import { usuario } from '$lib/firebase.js';

	let logs = $state([]);
	let stats = $state({});
	let total = $state(0);
	let carregando = $state(true);
	let erro = $state(null);
	let filtroNivel = $state('');
	let autoRefresh = $state(true);
	let logLevel = $state('info');
	let intervalo;

	onMount(() => {
		carregar();
		intervalo = setInterval(() => { if (autoRefresh) carregar(); }, 5000);
		return () => clearInterval(intervalo);
	});

	async function carregar() {
		try {
			const r = await buscarLogs({ n: 200, nivel: filtroNivel });
			logs = r?.logs ?? [];
			stats = r?.stats ?? {};
			total = r?.total ?? 0;
			erro = null;
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}
	}

	$effect(() => {
		filtroNivel;
		carregar();
	});

	async function mudarNivel() {
		try {
			await alterarNivelLog(logLevel);
		} catch (e) {
			erro = e.message;
		}
	}

	const nivelCor = { error: 'var(--erro-texto)', warn: 'var(--aviso-texto)', info: 'var(--sucesso-texto)', debug: '#6b7280' };
	const nivelBg = { error: 'var(--erro-fundo)', warn: 'var(--aviso-fundo)', info: 'var(--sucesso-fundo)', debug: '#f9fafb' };
</script>

<svelte:head>
	<title>Admin — Garimpei</title>
</svelte:head>

<section class="admin-page">
	<div class="admin-header">
		<h1>🛠 Admin — Logs</h1>
		<div class="admin-controls">
			<div class="log-level-control">
				<label>Granularidade:</label>
				<select bind:value={logLevel} onchange={mudarNivel}>
					<option value="debug">Debug (tudo)</option>
					<option value="info">Info</option>
					<option value="warn">Warn</option>
					<option value="error">Só erros</option>
				</select>
			</div>
			<label class="auto-refresh">
				<input type="checkbox" bind:checked={autoRefresh} />
				Auto-refresh (5s)
			</label>
		</div>
	</div>

	{#if !$usuario}
		<div class="aviso">Faça login para acessar o admin.</div>
	{:else}
		<!-- Stats -->
		<div class="stats-row">
			<div class="stat-card">
				<span class="stat-num">{total}</span>
				<span class="stat-label">Total</span>
			</div>
			<button class="stat-card" class:ativo={filtroNivel === ''} onclick={() => (filtroNivel = '')}>
				<span class="stat-num">{(stats.info ?? 0) + (stats.warn ?? 0) + (stats.error ?? 0) + (stats.debug ?? 0)}</span>
				<span class="stat-label">No buffer</span>
			</button>
			<button class="stat-card erro-card" class:ativo={filtroNivel === 'error'} onclick={() => (filtroNivel = filtroNivel === 'error' ? '' : 'error')}>
				<span class="stat-num">{stats.error ?? 0}</span>
				<span class="stat-label">Erros</span>
			</button>
			<button class="stat-card warn-card" class:ativo={filtroNivel === 'warn'} onclick={() => (filtroNivel = filtroNivel === 'warn' ? '' : 'warn')}>
				<span class="stat-num">{stats.warn ?? 0}</span>
				<span class="stat-label">Avisos</span>
			</button>
			<button class="stat-card info-card" class:ativo={filtroNivel === 'info'} onclick={() => (filtroNivel = filtroNivel === 'info' ? '' : 'info')}>
				<span class="stat-num">{stats.info ?? 0}</span>
				<span class="stat-label">Info</span>
			</button>
		</div>

		{#if erro}
			<div class="msg-erro">{erro}</div>
		{/if}

		<!-- Lista de logs -->
		{#if carregando}
			<p class="loading">Carregando logs…</p>
		{:else if logs.length === 0}
			<p class="vazio">Nenhum log {filtroNivel ? `de nível "${filtroNivel}"` : ''} no buffer.</p>
		{:else}
			<div class="logs-list">
				{#each logs as log, i (i)}
					<div class="log-entry" style="border-left-color: {nivelCor[log.nivel] ?? '#ccc'}; background: {nivelBg[log.nivel] ?? 'white'}">
						<div class="log-main">
							<span class="log-nivel" style="color: {nivelCor[log.nivel]}">{log.nivel?.toUpperCase()}</span>
							<span class="log-metodo">{log.metodo}</span>
							<span class="log-rota">{log.rota}</span>
							{#if log.status}
								<span class="log-status" class:erro-status={log.status >= 400}>{log.status}</span>
							{/if}
							{#if log.dur_ms}
								<span class="log-dur">{log.dur_ms.toFixed(1)}ms</span>
							{/if}
						</div>
						<span class="log-hora">{new Date(log.em).toLocaleTimeString('pt-BR')}</span>
					</div>
				{/each}
			</div>
		{/if}
	{/if}
</section>

<style>
	.admin-page { max-width: 900px; }
	.admin-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--r6); flex-wrap: wrap; gap: var(--r3); }
	.admin-header h1 { font-size: 1.4rem; margin: 0; }
	.admin-controls { display: flex; align-items: center; gap: var(--r4); flex-wrap: wrap; }
	.log-level-control { display: flex; align-items: center; gap: 6px; font-size: 0.82rem; }
	.log-level-control select { padding: 4px 10px; border: 1px solid var(--linha); border-radius: 8px; font-size: 0.82rem; background: var(--porcelana); }
	.auto-refresh { font-size: 0.82rem; color: var(--tinta-suave); display: flex; align-items: center; gap: 6px; cursor: pointer; }
	.auto-refresh input { accent-color: var(--ouro); }

	.aviso { background: var(--porcelana); padding: var(--r4); border-radius: var(--raio-sm); color: var(--tinta-suave); }
	.msg-erro { background: var(--erro-fundo); color: var(--erro-texto); padding: var(--r3) var(--r4); border-radius: 8px; margin-bottom: var(--r4); }
	.loading, .vazio { color: var(--tinta-suave); font-size: 0.9rem; }

	/* Stats */
	.stats-row { display: flex; gap: var(--r3); margin-bottom: var(--r5); flex-wrap: wrap; }
	.stat-card {
		border: 1px solid var(--linha); background: var(--branco); border-radius: var(--raio-sm);
		padding: var(--r3) var(--r4); text-align: center; cursor: pointer;
		display: flex; flex-direction: column; gap: 2px; min-width: 70px;
	}
	.stat-card.ativo { border-color: var(--ouro); background: var(--ouro-fundo); }
	.stat-card.erro-card.ativo { border-color: var(--erro-borda); background: var(--erro-fundo); }
	.stat-card.warn-card.ativo { border-color: var(--aviso-borda); background: var(--aviso-fundo); }
	.stat-card.info-card.ativo { border-color: var(--sucesso-borda); background: var(--sucesso-fundo); }
	.stat-num { font-size: 1.3rem; font-weight: 700; font-family: var(--mono); }
	.stat-label { font-size: 0.7rem; color: var(--tinta-suave); text-transform: uppercase; }

	/* Logs list */
	.logs-list { display: flex; flex-direction: column; gap: 2px; }
	.log-entry {
		display: flex; justify-content: space-between; align-items: center;
		padding: 6px 12px; border-left: 3px solid; border-radius: 4px;
		font-size: 0.82rem;
	}
	.log-main { display: flex; align-items: center; gap: var(--r3); }
	.log-nivel { font-weight: 700; font-size: 0.7rem; width: 40px; }
	.log-metodo { font-weight: 600; font-size: 0.75rem; color: var(--tinta-suave); }
	.log-rota { font-family: var(--mono); font-size: 0.78rem; }
	.log-status { font-family: var(--mono); font-weight: 600; font-size: 0.78rem; }
	.log-status.erro-status { color: var(--erro-texto); }
	.log-dur { font-family: var(--mono); font-size: 0.72rem; color: var(--tinta-suave); }
	.log-hora { font-size: 0.72rem; color: var(--tinta-suave); font-family: var(--mono); }
</style>
