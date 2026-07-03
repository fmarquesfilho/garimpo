<script>
	/**
	 * Button — shadcn-svelte style with Tailwind + brand tokens.
	 * @prop variant — 'primary' | 'secondary' | 'danger' | 'ghost' | 'link'
	 * @prop size — 'sm' | 'md' | 'lg' | 'icon'
	 */
	import { cn } from '$lib/utils';

	const VARIANTS = {
		primary: 'bg-primary text-primary-foreground hover:bg-ouro-hover shadow-sm',
		secondary: 'border border-border bg-background hover:bg-accent hover:text-accent-foreground',
		danger: 'bg-destructive text-destructive-foreground hover:bg-rosa-hover shadow-sm',
		ghost: 'hover:bg-accent hover:text-accent-foreground',
		link: 'text-primary underline-offset-4 hover:underline'
	};

	const SIZES = {
		sm: 'h-8 px-3 text-xs rounded-sm gap-1',
		md: 'h-9 px-4 text-sm rounded-md gap-2',
		lg: 'h-11 px-6 text-base rounded-md gap-2',
		icon: 'h-9 w-9 rounded-md'
	};

	let {
		variant = 'primary',
		size = 'md',
		class: className = '',
		type = /** @type {'button'|'submit'|'reset'} */ ('button'),
		disabled = false,
		children,
		...rest
	} = $props();
</script>

<button
	{type}
	{disabled}
	class={cn(
		'inline-flex items-center justify-center whitespace-nowrap font-semibold transition-colors focus-visible:outline-2 focus-visible:outline-ring focus-visible:outline-offset-2 disabled:pointer-events-none disabled:opacity-50',
		VARIANTS[variant] ?? VARIANTS.primary,
		SIZES[size] ?? SIZES.md,
		className
	)}
	{...rest}
>
	{@render children()}
</button>
