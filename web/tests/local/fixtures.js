/**
 * E2E Local — Fixtures e mock de API.
 *
 * Padrão: Playwright Fixture Composition + Route Interception.
 *
 * A API mock é registrada ANTES da navegação (via fixture) para garantir
 * que nenhuma request escapa do interceptor — elimina race conditions
 * entre o boot da SPA e o registro de rotas.
 *
 * Uso:
 *   import { test, expect } from './fixtures.js';
 *
 *   test('cenário', async ({ authedPage }) => {
 *     // authedPage já tem auth bypass + API mock (defaults vazios)
 *     await authedPage.goto('/');
 *   });
 *
 *   test('cenário custom', async ({ authedPage }) => {
 *     // Override de respostas antes de navegar
 *     authedPage.apiOverrides({ '/api/buscas': { buscas: [...], total: 1 } });
 *     await authedPage.goto('/');
 *   });
 */
import { test as base, expect } from '@playwright/test';
import { readFileSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const fixturesDir = resolve(__dirname, '../../../fixtures');

// ── Fixture data (golden files) ──────────────────────────────────────────────

function loadFixture(path) {
	return JSON.parse(readFileSync(resolve(fixturesDir, path), 'utf-8'));
}

export const FIXTURE_LOJAS = loadFixture('lojas.json');
export const FIXTURE_PRODUTOS = loadFixture('produtos.json');
export const FIXTURE_BUSCAS = loadFixture('buscas.json');
export const FIXTURE_API_BUSCAS = loadFixture('respostas/api-buscas.json');
export const FIXTURE_API_CANDIDATOS = loadFixture('respostas/api-candidatos.json');
export const FIXTURE_API_NOVIDADES = loadFixture('respostas/api-novidades.json');
export const FIXTURE_COLLECTOR_FETCHSHOP = loadFixture('respostas/collector-fetchshop.json');
export const DATASET_DESCOBRIR = loadFixture('respostas/dataset-descobrir.json');

export const TEST_USER = { uid: 'e2e-user', email: 'e2e@teste.dev', nome: 'Teste E2E', foto: null };

// ── Default responses (all endpoints the SPA calls) ──────────────────────────

const API_DEFAULTS = {
	'/api/admin/me': { admin: false, email: 'e2e@teste.dev', tools: {} },
	'/api/buscas': { buscas: [], total: 0 },
	'/api/favoritos': { favoritos: [] },
	'/api/candidatos': { candidatos: [], total_bruto: 0, estrategia: 'nicho' },
	'/api/lojas/novidades': { variacoes: [], produtos_novos: [], dias_janela: 7 },
	'/api/lojas/evolucao': { evolucao: [] },
	'/api/lojas/registro': { lojas: [], total: 0 },
	'/api/lojas/buscar': { lojas: [], total: 0 },
	'/api/lojas/resolver': {
		id: '999',
		nome: 'Loja Resolvida',
		nome_normalizado: 'lojaresolvida',
		marketplace: 'shopee',
		cron: null,
		origem: null,
		monitorada: false,
		imagem: null,
		seguidores: null,
		total_produtos: null,
		avaliacao: null
	},
	'/api/lojas': { id: 'loja-x', keyword: 'Loja', shop_ids: [1], status: 'adicionada' },
	'/api/categorias': { marketplaces: [{ marketplace: 'shopee', categorias: [] }] },
	'/api/alertas': { chat_id: '', threshold: 0.15, apenas_quedas: true, ativo: false },
	'/api/onboarding/status': { etapa: 'concluido' },
	'/api/destinos': { destinos: [] },
	'/api/templates': { templates: [] },
	'/api/publicacoes': { publicacoes: [] },
	'/api/estatisticas': {},
	'/api/coletas': { coletas: [] },
	'/api/conversoes': { conversoes: [] }
};

// ── Route handler ────────────────────────────────────────────────────────────

/**
 * Cria um handler de rota que resolve requests por longest-prefix-match.
 * Suporta overrides como objetos estáticos ou funções (request) → objeto.
 */
function createRouteHandler(responseTable) {
	const keys = Object.keys(responseTable).sort((a, b) => b.length - a.length);

	return async (route) => {
		const url = new URL(route.request().url());
		const path = url.pathname;
		// Re-sort keys each time in case responseTable was mutated
		const currentKeys = Object.keys(responseTable).sort((a, b) => b.length - a.length);
		const key = currentKeys.find((k) => path.startsWith(k));
		let body = key ? responseTable[key] : {};

		if (typeof body === 'function') {
			body = await body(route.request());
		}

		await route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify(body)
		});
	};
}

// ── Playwright fixtures ──────────────────────────────────────────────────────

export const test = base.extend({
	/**
	 * Página com auth bypass + API mock pré-registrada.
	 *
	 * A mock intercepta TODAS as rotas /api/ com defaults seguros.
	 * Use `page.apiOverrides(obj)` antes do goto para customizar respostas.
	 * Use `page.apiFromFixtures(overrides?)` para usar golden files.
	 */
	authedPage: async ({ page }, use) => {
		// Mutable response table — tests can override before navigating
		const responses = { ...API_DEFAULTS };

		// Register route FIRST (before any navigation or script injection)
		await page.route('**/api/**', createRouteHandler(responses));

		// Auth bypass
		await page.addInitScript((u) => {
			window.__E2E_AUTH_USER__ = u;
		}, TEST_USER);

		// Helpers attached to page for test convenience
		page.apiOverrides = (overrides) => {
			Object.assign(responses, overrides);
		};

		page.apiFromFixtures = (overrides = {}) => {
			Object.assign(responses, {
				'/api/buscas': FIXTURE_API_BUSCAS,
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
				...overrides
			});
		};

		/**
		 * Carrega o dataset completo da Descobrir (dados realistas controlados).
		 * Cobre todos os cenários do checklist TESTES_DESCOBRIR.
		 * Aceita overrides para customizar endpoints específicos.
		 */
		page.apiFromDataset = (overrides = {}) => {
			const ds = DATASET_DESCOBRIR;
			Object.assign(responses, {
				'/api/candidatos': ds.candidatos_serum,
				'/api/lojas/novidades': ds.novidades_glory,
				'/api/lojas/registro': ds.lojas_registro,
				'/api/lojas/buscar': ds.lojas_buscar_glory,
				'/api/lojas/resolver': ds.resolver_loja,
				'/api/categorias': ds.categorias,
				'/api/buscas': ds.buscas_salvas,
				'/api/favoritos': ds.favoritos,
				...overrides
			});
		};

		await use(page);
	},

	// Backward compat — old tests use garimparPage
	garimparPage: async ({ authedPage }, use) => {
		await use(authedPage);
	}
});

export { expect };

// ── Legacy export (backward compat for tests that call mockApi directly) ─────

/**
 * @deprecated Use `page.apiOverrides()` instead. This exists for backward
 * compatibility with tests written before the fixture refactor.
 */
export async function mockApi(page, overrides = {}) {
	// If page already has route registered (via authedPage fixture), just update
	if (page.apiOverrides) {
		page.apiOverrides(overrides);
		return;
	}
	// Fallback: register route (standalone usage)
	const responses = { ...API_DEFAULTS, ...overrides };
	await page.route('**/api/**', createRouteHandler(responses));
}

/**
 * @deprecated Use `page.apiFromFixtures()` instead.
 */
export async function mockApiFromFixtures(page, overrides = {}) {
	if (page.apiFromFixtures) {
		page.apiFromFixtures(overrides);
		return;
	}
	const responses = {
		...API_DEFAULTS,
		'/api/buscas': FIXTURE_API_BUSCAS,
		'/api/candidatos': FIXTURE_API_CANDIDATOS,
		'/api/lojas/novidades': FIXTURE_API_NOVIDADES,
		...overrides
	};
	await page.route('**/api/**', createRouteHandler(responses));
}
