<script>
	/**
	 * OpportunityCard — card de oportunidade (queda de preço ou produto novo).
	 * Visual chamativo para ação imediata.
	 */
	import { cn } from '$lib/utils';
	import Badge from './Badge.svelte';

	let { produto = {}, tipo = 'queda', class: className = '' } = $props();

	let variacao = $derived(produto.variacao ? Math.abs(produto.variacao * 100).toFixed(0) : null);
	let precoFormatado = $derived(
		produto.preco_atual ? `R$ ${produto.preco_atual.toFixed(2)}` : produto.preco ? `R$ ${produto.preco.toFixed(2)}` : ''
	);
</script>

<a
	href={produto.link || '#'}
	target="_blank"
	rel="noopener"
	class={cn(
		'group flex items-center gap-3 rounded-lg border border-border bg-card p-3 transition-all hover:border-primary/50 hover:shadow-md',
		className
	)}
>
	{#if produto.imagem}
		<img src={produto.imagem} alt="" class="h-12 w-12 rounded-md object-cover" loading="lazy" />
	{:else}
		<div class="flex h-12 w-12 items-center justify-center rounded-md bg-muted text-lg">
			{tipo === 'queda' ? '📉' : '🆕'}
		</div>
	{/if}

	<div class="min-w-0 flex-1">
		<p class="m-0 truncate text-sm font-medium text-foreground group-hover:text-primary">
			{produto.nome || '—'}
		</p>
		<div class="mt-0.5 flex items-center gap-2">
			{#if tipo === 'queda' && variacao}
				<Badge variant="success">−{variacao}%</Badge>
			{:else if tipo === 'novo'}
				<Badge variant="warning">Novo</Badge>
			{:else if tipo === 'alto_valor'}
				<Badge variant="default">Alto valor</Badge>
			{/if}
			<span class="text-xs text-muted-foreground">{produto.loja || ''}</span>
		</div>
	</div>

	<div class="text-right">
		<span class="text-sm font-bold text-foreground">{precoFormatado}</span>
		{#if tipo === 'queda' && produto.preco_anterior}
			<p class="m-0 text-xs text-muted-foreground line-through">
				R$ {produto.preco_anterior.toFixed(2)}
			</p>
		{/if}
	</div>
</a>
