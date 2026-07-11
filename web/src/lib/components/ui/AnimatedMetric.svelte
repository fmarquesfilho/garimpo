<script>
	/**
	 * AnimatedMetric — valor numérico com transição animada (tweened).
	 * Quando `value` muda, interpola suavemente do valor antigo ao novo.
	 */
	import { tweened } from 'svelte/motion';
	import { cubicOut } from 'svelte/easing';
	import { cn } from '$lib/utils';

	let { value = 0, format = (v) => String(Math.round(v)), duration = 600, class: className = '' } = $props();

	const display = tweened(0, { duration: 0, easing: cubicOut });

	// Use $derived for tracking and trigger set imperatively
	let prevValue = $state(undefined);

	$effect(() => {
		const v = value;
		const d = duration;
		if (prevValue === undefined) {
			// First render — no animation
			display.set(v, { duration: 0 });
		} else if (v !== prevValue) {
			display.set(v, { duration: d });
		}
		prevValue = v;
	});
</script>

<span class={cn('tabular-nums', className)}>{format($display)}</span>
