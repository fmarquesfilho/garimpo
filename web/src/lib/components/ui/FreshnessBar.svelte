<script>
	/**
	 * FreshnessBar — indicador de vivacidade do dashboard.
	 * Mostra: dot pulsante + tempo desde update + countdown para próximo poll.
	 */
	import StatusIndicator from './StatusIndicator.svelte';

	let { lastUpdate = 0, countdown = 30, status = 'live' } = $props();

	let tempoRelativo = $derived(() => {
		if (!lastUpdate) return '—';
		const diff = Math.floor((Date.now() - lastUpdate) / 1000);
		if (diff < 60) return `há ${diff}s`;
		if (diff < 3600) return `há ${Math.floor(diff / 60)}min`;
		return `há ${Math.floor(diff / 3600)}h`;
	});

	let statusMap = $derived({
		live: 'ok',
		paused: 'sem_dados',
		offline: 'atrasado',
		idle: 'sem_dados'
	});
</script>

<div class="flex items-center gap-3 text-xs text-muted-foreground">
	<StatusIndicator
		status={statusMap[status] ?? 'sem_dados'}
		label={status === 'paused' ? 'Pausado' : status === 'offline' ? 'Sem conexão' : ''}
	/>
	{#if status === 'live'}
		<span>Atualizado {tempoRelativo()}</span>
		<span class="tabular-nums">· próxima em {countdown}s</span>
	{:else if status === 'paused'}
		<span>Tab inativa</span>
	{:else if status === 'offline'}
		<span>Reconectando…</span>
	{/if}
</div>
