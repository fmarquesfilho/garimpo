<script>
	/**
	 * Lane — casca de uma "raia" da página Descobrir. Cabeçalho com título, tag,
	 * contador e ações (limpar raia), corpo colapsável. O estado `open` é bindable
	 * para permitir o "colapsar tudo" a partir dos controles superiores.
	 *
	 * @prop title — título da raia
	 * @prop tag — legenda curta (o que a raia contém)
	 * @prop count — texto do contador (ex.: "2 aplicados")
	 * @prop open — bind:open
	 * @prop actions — snippet opcional de ações no canto direito do cabeçalho
	 * @prop children — corpo da raia
	 */
	import { cn } from '$lib/utils';

	let { title = '', tag = '', count = '', open = $bindable(true), actions = null, children } = $props();
</script>

<section class="overflow-hidden rounded-md border border-border bg-card">
	<div
		class={cn(
			'flex items-center gap-2.5 border-b border-border bg-muted px-3.5 py-2.5',
			!open && 'border-b-transparent'
		)}
	>
		<button
			type="button"
			class="flex flex-1 items-center gap-2.5 text-left"
			aria-expanded={open}
			onclick={() => (open = !open)}
		>
			<span class="inline-block text-xs text-muted-foreground transition-transform duration-150" class:rotate-90={!open}
				>▾</span
			>
			<span class="flex items-center gap-2 font-[var(--display)] text-[1.02rem] font-bold text-foreground">
				{title}
				{#if tag}<span
						class="rounded border border-border bg-card px-1.5 py-0.5 font-[var(--mono)] text-[0.62rem] uppercase tracking-wider text-muted-foreground"
						>{tag}</span
					>{/if}
			</span>
			<span class="font-[var(--mono)] text-xs text-muted-foreground">{count}</span>
		</button>
		{#if actions}
			<div class="ml-auto flex items-center gap-1">{@render actions()}</div>
		{/if}
	</div>
	{#if open}
		<div class="p-3.5">{@render children()}</div>
	{/if}
</section>
