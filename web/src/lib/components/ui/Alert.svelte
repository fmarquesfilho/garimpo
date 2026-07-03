<script>
	/**
	 * Alert — feedback banner with semantic variants.
	 * @prop variant — 'info' | 'success' | 'warning' | 'error'
	 * @prop inline — compact mode (no padding, transparent bg)
	 */
	import { cn } from '$lib/utils';

	const VARIANTS = {
		info: 'border-border bg-background text-foreground',
		success: 'border-sucesso-borda bg-sucesso-fundo text-sucesso',
		warning: 'border-aviso-borda bg-aviso-fundo text-aviso',
		error: 'border-erro-borda bg-erro-fundo text-erro'
	};

	let { variant = 'info', inline = false, class: className = '', children, ...rest } = $props();
</script>

<div
	role="alert"
	class={cn(
		'rounded-md border text-sm',
		inline ? 'px-0 py-0 border-none bg-transparent' : 'px-4 py-3',
		VARIANTS[variant] ?? VARIANTS.info,
		className
	)}
	{...rest}
>
	{@render children()}
</div>
