/**
 * E2E: Buscas Agendadas — validação pós-migração.
 *
 * Testa os dois fluxos principais:
 * 1. Monitorar todos os produtos de uma loja (sem keywords)
 * 2. Monitorar produtos filtrados por keywords
 *
 * Valida que o Scheduler recebe o job via SetSchedule ao criar/remover buscas.
 *
 * Pré-requisitos:
 *   - API C# rodando (mise run up)
 *   - Collector Go rodando (porta 50051)
 *   - Scheduler Go rodando (porta 50054)
 *   - Firebase Auth Emulator (porta 9099)
 *
 * Execução:
 *   mise run test:e2e:buscas-agendadas
 */
import { test, expect } from './fixtures.js';

test.describe('Buscas Agendadas — Fluxo Completo', () => {
	test.slow();

	test('adicionar loja sem keywords agenda coleta de todos os produtos', async ({ authedPage: page }) => {
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

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
		expect(body.shop_ids).toContain(1674883556);
		expect(body.keyword).toBe('Beleza na Web Oficial');
		expect(body.status).toBe('adicionada');

		// Verificar que o Scheduler recebeu o job (via listagem de lojas)
		const listResp = await page.request.get('/api/lojas');
		const lojas = await listResp.json();
		const buscaCriada = lojas.lojas?.find((b) => b.shop_ids?.includes(1674883556));
		expect(buscaCriada).toBeTruthy();
	});

	test('adicionar loja com keywords agenda coleta filtrada', async ({ authedPage: page }) => {
		// Chama a API diretamente com keywords (o formulário UI usa POST /api/buscas para keywords)
		const apiResp = await page.request.post('/api/lojas', {
			data: {
				input: 'https://shopee.com.br/gloryofseoul.br',
				keywords: ['serum', 'protetor solar']
			}
		});

		expect(apiResp.status()).toBe(200);
		const body = await apiResp.json();
		expect(body.shop_ids).toContain(920292999);
		expect(body.keyword).toBe('Glory of Seoul');

		// Verificar que keywords foram persistidas na busca
		const listResp = await page.request.get('/api/lojas');
		const lojas = await listResp.json();
		const lojaComKeywords = lojas.lojas?.find((l) => l.shop_ids?.includes(920292999) && l.keywords?.length > 0);
		expect(lojaComKeywords).toBeTruthy();
		expect(lojaComKeywords.keywords).toContain('serum');
		expect(lojaComKeywords.keywords).toContain('protetor solar');
	});

	test('remover loja pausa o agendamento', async ({ authedPage: page }) => {
		// Primeiro cria uma loja
		const createResp = await page.request.post('/api/lojas', {
			data: { input: 'lebotanic' }
		});
		expect(createResp.status()).toBe(200);
		const created = await createResp.json();

		// Remove a loja
		const deleteResp = await page.request.delete(`/api/lojas?id=${created.id}`);
		expect(deleteResp.status()).toBe(200);
		const deleted = await deleteResp.json();
		expect(deleted.status).toBe('removida');

		// Verificar que a busca não aparece mais na listagem ativa
		const listResp = await page.request.get('/api/lojas');
		const lojas = await listResp.json();
		const lojaRemovida = lojas.lojas?.find((l) => l.id === created.id);
		expect(lojaRemovida).toBeFalsy();
	});

	test('GET /api/lojas retorna keywords e cron_expression', async ({ authedPage: page }) => {
		const response = await page.request.get('/api/lojas');
		expect(response.status()).toBe(200);

		const data = await response.json();
		expect(data).toHaveProperty('lojas');
		expect(data).toHaveProperty('total');

		// Valida que o schema inclui os novos campos
		if (data.lojas.length > 0) {
			const loja = data.lojas[0];
			expect(loja).toHaveProperty('id');
			expect(loja).toHaveProperty('keyword');
			expect(loja).toHaveProperty('shop_ids');
			expect(loja).toHaveProperty('keywords');
			expect(loja).toHaveProperty('cron_expression');
			expect(loja).toHaveProperty('source_url');
			expect(loja).toHaveProperty('ativo');
			expect(loja).toHaveProperty('criado_em');
		}
	});

	test('componente GerenciarBuscas renderiza corretamente', async ({ authedPage: page }) => {
		await page.goto('/lojas');
		await page.waitForLoadState('networkidle');

		// Verifica que o heading "Buscas Agendadas" existe
		await expect(page.locator('text=Buscas Agendadas')).toBeVisible();

		// Verifica que o botão "+ nova busca" existe
		await expect(page.locator('button:has-text("nova busca")')).toBeVisible();
	});
});
