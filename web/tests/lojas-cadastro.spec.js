import { test, expect } from './fixtures.js';

test.describe('Lojas — Cadastro via BuscaUnificada', () => {
	test('adiciona loja e exibe como tag no componente', async ({ authedPage: page }) => {
		let chamouAdicionar = false;

		await page.route('**/api/buscas', async (route) => {
			if (route.request().method() === 'GET') {
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({ buscas: [], total: 0 })
				});
			} else {
				await route.fallback();
			}
		});

		await page.route('**/api/lojas', async (route) => {
			if (route.request().method() === 'POST') {
				chamouAdicionar = true;
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
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ estrategia: 'nicho', candidatos: [], total_bruto: 0 })
			});
		});
		await page.route('**/api/favoritos*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ favoritos: [] })
			});
		});
		await page.route('**/api/admin/me', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ admin: false })
			});
		});

		await page.goto('/');
		await page.waitForLoadState('networkidle');

		// Preenche campo de loja integrado no BuscaUnificada
		const lojaInput = page.locator('input[placeholder*="loja"]').first();
		await lojaInput.fill('123456789');
		await lojaInput.press('Enter');

		// Espera resolução
		await page.waitForTimeout(1000);

		// Verifica que chamou POST /api/lojas
		expect(chamouAdicionar).toBe(true);

		// Verifica que a loja apareceu como tag (badge)
		await expect(page.locator('text=Loja Nova Teste')).toBeVisible({ timeout: 5000 });
	});
});
