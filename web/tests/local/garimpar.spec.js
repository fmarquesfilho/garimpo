/**
 * E2E LOCAL — Fluxos básicos da página Garimpar.
 * Prova que o harness (auth bypass + API mock) funciona.
 */
import { test, expect } from './fixtures.js';
import { abrirRaiaLojas, adicionarLoja, waitForEngineReady, SEL } from './helpers.js';

test.describe('Garimpar — E2E local', () => {
	test('carrega autenticado (bypass), sem tela de login', async ({ authedPage: page }) => {
		await page.goto('/');
		await expect(page.locator(SEL.searchInput)).toBeVisible();
		await expect(page.getByRole('button', { name: /Entrar com Google/i })).toHaveCount(0);
	});

	test('busca por palavra-chave renderiza resultados', async ({ authedPage: page }) => {
		page.apiOverrides({
			'/api/candidatos': {
				candidatos: [
					{
						id: 'p1',
						produto_id: 'p1',
						nome: 'Serum Vitamina C',
						preco: 79.9,
						comissao: 0.12,
						vendas: 100,
						loja: 'Loja X',
						link: 'https://x'
					}
				],
				total_bruto: 1,
				estrategia: 'nicho'
			}
		});
		await page.goto('/');
		await page.locator(SEL.searchInput).fill('serum');
		await expect(page.getByText('Serum Vitamina C')).toBeVisible({ timeout: 10000 });
	});

	test('adicionar loja mostra badge', async ({ authedPage: page }) => {
		page.apiOverrides({
			'/api/lojas': { id: 'loja-1', keyword: 'Le Botanic', shop_ids: [920292999], status: 'adicionada' }
		});
		await page.goto('/');
		await abrirRaiaLojas(page);
		await adicionarLoja(page, 'https://s.shopee.com.br/8fQYnxWQqu');
		await expect(page.getByText('Le Botanic')).toBeVisible({ timeout: 10000 });
	});
});
