/**
 * E2E: Fluxo "Adicionar Loja" com ResolveShop real.
 *
 * Testa o fluxo completo sem mocks:
 *   Frontend (BuscaUnificada) → C# API → Collector gRPC (ResolveShop) → Shopee API v4
 *
 * Pré-requisitos:
 *   - API C# rodando (mise run up)
 *   - Collector Go rodando (go run ./services/collector/)
 *   - Firebase Auth Emulator rodando (porta 9099)
 *
 * Execução isolada:
 *   mise run test:e2e:lojas
 */
import { test, expect } from './fixtures.js';

test.describe('Lojas — ResolveShop E2E (sem mocks)', () => {
	test.slow();

	test('resolve link direto shopee.com.br → shop_id + shop_name', async ({ authedPage: page }) => {
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		const responsePromise = page.waitForResponse(
			(resp) => resp.url().includes('/api/lojas') && resp.request().method() === 'POST'
		);

		// Campo de loja no BuscaUnificada
		const lojaInput = page.locator('input[placeholder*="loja"]').first();
		await lojaInput.fill('https://shopee.com.br/belezanaweb_oficial');
		await lojaInput.press('Enter');

		const response = await responsePromise;
		expect(response.status()).toBe(200);

		const body = await response.json();
		expect(body.keyword).toBe('Beleza na Web Oficial');
		expect(body.shop_ids).toContain(1674883556);
		expect(body.status).toBe('adicionada');
	});

	test('resolve link curto s.shopee.com.br → shop_id + shop_name', async ({ authedPage: page }) => {
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		const responsePromise = page.waitForResponse(
			(resp) => resp.url().includes('/api/lojas') && resp.request().method() === 'POST'
		);

		const lojaInput = page.locator('input[placeholder*="loja"]').first();
		await lojaInput.fill('https://s.shopee.com.br/70IKp57jnV');
		await lojaInput.press('Enter');

		const response = await responsePromise;
		expect(response.status()).toBe(200);

		const body = await response.json();
		expect(body.keyword).toBe('Glory of Seoul');
		expect(body.shop_ids).toContain(920292999);
	});

	test('resolve username puro → shop_id + shop_name', async ({ authedPage: page }) => {
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		const responsePromise = page.waitForResponse(
			(resp) => resp.url().includes('/api/lojas') && resp.request().method() === 'POST'
		);

		const lojaInput = page.locator('input[placeholder*="loja"]').first();
		await lojaInput.fill('gloryofseoul.br');
		await lojaInput.press('Enter');

		const response = await responsePromise;
		expect(response.status()).toBe(200);

		const body = await response.json();
		expect(body.keyword).toBe('Glory of Seoul');
		expect(body.shop_ids).toContain(920292999);
	});

	test('loja adicionada aparece no GET /api/lojas com shop_ids', async ({ authedPage: page }) => {
		const response = await page.request.get('/api/lojas');
		expect(response.status()).toBe(200);

		const data = await response.json();
		const lojasComShopIds = data.lojas.filter((l) => l.shop_ids?.length > 0);
		expect(lojasComShopIds.length).toBeGreaterThan(0);

		const loja = lojasComShopIds[0];
		expect(loja).toHaveProperty('id');
		expect(loja).toHaveProperty('keyword');
		expect(loja).toHaveProperty('shop_ids');
		expect(loja).toHaveProperty('ativo', true);
	});

	test('link inválido retorna erro amigável', async ({ authedPage: page }) => {
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		const responsePromise = page.waitForResponse(
			(resp) => resp.url().includes('/api/lojas') && resp.request().method() === 'POST'
		);

		const lojaInput = page.locator('input[placeholder*="loja"]').first();
		await lojaInput.fill('https://shopee.com.br/loja_que_nao_existe_xyz_999');
		await lojaInput.press('Enter');

		const response = await responsePromise;
		expect(response.status()).toBe(400);

		const body = await response.json();
		expect(body.error).toContain('não encontrada');
	});
});
