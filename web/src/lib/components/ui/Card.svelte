<script>
	/**
	 * Card — surface container with border and radius.
	 * @prop class — extra Tailwind classes
	 */
	import { cn } from '$lib/utils';

	let { class: className = '', children, ...rest } = $props();
</script>

<div class={cn('rounded-md border border-border bg-card text-card-foreground shadow-sm', className)} {...rest}>
	{@render children()}
</div>
