import { test, expect } from './fixtures.js';

/**
 * Testes E2E do fluxo de monitoramento de lojas na página unificada.
 *
 * Valida:
 * - Toggle 🏪 Lojas exibe produtos das lojas monitoradas
 * - Seletor de loja filtra por loja específica
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
			categoria: 'cosmeticos',
			estrategia: 'nicho',
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
			nome: 'Jean Paul Gaultier Le Male 125ml',
			preco: 280.0,
			comissao: 0.13,
			vendas: 450,
			nota: 4.5,
			imagem: 'https://cf.shopee.com.br/file/jpg-le-male.jpg',
			link: 'https://shopee.com.br/product/123456/SP-006',
			loja: 'ImportsPerfumaria',
			detectado_em: '2026-07-04T00:00:00Z'
		}
	],
	variacoes: [
		{
			produto_id: 'SP-001',
			nome: 'Perfume CK One 100ml EDT',
			preco_anterior: 189.9,
			preco_atual: 151.9,
			variacao: -0.2001,
			variacao_pct: -0.2001,
			imagem: 'https://cf.shopee.com.br/file/ck-one-100ml.jpg',
			link: 'https://shopee.com.br/product/123456/SP-001',
			loja: 'ImportsPerfumaria',
			detectado_em: '2026-07-04T00:00:00Z'
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
			preco: 194.0,
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

async function interceptarAPIs(page, { novidadesResponse = mockNovidades } = {}) {
	await page.route('**/api/buscas', async (route) => {
		await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(mockBuscas) });
	});

	await page.route('**/api/lojas/novidades*', async (route) => {
		await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(novidadesResponse) });
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

	await page.route('**/api/alertas*', async (route) => {
		await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ regras: [] }) });
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

test.describe('Fonte Lojas — toggle e seleção na página unificada', () => {
	test('toggle 🏪 Lojas exibe produtos das lojas monitoradas', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		// Ativar fonte Lojas
		const toggleLojas = page.locator('button:has-text("🏪 Lojas")');
		await toggleLojas.click();
		await page.waitForTimeout(800);

		// Deve exibir produtos da loja
		await expect(page.locator('.grade')).toBeVisible({ timeout: 5000 });
		await expect(page.locator('text=Perfume CK One 100ml')).toBeVisible();
	});

	test('seletor de loja aparece quando toggle ativo', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		// Ativar fonte Lojas
		await page.locator('button:has-text("🏪 Lojas")').click();
		await page.waitForTimeout(500);

		// Seletor com "Todas" e "Glory of Seoul" deve aparecer
		await expect(page.locator('button:has-text("Todas")')).toBeVisible();
		await expect(page.locator('button:has-text("Glory of Seoul")')).toBeVisible();
	});

	test('selecionar loja específica filtra resultados', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		await page.locator('button:has-text("🏪 Lojas")').click();
		await page.waitForTimeout(500);

		// Clicar em "Glory of Seoul" no seletor
		await page.locator('button:has-text("Glory of Seoul")').click();
		await page.waitForTimeout(500);

		// Resultados devem ser da loja selecionada
		const cards = page.locator('.grade > *');
		expect(await cards.count()).toBeGreaterThan(0);
	});

	test('quedas aparecem no toggle 📉 Quedas com badge', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/');
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(800);

		// O toggle Quedas deve ter badge (loja monitorada gera quedas)
		const badgeQuedas = page.locator('button:has-text("📉 Quedas") .fonte-badge');
		await expect(badgeQuedas).toBeVisible({ timeout: 5000 });
	});

	test('novos aparecem no toggle 🆕 Novos', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/');
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(800);

		// O toggle Novos deve ter badge
		const badgeNovos = page.locator('button:has-text("🆕 Novos") .fonte-badge');
		await expect(badgeNovos).toBeVisible({ timeout: 5000 });
	});
});

test.describe('Fonte Lojas — graceful degradation', () => {
	test('erro no analyzer não crasha a página', async ({ authedPage: page }) => {
		await page.route('**/api/lojas/novidades*', async (route) => {
			await route.abort('timedout');
		});
		await page.route('**/api/buscas', async (route) => {
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

	test('sem lojas monitoradas + toggle 🏪 ativo mostra empty state', async ({ authedPage: page }) => {
		// Mock sem lojas
		await page.route('**/api/buscas', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ buscas: [], total: 0 })
			});
		});
		await page.route('**/api/candidatos*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ estrategia: 'nicho', total_bruto: 0, candidatos: [] })
			});
		});
		await page.route('**/api/favoritos*', async (route) => {
			await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ favoritos: [] }) });
		});
		await page.route('**/api/admin/me', async (route) => {
			await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ admin: false }) });
		});

		await page.goto('/');
		await page.waitForLoadState('networkidle');

		// Desativar todas as fontes e ativar só Lojas
		const fontes = page.locator('.fonte-btn.ativa');
		const count = await fontes.count();
		for (let i = 0; i < count; i++) {
			await fontes.nth(0).click();
			await page.waitForTimeout(100);
		}
		await page.locator('button:has-text("🏪 Lojas")').click();
		await page.waitForTimeout(800);

		// Deve mostrar mensagem de nenhuma loja
		await expect(page.locator('text=Nenhuma loja monitorada')).toBeVisible({ timeout: 5000 });
	});
});
