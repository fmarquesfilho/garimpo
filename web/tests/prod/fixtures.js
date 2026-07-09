/**
 * Fixtures para E2E de PRODUÇÃO.
 *
 * - `authedPage`: injeta token Firebase real no app antes do page load.
 *   O SPA detecta __E2E_AUTH_USER__ e pula o fluxo de login normal.
 *   O token é real (obtido via REST no auth.setup.js), então as APIs aceitam.
 *
 * - A fixture também injeta o token na função getIdToken() para que chamadas
 *   fetch incluam o Bearer token real.
 */
import { test as base, expect } from '@playwright/test';
import { readFileSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));

function loadToken() {
	const tokenPath = resolve(__dirname, '../.auth/prod-token.json');
	try {
		return JSON.parse(readFileSync(tokenPath, 'utf-8'));
	} catch {
		throw new Error(
			'Token não encontrado. Rode o auth setup primeiro: npx playwright test --config=playwright.prod.config.js --project=auth-setup'
		);
	}
}

export const test = base.extend({
	authedPage: async ({ page }, use) => {
		const token = loadToken();

		// Injeta o usuário E2E (bypassa Firebase Auth no SPA) E o token real
		// para que getIdToken() retorne o token válido em chamadas de API
		await page.addInitScript(
			({ uid, email, idToken }) => {
				// O SPA checa __E2E_AUTH_USER__ e pula Firebase init
				window.__E2E_AUTH_USER__ = { uid, email, nome: 'E2E Prod', foto: null };
				// Override de getIdToken para retornar o token real
				window.__E2E_ID_TOKEN__ = idToken;
			},
			{ uid: token.uid, email: token.email, idToken: token.idToken }
		);

		await use(page);
	}
});

export { expect };
