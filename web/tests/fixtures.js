/**
 * Fixtures compartilhados para testes E2E autenticados.
 *
 * Extende o `test` do Playwright com uma `page` que já está
 * conectada ao Firebase Auth Emulator e logada.
 */
import { test as base, expect } from '@playwright/test';

const API_KEY = 'AIzaSyA5sBUoVkNHiq58KUkmwbxIMLhvgTn7N8A';
const TEST_EMAIL = 'teste-e2e@garimpo.dev';
const TEST_PASSWORD = 'senha-teste-123';

export const test = base.extend({
	/** Page já autenticada via emulator. */
	authedPage: async ({ page }, use) => {
		const emulatorHost = process.env.FIREBASE_AUTH_EMULATOR_HOST;

		if (emulatorHost) {
			// Injetar variável para firebase.js conectar ao emulator
			await page.addInitScript(({ host }) => {
				window.__FIREBASE_AUTH_EMULATOR_HOST__ = host;
			}, { host: emulatorHost });
		}

		await page.goto('/');
		await page.waitForLoadState('networkidle');

		if (emulatorHost) {
			// Esperar __TEST_SIGN_IN__ ficar disponível e logar
			await page.waitForFunction(
				() => typeof window.__TEST_SIGN_IN__ === 'function',
				{},
				{ timeout: 10000 }
			);
			await page.evaluate(async ({ email, password }) => {
				await window.__TEST_SIGN_IN__(email, password);
			}, { email: TEST_EMAIL, password: TEST_PASSWORD });
		}

		// Esperar o conteúdo autenticado
		await expect(page.locator('h1')).toContainText('O que publicar hoje', { timeout: 15000 });

		await use(page);
	}
});

export { expect };
