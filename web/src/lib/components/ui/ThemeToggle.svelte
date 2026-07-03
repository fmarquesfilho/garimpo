<script>
	/**
	 * ThemeToggle — botão que cicla entre light/dark/system.
	 * Mostra ícone indicando o modo ativo.
	 */
	import { theme } from '$lib/theme.js';
	import { onMount } from 'svelte';

	const MODES = ['light', 'dark', 'system'];
	const ICONS = { light: '☀️', dark: '🌙', system: '🖥️' };
	const LABELS = {
		light: 'Modo claro ativo. Clique para modo escuro.',
		dark: 'Modo escuro ativo. Clique para seguir sistema.',
		system: 'Seguindo sistema. Clique para modo claro.'
	};

	let current = $state('system');

	onMount(() => {
		const unsub = theme.subscribe((value) => {
			current = value;
		});
		return unsub;
	});

	function cycle() {
		const idx = MODES.indexOf(current);
		const next = /** @type {'light'|'dark'|'system'} */ (MODES[(idx + 1) % MODES.length]);
		current = next;
		theme.set(next);
	}
</script>

<button
	class="inline-flex h-9 w-9 items-center justify-center rounded-sm border border-border bg-background text-lg transition-colors hover:border-primary hover:bg-accent"
	onclick={cycle}
	aria-label={LABELS[current]}
	title={LABELS[current]}
	type="button"
>
	{ICONS[current]}
</button>
