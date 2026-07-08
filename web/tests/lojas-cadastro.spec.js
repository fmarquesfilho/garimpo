import { test, expect } from './fixtures.js';

test.describe('Lojas — Cadastro e Agendamento', () => {
	test('adiciona loja e exibe na lista de buscas', async ({ authedPage: page }) => {
		let chamouAdicionar = false;
		let bodyEnviado = null;

		// Mock das APIs
		await page.route('**/api/buscas', async (route) => {
			if (route.request().method() === 'GET') {
				// Na primeira chamada (antes de adicionar), retorna lista vazia
				if (!chamouAdicionar) {
					await route.fulfill({
						status: 200,
						contentType: 'application/json',
						body: JSON.stringify({ buscas: [], total: 0 })
					});
				} else {
					// Após adicionar, retorna a loja recém cadastrada
					await route.fulfill({
						status: 200,
						contentType: 'application/json',
						body: JSON.stringify({
							buscas: [
								{
									id: 'loja-123456789',
									nome: 'Loja Nova Teste',
									shop_ids: [123456789],
									ativo: true,
									criado_em: '2026-07-01T00:00:00Z'
								}
							],
							total: 1
						})
					});
				}
			} else {
				await route.fallback();
			}
		});

		await page.route('**/api/lojas', async (route) => {
			if (route.request().method() === 'POST') {
				chamouAdicionar = true;
				bodyEnviado = route.request().postDataJSON();
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({
						id: 'loja-123456789',
						keyword: 'Loja Nova Teste',
						shop_ids: [123456789],
						status: 'adicionada'
					})
				});
			} else {
				await route.fallback();
			}
		});

		await page.route('**/api/lojas/novidades*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ produtos_novos: [], variacoes: [] })
			});
		});
		await page.route('**/api/candidatos*', async (route) => {
			await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ candidatos: [] }) });
		});

		await page.goto('/');
		await page.waitForLoadState('networkidle');
		// Expandir seção Configuração
		const configBtn = page.locator('button:has-text("Configuração")');
		if (await configBtn.isVisible()) await configBtn.click();

		// Preenche formulário de Adicionar Loja
		await page.fill('input[placeholder="Cole a URL da loja (shopee.com.br/shop/123) ou ID numérico"]', '123456789');
		await page.click('button:has-text("Adicionar")');

		// Espera a mensagem de sucesso (usa o nome resolvido da loja)
		await expect(page.getByText('Loja "Loja Nova Teste" adicionada com sucesso!')).toBeVisible();

		// Verifica se a loja apareceu na lista
		await expect(page.locator('button:has-text("Loja Nova Teste")')).toBeVisible();

		// Verifica que os parâmetros corretos foram enviados no POST
		expect(bodyEnviado.input).toBe('123456789');
		// Loja monitorada agenda coleta periódica por padrão (a cada 8h)
		expect(bodyEnviado.cron).toBe('0 */8 * * *');
	});
});
