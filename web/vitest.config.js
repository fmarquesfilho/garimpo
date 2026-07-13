import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vitest/config';
import path from 'path';

export default defineConfig({
	plugins: [svelte({ hot: false })],
	test: {
		include: ['src/**/*.test.{js,ts}'],
		environment: 'jsdom',
		setupFiles: ['src/tests/setup.js'],
		coverage: {
			provider: 'v8',
			reporter: ['text', 'html', 'lcov'],
			reportsDirectory: '../reports/coverage',
			include: ['src/lib/**'],
			exclude: ['src/lib/components/ui/**', 'src/lib/firebase.js', 'src/lib/telemetry.js'],
			thresholds: {
				'src/lib/busca-engine.svelte.js': { lines: 70, branches: 55 },
				'src/lib/busca-config.js': { lines: 80, branches: 75 },
				'src/lib/omnibox-intencao.js': { lines: 85, branches: 70 },
				'src/lib/omnibox-parser.js': { lines: 85, branches: 80 },
				'src/lib/omnibox-sugestoes.js': { lines: 80, branches: 75 },
				'src/lib/loja-registry.js': { lines: 80, branches: 75 }
			}
		}
	},
	resolve: {
		conditions: ['browser'],
		alias: {
			$lib: path.resolve('./src/lib')
		}
	}
});
