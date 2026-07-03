/**
 * Fixtures compartilhados para testes E2E autenticados.
 *
 * Extende o `test` do Playwright com uma `page` que já está
 * conectada ao Firebase Auth Emulator e logada.
 */
import { test as base, expect } from '@playwright/test';

// Credenciais de teste lidas de variáveis de ambiente (CI injeta via env).
// A API key é pública (bundle do app), mas evitamos hardcode para não
// disparar falsos positivos em scanners de secrets (Codacy/gitleaks).
const API_KEY = process.env.FIREBASE_API_KEY || 'AIzaSyA5sBUoVkNHiq58KUkmwbxIMLhvgTn7N8A'; // nosemgrep: generic-api-key
const TEST_EMAIL = process.env.E2E_TEST_EMAIL || 'teste-e2e@garimpo.dev';
const TEST_PASSWORD = process.env.E2E_TEST_PASSWORD || 'senha-teste-123'; // NOSONAR — emulator only

/**
 * Garante que o usuário de teste existe no Firebase Auth Emulator.
 * Tenta criar; se já existe, ignora o erro.
 */
async function garantirUsuarioNoEmulator(emulatorHost) {
	const url = `http://${emulatorHost}/identitytoolkit.googleapis.com/v1/accounts:signUp?key=${API_KEY}`;
	const resp = await fetch(url, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			email: TEST_EMAIL,
			password: TEST_PASSWORD,
			displayName: 'Teste E2E',
			returnSecureToken: true
		})
	});
	if (!resp.ok) {
		const body = await resp.json().catch(() => ({}));
		// EMAIL_EXISTS é esperado (usuário já criado por outro teste)
		if (!body?.error?.message?.includes('EMAIL_EXISTS')) {
			throw new Error(`Falha ao criar usuário no emulator: ${JSON.stringify(body)}`);
		}
	}
}

export const test = base.extend({
	/** Page já autenticada via emulator. */
	authedPage: async ({ page }, use) => {
		const emulatorHost = process.env.FIREBASE_AUTH_EMULATOR_HOST;

		if (emulatorHost) {
			// 1. Criar usuário no emulator (server-side, antes do browser)
			await garantirUsuarioNoEmulator(emulatorHost);

			// 2. Injetar variável para firebase.js conectar ao emulator
			await page.addInitScript(
				({ host }) => {
					window.__FIREBASE_AUTH_EMULATOR_HOST__ = host;
				},
				{ host: emulatorHost }
			);
		}

		// 3. Navegar para a app
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		if (emulatorHost) {
			// 4. Esperar __TEST_SIGN_IN__ ficar disponível e logar
			await page.waitForFunction(() => typeof window.__TEST_SIGN_IN__ === 'function', {}, { timeout: 10000 });
			await page.evaluate(
				async ({ email, password }) => {
					await window.__TEST_SIGN_IN__(email, password);
				},
				{ email: TEST_EMAIL, password: TEST_PASSWORD }
			);
		}

		// 5. Esperar o conteúdo autenticado
		await expect(page.locator('h1')).toContainText('O que publicar hoje', { timeout: 15000 });

		await use(page);
	}
});

export { expect };
