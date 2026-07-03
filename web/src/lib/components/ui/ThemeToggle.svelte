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
	class="theme-toggle"
	onclick={cycle}
	aria-label={LABELS[current]}
	title={LABELS[current]}
>
	{ICONS[current]}
</button>

<style>
	.theme-toggle {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border: 1px solid var(--linha);
		border-radius: var(--raio-sm);
		background: var(--porcelana);
		cursor: pointer;
		font-size: var(--text-lg);
		transition: border-color 0.15s ease, background 0.15s ease;
	}
	.theme-toggle:hover {
		border-color: var(--ouro);
		background: var(--ouro-fundo);
	}
	.theme-toggle:focus-visible {
		outline: 2px solid var(--ouro);
		outline-offset: 2px;
	}

	@media (prefers-reduced-motion: reduce) {
		.theme-toggle { transition-duration: 0ms; }
	}
</style>
