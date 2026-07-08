import { test, expect } from './fixtures.js';

/**
 * Testes E2E do fluxo de monitoramento de lojas na página unificada.
 *
 * Valida:
 * - Toggle 🏪 Lojas ativa a fonte e carrega produtos das lojas monitoradas
 * - Quedas e novos das lojas aparecem nos respectivos toggles
 * - Graceful degradation quando analyzer está offline
 */

// ── Dados mock ────────────────────────────────────────────────────────────

const mockBuscas = {
	buscas: [
		{
			id: 'loja-920292999',
			nome: 'Glory of Seoul',
			keywords: ['cosmeticos-coreanos'],
			shop_ids: [920292999],
			cron: '0 */8 * * *',
			ativo: true,
			criado_em: '2026-07-01T00:00:00Z'
		}
	],
	total: 1
};

const mockNovidades = {
	busca_id: 'loja-920292999',
	dias: 7,
	produtos_novos: [
		{
			produto_id: 'SP-006',
			nome: 'Jean Paul Gaultier Le Male',
			preco: 280,
			comissao: 0.13,
			vendas: 450,
			imagem: '',
			link: '',
			loja: 'Glory of Seoul',
			detectado_em: '2026-07-04'
		}
	],
	variacoes: [
		{
			produto_id: 'SP-001',
			nome: 'Perfume CK One 100ml EDT',
			preco_anterior: 189.9,
			preco_atual: 151.9,
			variacao: -0.2,
			variacao_pct: -0.2,
			imagem: '',
			link: '',
			loja: 'Glory of Seoul',
			detectado_em: '2026-07-04'
		}
	],
	total_novos: 1,
	total_variacoes: 1
};

const mockCandidatos = {
	estrategia: 'nicho',
	total_bruto: 2,
	candidatos: [
		{
			id: 'SP-001',
			nome: 'Perfume CK One 100ml',
			preco: 151.9,
			comissao: 0.12,
			vendas: 3520,
			avaliacao: 4.8,
			loja: 'Glory of Seoul',
			imagem: '',
			link: ''
		},
		{
			id: 'SP-002',
			nome: 'D&G Light Blue 75ml',
			preco: 194,
			comissao: 0.1,
			vendas: 1920,
			avaliacao: 4.9,
			loja: 'Glory of Seoul',
			imagem: '',
			link: ''
		}
	]
};

// ── Helpers ───────────────────────────────────────────────────────────────

async function interceptarAPIs(page) {
	await page.route('**/api/buscas*', async (route) => {
		if (route.request().method() === 'GET') {
			await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(mockBuscas) });
		} else {
			await route.fulfill({ status: 200, contentType: 'application/json', body: '{}' });
		}
	});
	await page.route('**/api/lojas/novidades*', async (route) => {
		await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(mockNovidades) });
	});
	await page.route('**/api/candidatos*', async (route) => {
		await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(mockCandidatos) });
	});
	await page.route('**/api/favoritos*', async (route) => {
		await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ favoritos: [] }) });
	});
	await page.route('**/api/admin/me', async (route) => {
		await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ admin: false }) });
	});
	await page.route('**/api/lojas', async (route) => {
		if (route.request().method() === 'GET') {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ lojas: mockBuscas.buscas, total: 1 })
			});
		} else {
			await route.fulfill({ status: 200, contentType: 'application/json', body: '{}' });
		}
	});
}

// ── Testes ────────────────────────────────────────────────────────────────

test.describe('Fonte Lojas — toggle e resultados na página unificada', () => {
	test('toggle 🏪 Lojas existe e é clicável', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		const toggleLojas = page.locator('button:has-text("🏪 Lojas")');
		await expect(toggleLojas).toBeVisible({ timeout: 10000 });
		await toggleLojas.click();
		// Toggle deve ficar ativo (data-state=on via Bits UI)
		await expect(toggleLojas).toHaveAttribute('data-state', 'on');
	});

	test('quedas aparecem no toggle 📉 Quedas com badge', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/');
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(3000);

		const badgeQuedas = page.locator('button:has-text("📉 Quedas") .fonte-badge');
		await expect(badgeQuedas).toBeVisible({ timeout: 15000 });
	});

	test('novos aparecem no toggle 🆕 Novos com badge', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/');
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(3000);

		const badgeNovos = page.locator('button:has-text("🆕 Novos") .fonte-badge');
		await expect(badgeNovos).toBeVisible({ timeout: 15000 });
	});
});

test.describe('Fonte Lojas — graceful degradation', () => {
	test('erro no analyzer não crasha a página', async ({ authedPage: page }) => {
		await page.route('**/api/lojas/novidades*', async (route) => {
			await route.abort('timedout');
		});
		await page.route('**/api/buscas*', async (route) => {
			await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(mockBuscas) });
		});
		await page.route('**/api/candidatos*', async (route) => {
			await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(mockCandidatos) });
		});
		await page.route('**/api/favoritos*', async (route) => {
			await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ favoritos: [] }) });
		});
		await page.route('**/api/admin/me', async (route) => {
			await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ admin: false }) });
		});

		const errors = [];
		page.on('pageerror', (err) => errors.push(err.message));

		await page.goto('/');
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(2000);

		expect(errors).toHaveLength(0);
	});
});
