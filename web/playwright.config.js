import { defineConfig } from '@playwright/test';

export default defineConfig({
	testDir: 'tests',
	timeout: 60000,
	retries: 1,
	use: {
		baseURL: 'http://localhost:4173',
		headless: true
	},
	webServer: {
		command: 'bun run preview',
		port: 4173,
		reuseExistingServer: true,
		timeout: 120000
	},
	projects: [
		// Testes autenticados (usam fixture authedPage com emulator)
		{
			name: 'autenticado',
			use: { browserName: 'chromium' },
			testIgnore: [/auth\.setup\.js/, /smoke\.spec\.js/]
		},
		// Testes de smoke: sem autenticação
		{
			name: 'smoke',
			use: { browserName: 'chromium' },
			testMatch: /smoke\.spec\.js/
		}
	]
});
