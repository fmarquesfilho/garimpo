/**
 * E2E: Fluxo "Adicionar Loja" com ResolveShop real.
 *
 * Testa o fluxo completo sem mocks:
 *   Frontend → C# API → Collector gRPC (ResolveShop) → Shopee API v4 → PostgreSQL
 *
 * Pré-requisitos:
 *   - API C# rodando (mise run up)
 *   - Collector Go rodando (go run ./services/collector/)
 *   - Firebase Auth Emulator rodando (porta 9099)
 *   - FIREBASE_AUTH_EMULATOR_HOST=localhost:9099
 *
 * Execução isolada:
 *   FIREBASE_AUTH_EMULATOR_HOST=localhost:9099 npx playwright test lojas-resolve-shop
 */
import { test, expect } from './fixtures.js';

test.describe('Lojas — ResolveShop E2E (sem mocks)', () => {
	// Marca como slow: depende de APIs externas (Shopee) e serviços locais
	test.slow();

	test('resolve link direto shopee.com.br → shop_id + shop_name', async ({ authedPage: page }) => {
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		// Intercepta a resposta do POST para validar o payload
		const responsePromise = page.waitForResponse(
			(resp) => resp.url().includes('/api/lojas') && resp.request().method() === 'POST'
		);

		await page.fill(
			'input[placeholder="Cole a URL da loja (shopee.com.br/shop/123) ou ID numérico"]',
			'https://shopee.com.br/belezanaweb_oficial'
		);
		await page.click('button:has-text("Adicionar")');

		const response = await responsePromise;
		expect(response.status()).toBe(200);

		const body = await response.json();
		expect(body.keyword).toBe('Beleza na Web Oficial');
		expect(body.shop_ids).toContain(1674883556);
		expect(body.status).toBe('adicionada');
	});

	test('resolve link curto s.shopee.com.br → shop_id + shop_name', async ({ authedPage: page }) => {
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		const responsePromise = page.waitForResponse(
			(resp) => resp.url().includes('/api/lojas') && resp.request().method() === 'POST'
		);

		await page.fill(
			'input[placeholder="Cole a URL da loja (shopee.com.br/shop/123) ou ID numérico"]',
			'https://s.shopee.com.br/70IKp57jnV'
		);
		await page.click('button:has-text("Adicionar")');

		const response = await responsePromise;
		expect(response.status()).toBe(200);

		const body = await response.json();
		expect(body.keyword).toBe('Glory of Seoul');
		expect(body.shop_ids).toContain(920292999);
		expect(body.status).toBe('adicionada');
	});

	test('resolve username puro → shop_id + shop_name', async ({ authedPage: page }) => {
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		const responsePromise = page.waitForResponse(
			(resp) => resp.url().includes('/api/lojas') && resp.request().method() === 'POST'
		);

		await page.fill(
			'input[placeholder="Cole a URL da loja (shopee.com.br/shop/123) ou ID numérico"]',
			'gloryofseoul.br'
		);
		await page.click('button:has-text("Adicionar")');

		const response = await responsePromise;
		expect(response.status()).toBe(200);

		const body = await response.json();
		expect(body.keyword).toBe('Glory of Seoul');
		expect(body.shop_ids).toContain(920292999);
	});

	test('loja adicionada aparece no GET /api/lojas com shop_ids', async ({ authedPage: page }) => {
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		// Intercepta o GET que carrega a lista
		const getResponse = await page.waitForResponse(
			(resp) => resp.url().includes('/api/lojas') && resp.request().method() === 'GET'
		);

		const data = await getResponse.json();
		// Deve haver pelo menos uma loja com shop_ids preenchido (dos testes anteriores ou do banco)
		const lojasComShopIds = data.lojas.filter((l) => l.shop_ids && l.shop_ids.length > 0);
		expect(lojasComShopIds.length).toBeGreaterThan(0);

		// Valida estrutura
		const loja = lojasComShopIds[0];
		expect(loja).toHaveProperty('id');
		expect(loja).toHaveProperty('keyword');
		expect(loja).toHaveProperty('shop_ids');
		expect(loja).toHaveProperty('ativo', true);
		expect(loja).toHaveProperty('criado_em');
	});

	test('link inválido retorna erro amigável', async ({ authedPage: page }) => {
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		const responsePromise = page.waitForResponse(
			(resp) => resp.url().includes('/api/lojas') && resp.request().method() === 'POST'
		);

		await page.fill(
			'input[placeholder="Cole a URL da loja (shopee.com.br/shop/123) ou ID numérico"]',
			'https://shopee.com.br/loja_que_nao_existe_xyz_999'
		);
		await page.click('button:has-text("Adicionar")');

		const response = await responsePromise;
		expect(response.status()).toBe(400);

		const body = await response.json();
		expect(body.error).toContain('não encontrada');
	});
});
