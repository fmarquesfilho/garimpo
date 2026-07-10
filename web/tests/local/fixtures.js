/**
 * Fixtures para E2E LOCAL (sem Firebase, sem emulador, sem backend).
 *
 * - `garimparPage`: injeta uma conta de teste (`window.__E2E_AUTH_USER__`) ANTES
 *   do boot, então o app sobe já autenticado (ver src/lib/firebase.js).
 * - `mockApi(page, overrides)`: intercepta `/api/**` com respostas controladas,
 *   permitindo montar cenários determinísticos sem stack.
 * - `mockApiFromFixtures(page)`: carrega respostas dos golden files compartilhados
 *   (`fixtures/respostas/`), garantindo contrato cross-stack.
 *
 * Roda via `npm run test:e2e:local` (playwright.local.config.js) — pensado para
 * validar a página Garimpar localmente antes do push.
 */
import { test as base, expect } from '@playwright/test';
import { readFileSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const fixturesDir = resolve(__dirname, '../../../fixtures');

// ── Dados dos fixtures compartilhados ────────────────────────────────────────

function loadFixture(relativePath) {
	return JSON.parse(readFileSync(resolve(fixturesDir, relativePath), 'utf-8'));
}

export const FIXTURE_LOJAS = loadFixture('lojas.json');
export const FIXTURE_PRODUTOS = loadFixture('produtos.json');
export const FIXTURE_BUSCAS = loadFixture('buscas.json');
export const FIXTURE_API_BUSCAS = loadFixture('respostas/api-buscas.json');
export const FIXTURE_API_CANDIDATOS = loadFixture('respostas/api-candidatos.json');
export const FIXTURE_API_NOVIDADES = loadFixture('respostas/api-novidades.json');
export const FIXTURE_COLLECTOR_FETCHSHOP = loadFixture('respostas/collector-fetchshop.json');

// ── User de teste ────────────────────────────────────────────────────────────

export const TEST_USER = { uid: 'e2e-user', email: 'e2e@teste.dev', nome: 'Teste E2E', foto: null };

// ── Fixture Playwright: página já autenticada ────────────────────────────────

export const test = base.extend({
	garimparPage: async ({ page }, use) => {
		await page.addInitScript((u) => {
			window.__E2E_AUTH_USER__ = u;
		}, TEST_USER);
		await use(page);
	}
});

export { expect };

// ── Mock API com defaults dos fixtures ───────────────────────────────────────

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
		'/api/categorias': { marketplaces: [{ marketplace: 'shopee', categorias: [] }] },
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

/**
 * Mocka a API usando dados reais dos golden files (fixtures/respostas/).
 * Simula o que produção retornaria para as 3 lojas de teste.
 * Ideal para testes que validam o contrato cross-stack end-to-end.
 */
export async function mockApiFromFixtures(page, overrides = {}) {
	const fixtureDefaults = {
		'/api/buscas': FIXTURE_API_BUSCAS,
		'/api/favoritos': { favoritos: [] },
		'/api/candidatos': FIXTURE_API_CANDIDATOS,
		'/api/lojas/novidades': FIXTURE_API_NOVIDADES,
		'/api/lojas': { id: 'loja-glory', keyword: 'Glory of Seoul', shop_ids: [920292999], status: 'adicionada' },
		'/api/categorias': {
			marketplaces: [
				{
					marketplace: 'shopee',
					categorias: [
						{ id: 100630, nome: 'Beleza' },
						{ id: 100664, nome: 'Cuidados com a Pele' },
						{ id: 100640, nome: 'Perfumaria' }
					]
				}
			]
		},
		'/api/alertas': { chat_id: '', threshold: 0.15, apenas_quedas: true, ativo: false }
	};
	await mockApi(page, { ...fixtureDefaults, ...overrides });
}
