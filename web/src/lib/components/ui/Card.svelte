<script>
	/**
	 * Card — surface container with border and radius.
	 * @prop class — extra Tailwind classes
	 */
	import { cn } from '$lib/utils';

	let { class: className = '', children, ...rest } = $props();
</script>

<div
	class={cn(
		'rounded-2xl border border-border bg-card/95 backdrop-blur-md text-card-foreground shadow-md transition-all duration-250',
		className
	)}
	{...rest}
>
	{@render children()}
</div>
