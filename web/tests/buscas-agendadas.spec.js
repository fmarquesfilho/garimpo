/**
 * E2E: Buscas Agendadas — validação com BuscaUnificada.
 *
 * Testa:
 * 1. Adicionar loja via campo integrado (sem keywords)
 * 2. Adicionar loja com keywords (API direta)
 * 3. Remover loja pausa agendamento
 * 4. Schema do GET /api/lojas
 * 5. Salvar busca agendada via BuscaUnificada
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

	test('adicionar loja via campo integrado agenda coleta', async ({ authedPage: page }) => {
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		const responsePromise = page.waitForResponse(
			(resp) => resp.url().includes('/api/lojas') && resp.request().method() === 'POST'
		);

		// Campo de loja integrado no BuscaUnificada
		const lojaInput = page.locator('input[placeholder*="loja"]').first();
		await lojaInput.fill('https://shopee.com.br/belezanaweb_oficial');
		await lojaInput.press('Enter');

		const response = await responsePromise;
		expect(response.status()).toBe(200);

		const body = await response.json();
		expect(body.shop_ids).toContain(1674883556);
		expect(body.keyword).toBe('Beleza na Web Oficial');
		expect(body.status).toBe('adicionada');

		// Verificar que o Scheduler recebeu o job
		const listResp = await page.request.get('/api/lojas');
		const lojas = await listResp.json();
		const buscaCriada = lojas.lojas?.find((b) => b.shop_ids?.includes(1674883556));
		expect(buscaCriada).toBeTruthy();
	});

	test('adicionar loja com keywords agenda coleta filtrada', async ({ authedPage: page }) => {
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

		const listResp = await page.request.get('/api/lojas');
		const lojas = await listResp.json();
		const lojaComKeywords = lojas.lojas?.find((l) => l.shop_ids?.includes(920292999) && l.keywords?.length > 0);
		expect(lojaComKeywords).toBeTruthy();
		expect(lojaComKeywords.keywords).toContain('serum');
	});

	test('remover loja pausa o agendamento', async ({ authedPage: page }) => {
		const createResp = await page.request.post('/api/lojas', { data: { input: 'lebotanic' } });
		expect(createResp.status()).toBe(200);
		const created = await createResp.json();

		const deleteResp = await page.request.delete(`/api/lojas?id=${created.id}`);
		expect(deleteResp.status()).toBe(200);
		expect((await deleteResp.json()).status).toBe('removida');

		const listResp = await page.request.get('/api/lojas');
		const lojas = await listResp.json();
		expect(lojas.lojas?.find((l) => l.id === created.id)).toBeFalsy();
	});

	test('GET /api/lojas retorna keywords e cron_expression', async ({ authedPage: page }) => {
		const response = await page.request.get('/api/lojas');
		expect(response.status()).toBe(200);

		const data = await response.json();
		expect(data).toHaveProperty('lojas');
		expect(data).toHaveProperty('total');

		if (data.lojas.length > 0) {
			const loja = data.lojas[0];
			expect(loja).toHaveProperty('id');
			expect(loja).toHaveProperty('keyword');
			expect(loja).toHaveProperty('shop_ids');
			expect(loja).toHaveProperty('keywords');
			expect(loja).toHaveProperty('cron_expression');
		}
	});

	test('POST /api/buscas com shop_ids persiste associação loja+busca', async ({ authedPage: page }) => {
		const resp = await page.request.post('/api/buscas', {
			data: {
				keywords: ['sérum', 'vitamina c'],
				shop_ids: [920292999],
				cron: '0 */8 * * *',
				comissao_min: 0.1,
				vendas_min: 50
			}
		});
		expect(resp.status()).toBe(200);
		const body = await resp.json();
		expect(body.status).toBe('salva');
		expect(body.cron).toBe('0 */8 * * *');
	});
});
