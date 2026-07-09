/**
 * Fixtures para E2E LOCAL (sem Firebase, sem emulador, sem backend).
 *
 * - `garimparPage`: injeta uma conta de teste (`window.__E2E_AUTH_USER__`) ANTES
 *   do boot, então o app sobe já autenticado (ver src/lib/firebase.js).
 * - `mockApi(page, overrides)`: intercepta `/api/**` com respostas controladas,
 *   permitindo montar cenários determinísticos sem stack.
 *
 * Roda via `npm run test:e2e:local` (playwright.local.config.js) — pensado para
 * validar a página Garimpar localmente antes do push.
 */
import { test as base, expect } from '@playwright/test';

export const TEST_USER = { uid: 'e2e-user', email: 'e2e@teste.dev', nome: 'Teste E2E', foto: null };

export const test = base.extend({
	garimparPage: async ({ page }, use) => {
		await page.addInitScript((u) => {
			window.__E2E_AUTH_USER__ = u;
		}, TEST_USER);
		await use(page);
	}
});

export { expect };

/**
 * Mocka as rotas `/api/**`. `overrides` mapeia um trecho do path → body (objeto
 * ou função(request) → objeto). Rotas não cobertas recebem defaults vazios.
 */
export async function mockApi(page, overrides = {}) {
	const defaults = {
		'/api/buscas': { buscas: [], total: 0 },
		'/api/favoritos': { favoritos: [] },
		'/api/candidatos': { candidatos: [] },
		'/api/lojas/novidades': { variacoes: [], produtos_novos: [], dias_janela: 7 },
		'/api/lojas': { id: 'loja-x', keyword: 'Loja', shop_ids: [1], status: 'adicionada' },
		'/api/categorias': { categorias: [] },
		'/api/alertas': { chat_id: '', threshold: 0.15, apenas_quedas: true, ativo: false }
	};
	const tabela = { ...defaults, ...overrides };
	// Chaves mais específicas primeiro (ex.: /api/lojas/novidades antes de /api/lojas).
	const chaves = Object.keys(tabela).sort((a, b) => b.length - a.length);

	await page.route('**/api/**', async (route) => {
		const path = new URL(route.request().url()).pathname;
		const chave = chaves.find((k) => path.startsWith(k));
		let body = chave ? tabela[chave] : {};
		if (typeof body === 'function') body = body(route.request());
		await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(body) });
	});
}
