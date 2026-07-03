<script>
	/**
	 * Badge — small label with semantic variants.
	 * @prop variant — 'default' | 'secondary' | 'success' | 'warning' | 'error' | 'outline'
	 */
	import { cn } from '$lib/utils';

	const VARIANTS = {
		default: 'bg-primary text-primary-foreground border-transparent',
		secondary: 'bg-secondary text-secondary-foreground border-transparent',
		success: 'bg-sucesso-fundo text-sucesso border-sucesso-borda',
		warning: 'bg-aviso-fundo text-aviso border-aviso-borda',
		error: 'bg-erro-fundo text-erro border-erro-borda',
		outline: 'border-border text-foreground'
	};

	let { variant = 'default', class: className = '', children, ...rest } = $props();
</script>

<span
	class={cn(
		'inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-bold whitespace-nowrap transition-colors',
		VARIANTS[variant] ?? VARIANTS.default,
		className
	)}
	{...rest}
>
	{@render children()}
</span>
