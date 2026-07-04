import { test, expect } from './fixtures.js';

/**
 * Testes E2E do fluxo de variação de preços.
 *
 * Valida a página /lojas com aba "📉 Preços":
 * - Exibe variações de preço (quedas e altas)
 * - Exibe produtos novos na aba "🆕 Novidades"
 * - Botão "📤 Publicar" navega com dados preenchidos
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
		},
		{
			produto_id: 'SP-002',
			nome: 'Dolce & Gabbana Light Blue 75ml',
			preco_anterior: 299.0,
			preco_atual: 194.0,
			variacao: -0.3512,
			variacao_pct: -0.3512,
			imagem: 'https://cf.shopee.com.br/file/dg-light-blue.jpg',
			link: 'https://shopee.com.br/product/123456/SP-002',
			loja: 'ImportsPerfumaria',
			detectado_em: '2026-07-04T00:00:00Z'
		},
		{
			produto_id: 'SP-004',
			nome: 'Carolina Herrera Good Girl 80ml',
			preco_anterior: 420.0,
			preco_atual: 462.0,
			variacao: 0.1,
			variacao_pct: 0.1,
			imagem: 'https://cf.shopee.com.br/file/ch-good-girl.jpg',
			link: 'https://shopee.com.br/product/123456/SP-004',
			loja: 'ImportsPerfumaria',
			detectado_em: '2026-07-04T00:00:00Z'
		}
	],
	total_novos: 1,
	total_variacoes: 3
};

const mockNovidadesVazio = {
	busca_id: '',
	dias: 7,
	produtos_novos: [],
	variacoes: [],
	total_novos: 0,
	total_variacoes: 0
};

const mockCandidatos = {
	estrategia: 'nicho',
	total_bruto: 2,
	candidatos: [
		{ id: 'SP-001', nome: 'Perfume CK One 100ml', preco: 151.9, comissao: 0.12, vendas: 3520, avaliacao: 4.8, loja: 'ImportsPerfumaria', imagem: '', link: '' },
		{ id: 'SP-002', nome: 'D&G Light Blue 75ml', preco: 194.0, comissao: 0.1, vendas: 1920, avaliacao: 4.9, loja: 'ImportsPerfumaria', imagem: '', link: '' }
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

test.describe('Lojas — Aba Preços (variação de preços)', () => {
	test('exibe lojas monitoradas após login', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		// Deve mostrar a loja "Glory of Seoul"
		await expect(page.locator('text=Glory of Seoul')).toBeVisible({ timeout: 10000 });
	});

	test('selecionar loja mostra abas (Produtos, Novidades, Preços)', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		// Clicar na loja
		await page.locator('text=Glory of Seoul').click();
		await page.waitForTimeout(500);

		// Abas devem estar visíveis (usar role tab para ser específico)
		await expect(page.getByRole('tab', { name: /Produtos/ })).toBeVisible();
		await expect(page.getByRole('tab', { name: /Novidades/ })).toBeVisible();
		await expect(page.getByRole('tab', { name: /Preços/ })).toBeVisible();
	});

	test('aba Preços exibe tabela de variações', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		// Selecionar loja
		await page.locator('text=Glory of Seoul').click();
		await page.waitForTimeout(500);

		// Clicar na aba Preços
		await page.locator('text=📉 Preços').click();
		await page.waitForTimeout(500);

		// Deve exibir a tabela com variações
		await expect(page.locator('table')).toBeVisible({ timeout: 5000 });

		// Deve ter os cabeçalhos corretos
		await expect(page.locator('th', { hasText: 'Produto' })).toBeVisible();
		await expect(page.locator('th', { hasText: 'Antes' })).toBeVisible();
		await expect(page.locator('th', { hasText: 'Agora' })).toBeVisible();
		await expect(page.locator('th', { hasText: 'Variação' })).toBeVisible();
	});

	test('aba Preços mostra contagem de variações', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		await page.locator('text=Glory of Seoul').click();
		await page.waitForTimeout(500);
		await page.getByRole('tab', { name: /Preços/ }).click();
		await page.waitForTimeout(500);

		// Deve mostrar "3 variação(ões)"
		await expect(page.locator('text=/3.*variação/')).toBeVisible();
	});

	test('quedas de preço aparecem em verde com seta ↓', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		await page.locator('text=Glory of Seoul').click();
		await page.waitForTimeout(500);
		await page.locator('text=📉 Preços').click();
		await page.waitForTimeout(500);

		// Deve ter badge com seta para baixo (queda)
		const quedas = page.locator('span:has-text("↓")');
		expect(await quedas.count()).toBeGreaterThanOrEqual(2);
	});

	test('altas de preço aparecem com seta ↑', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		await page.locator('text=Glory of Seoul').click();
		await page.waitForTimeout(500);
		await page.locator('text=📉 Preços').click();
		await page.waitForTimeout(500);

		// Carolina Herrera +10% — deve ter seta para cima
		const altas = page.locator('span:has-text("↑")');
		expect(await altas.count()).toBeGreaterThanOrEqual(1);
	});

	test('botão publicar navega para /publicar com dados', async ({ authedPage: page }) => {
		await interceptarAPIs(page);

		// Interceptar navegação para /publicar
		await page.route('**/api/destinos*', async (route) => {
			await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ destinos: [] }) });
		});
		await page.route('**/api/templates*', async (route) => {
			await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ templates: [] }) });
		});

		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		await page.locator('text=Glory of Seoul').click();
		await page.waitForTimeout(500);
		await page.locator('text=📉 Preços').click();
		await page.waitForTimeout(500);

		// Clicar no botão 📤 da primeira variação
		const btnPublicar = page.locator('button[title="Publicar esta oferta"]').first();
		await expect(btnPublicar).toBeVisible();
		await btnPublicar.click();

		// Deve navegar para /publicar
		await page.waitForURL('**/publicar*', { timeout: 5000 });
		expect(page.url()).toContain('/publicar');
	});

	test('aba Novidades exibe produtos novos', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		await page.locator('text=Glory of Seoul').click();
		await page.waitForTimeout(500);

		// Clicar na aba Novidades
		await page.getByRole('tab', { name: /Novidades/ }).click();
		await page.waitForTimeout(500);

		// Deve exibir o produto novo
		await expect(page.locator('text=Jean Paul Gaultier Le Male 125ml')).toBeVisible({ timeout: 5000 });
	});

	test('badge na aba Preços mostra contagem', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		await page.locator('text=Glory of Seoul').click();
		await page.waitForTimeout(1000);

		// A aba Preços deve ter um badge com a contagem "3"
		const abaPrecos = page.locator('button', { hasText: '📉 Preços' });
		await expect(abaPrecos).toBeVisible();
		// O badge é renderizado dentro do tab — verificar que o texto contém o número
		await expect(abaPrecos).toContainText('3');
	});
});

test.describe('Lojas — Graceful degradation (analyzer offline)', () => {
	test('sem variações mostra mensagem amigável', async ({ authedPage: page }) => {
		await interceptarAPIs(page, { novidadesResponse: mockNovidadesVazio });
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		await page.locator('text=Glory of Seoul').click();
		await page.waitForTimeout(500);
		await page.locator('text=📉 Preços').click();
		await page.waitForTimeout(500);

		// Deve mostrar mensagem de "nenhuma variação"
		await expect(page.locator('text=Nenhuma variação de preço')).toBeVisible();
	});

	test('sem produtos novos mostra mensagem amigável', async ({ authedPage: page }) => {
		await interceptarAPIs(page, { novidadesResponse: mockNovidadesVazio });
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		await page.locator('text=Glory of Seoul').click();
		await page.waitForTimeout(500);
		await page.locator('text=🆕 Novidades').click();
		await page.waitForTimeout(500);

		await expect(page.locator('text=Nenhum produto novo')).toBeVisible();
	});

	test('erro no analyzer não crasha a página', async ({ authedPage: page }) => {
		// Simular timeout do analyzer
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
		await page.route('**/api/alertas*', async (route) => {
			await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ regras: [] }) });
		});
		await page.route('**/api/lojas', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ lojas: mockBuscas.buscas, total: 1 })
			});
		});

		const errors = [];
		page.on('pageerror', (err) => errors.push(err.message));

		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		await page.locator('text=Glory of Seoul').click();
		await page.waitForTimeout(2000);

		// Página não deve ter crashado
		expect(errors).toHaveLength(0);
		// A aba Preços ainda deve existir
		await expect(page.locator('text=📉 Preços')).toBeVisible();
	});
});
