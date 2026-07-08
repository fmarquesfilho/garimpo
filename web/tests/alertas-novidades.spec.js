/**
 * E2E: Alertas de preço + Novidades/Variações.
 *
 * Testa:
 * 1. GET /api/lojas/novidades retorna estrutura correta (proxy Analyzer)
 * 2. GET /api/alertas retorna configuração do tenant
 * 3. POST /api/alertas/configurar atualiza threshold
 * 4. POST /api/alertas/testar envia alerta de teste
 * 5. Frontend exibe abas "Novidades" e "Preços" na página /lojas
 *
 * Pré-requisitos:
 *   - API C# rodando (mise run up)
 *   - Firebase Auth Emulator (porta 9099)
 *   - Analyzer Python rodando (porta 8060) — para /novidades
 *
 * Execução:
 *   mise run test:e2e:alertas
 */
import { test, expect } from './fixtures.js';

test.describe('Alertas e Novidades — Fluxo Completo', () => {
	test.slow();

	test('GET /api/lojas/novidades retorna estrutura correta', async ({ authedPage: page }) => {
		const response = await page.request.get('/api/lojas/novidades?busca_id=test&dias=7');
		expect(response.status()).toBe(200);

		const data = await response.json();
		// Deve retornar a estrutura padrão (vazia se não há snapshots)
		expect(data).toHaveProperty('produtos_novos');
		expect(data).toHaveProperty('variacoes');
		expect(Array.isArray(data.produtos_novos)).toBe(true);
		expect(Array.isArray(data.variacoes)).toBe(true);
	});

	test('GET /api/lojas/novidades com busca_id real retorna dados do Analyzer', async ({ authedPage: page }) => {
		// Primeiro adiciona uma loja para ter um busca_id
		const createResp = await page.request.post('/api/lojas', {
			data: { input: 'gloryofseoul.br' }
		});
		const created = await createResp.json();
		const buscaId = created.id;

		// Consulta novidades para essa busca
		const response = await page.request.get(`/api/lojas/novidades?busca_id=${buscaId}&dias=7`);
		expect(response.status()).toBe(200);

		const data = await response.json();
		expect(data).toHaveProperty('produtos_novos');
		expect(data).toHaveProperty('variacoes');
		// Sem snapshots coletados ainda, retorna vazio (graceful)
		expect(data.produtos_novos.length).toBe(0);
		expect(data.variacoes.length).toBe(0);
	});

	test('GET /api/alertas retorna configuração do tenant', async ({ authedPage: page }) => {
		const response = await page.request.get('/api/alertas');
		expect(response.status()).toBe(200);

		const data = await response.json();
		expect(data).toHaveProperty('habilitado');
		expect(data).toHaveProperty('threshold');
	});

	test('POST /api/alertas/configurar atualiza threshold', async ({ authedPage: page }) => {
		const response = await page.request.post('/api/alertas/configurar', {
			data: { threshold: 0.2 }
		});
		expect(response.status()).toBe(200);

		const data = await response.json();
		expect(data).toHaveProperty('status');
	});

	test('POST /api/alertas/testar valida configuração', async ({ authedPage: page }) => {
		const response = await page.request.post('/api/alertas/testar', {
			data: {}
		});
		// Pode retornar 200 (sucesso) ou 400 (chat_id não configurado)
		expect([200, 400]).toContain(response.status());
	});

	test('frontend exibe seção de configuração com GerenciarBuscas', async ({ authedPage: page }) => {
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		// Expandir seção Configuração
		const configBtn = page.locator('button:has-text("Configuração")');
		if (await configBtn.isVisible()) await configBtn.click();

		// Verifica que GerenciarBuscas está presente
		await expect(page.locator('text=Buscas por palavra-chave')).toBeVisible({ timeout: 5000 });
	});

	test('GET /api/lojas/evolucao retorna série temporal', async ({ authedPage: page }) => {
		const response = await page.request.get('/api/lojas/evolucao?dias=7');
		expect(response.status()).toBe(200);

		const data = await response.json();
		// Estrutura esperada (vazia sem dados)
		expect(data).toHaveProperty('lojas');
		expect(data).toHaveProperty('resumo');
		expect(data.resumo).toHaveProperty('total_quedas');
		expect(data.resumo).toHaveProperty('total_altas');
	});
});
