import { defineConfig } from '@playwright/test';

/**
 * Config dos E2E LOCAIS (tests/local/) — sem Firebase/emulador/backend.
 * Usa bypass de auth (window.__E2E_AUTH_USER__) + API mockada.
 *
 * `PW_CHROMIUM` permite apontar para um Chromium pré-instalado (ambientes onde
 * `playwright install` não roda); em CI fica vazio e usa o browser padrão.
 *
 *   npm run test:e2e:local
 */
export default defineConfig({
	testDir: 'tests/local',
	timeout: 60000,
	retries: 0,
	use: {
		baseURL: 'http://localhost:4173',
		headless: true,
		launchOptions: process.env.PW_CHROMIUM ? { executablePath: process.env.PW_CHROMIUM } : {}
	},
	webServer: {
		command: 'bun run preview',
		port: 4173,
		reuseExistingServer: true,
		timeout: 120000
	},
	projects: [{ name: 'local', use: { browserName: 'chromium' } }]
});
