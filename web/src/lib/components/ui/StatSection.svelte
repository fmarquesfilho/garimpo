<script>
	/**
	 * StatSection — seção do dashboard com ícone, título e conteúdo.
	 * Agrupa métricas com visual clean e suporte a estado de loading/erro/vazio.
	 */
	import { cn } from '$lib/utils';

	let {
		icon = '📊',
		title = '',
		subtitle = '',
		loading = false,
		error = null,
		empty = false,
		emptyMessage = 'Sem dados no período.',
		class: className = '',
		children
	} = $props();
</script>

<section class={cn('rounded-xl border border-border bg-card p-5', className)}>
	<header class="mb-4 flex items-center gap-2">
		<span class="text-xl">{icon}</span>
		<div>
			<h2 class="m-0 text-base font-semibold text-foreground">{title}</h2>
			{#if subtitle}
				<p class="m-0 text-xs text-muted-foreground">{subtitle}</p>
			{/if}
		</div>
	</header>

	{#if loading}
		<div class="flex items-center justify-center py-8">
			<div class="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
			<span class="ml-2 text-sm text-muted-foreground">Carregando…</span>
		</div>
	{:else if error}
		<div class="rounded-md border border-destructive/20 bg-destructive/5 px-4 py-3">
			<p class="m-0 text-sm text-destructive">{error}</p>
		</div>
	{:else if empty}
		<div class="py-6 text-center">
			<p class="m-0 text-sm italic text-muted-foreground">{emptyMessage}</p>
		</div>
	{:else}
		{@render children()}
	{/if}
</section>
