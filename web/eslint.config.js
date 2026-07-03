import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';

export default [
	js.configs.recommended,
	...svelte.configs['flat/recommended'],
	{
		rules: {
			'no-unused-vars': ['warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^\\$' }],
			'no-undef': 'off', // Svelte 5 runes ($state, $derived, $effect, $props, $bindable, $derived)
			'svelte/require-each-key': 'off', // nem todo each precisa de key
			'svelte/no-navigation-without-resolve': 'off', // SvelteKit resolve automaticamente
			'svelte/no-at-html-tags': 'warn', // {@html} é XSS risk — warn para review consciente
			'svelte/valid-compile': 'off', // conflita com Svelte 5 runes
			'svelte/prefer-svelte-reactivity': 'off', // Map/Set em $derived é OK
			'no-useless-assignment': 'off', // falsos positivos em Svelte (variáveis reativas)
			'no-useless-escape': 'warn', // downgrade para warning

			// ── Complexidade — mantém código manutenível ──────────────────
			'complexity': ['warn', { max: 15 }], // cyclomatic complexity
			'max-depth': ['warn', { max: 4 }], // nesting depth
			'max-lines-per-function': ['warn', { max: 80, skipBlankLines: true, skipComments: true }],
			'max-params': ['warn', { max: 4 }], // function parameters
		}
	},
	{
		files: ['**/*.svelte'],
		rules: {
			// Svelte-specific complexity
			'svelte/max-lines-per-block': ['warn', { script: 180, style: 80, template: 300 }],
		}
	},
	{
		ignores: ['build/', '.svelte-kit/', 'node_modules/']
	}
];
