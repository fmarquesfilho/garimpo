import { defineConfig } from '@playwright/test';
import { config } from 'dotenv';
import { resolve } from 'path';

// Carrega .env.e2e.local (credenciais reais, gitignored)
// Fallback: .env.e2e (template sem senha — falha no login)
config({ path: resolve(import.meta.dirname, '.env.e2e.local') });
config({ path: resolve(import.meta.dirname, '.env.e2e') });

/**
 * Config dos E2E de PRODUÇÃO — roda contra APIs e serviços reais.
 * NÃO mocka nada. Usa Firebase Auth real (email/senha).
 *
 * Pré-requisitos:
 *   1. Copiar .env.e2e → .env.e2e.local e preencher E2E_PASSWORD
 *   2. Usuário de teste criado no Firebase Console
 *
 * Rodar:
 *   npm run test:e2e:prod
 *   npm run test:e2e:prod -- --headed  (com browser visível)
 */
export default defineConfig({
	testDir: 'tests/prod',
	timeout: 60000,
	retries: 1,
	use: {
		baseURL: process.env.E2E_BASE_URL || 'https://garimpei.app.br',
		headless: true,
		trace: 'on-first-retry',
		launchOptions: process.env.PW_CHROMIUM ? { executablePath: process.env.PW_CHROMIUM } : {}
	},
	projects: [
		{
			name: 'auth-setup',
			testMatch: /auth\.setup\.js/
		},
		{
			name: 'prod',
			use: {
				browserName: 'chromium',
				storageState: 'tests/.auth/prod-user.json'
			},
			dependencies: ['auth-setup']
		}
	]
});
