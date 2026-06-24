import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	test: {
		include: ['src/**/*.test.{js,ts}'],
		environment: 'jsdom',
		setupFiles: ['src/tests/setup.js'],
		alias: {
			// Garante que Svelte resolve no modo client (não server/SSR)
			'svelte': 'svelte'
		},
		deps: {
			// Força o vitest a processar o Svelte no modo client
			optimizer: {
				web: {
					include: ['@testing-library/svelte']
				}
			}
		}
	},
	resolve: {
		conditions: ['browser']
	}
});
