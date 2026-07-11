<script>
	/**
	 * StatusIndicator — badge pulsante com cor semântica para status de sistema.
	 * @prop status — 'ok' | 'atrasado' | 'sem_dados' | 'erro' | 'indisponivel'
	 * @prop label — texto do badge (opcional, usa default por status)
	 */
	import { cn } from '$lib/utils';

	let { status = 'ok', label = '', class: className = '' } = $props();

	const statusConfig = {
		ok: { cor: 'bg-emerald-500', texto: 'text-emerald-700 dark:text-emerald-400', label: 'Operando', pulso: true },
		atrasado: { cor: 'bg-amber-500', texto: 'text-amber-700 dark:text-amber-400', label: 'Atrasado', pulso: true },
		sem_dados: { cor: 'bg-muted-foreground/50', texto: 'text-muted-foreground', label: 'Sem dados', pulso: false },
		erro: { cor: 'bg-destructive', texto: 'text-destructive', label: 'Erro', pulso: true },
		indisponivel: { cor: 'bg-muted-foreground/50', texto: 'text-muted-foreground', label: 'Indisponível', pulso: false }
	};

	let cfg = $derived(statusConfig[status] ?? statusConfig.sem_dados);
	let displayLabel = $derived(label || cfg.label);
</script>

<span
	class={cn(
		'inline-flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1 text-sm font-medium',
		cfg.texto,
		className
	)}
>
	<span class="relative flex h-2.5 w-2.5">
		{#if cfg.pulso}
			<span class={cn('absolute inline-flex h-full w-full animate-ping rounded-full opacity-75', cfg.cor)}></span>
		{/if}
		<span class={cn('relative inline-flex h-2.5 w-2.5 rounded-full', cfg.cor)}></span>
	</span>
	{displayLabel}
</span>
