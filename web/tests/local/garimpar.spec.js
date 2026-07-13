/**
 * E2E LOCAL — Fluxos basicos da pagina Garimpar.
 * Prova que o harness (auth bypass + API mock) funciona.
 */
import { test, expect } from './fixtures.js';
import { adicionarLojaViaOmnibox, waitForEngineReady, SEL } from './helpers.js';

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
		await page.locator(SEL.searchInput).press('Enter');
		await expect(page.getByText('Serum Vitamina C')).toBeVisible({ timeout: 10000 });
	});

	test('adicionar loja via Omnibox mostra chip', async ({ authedPage: page }) => {
		page.apiOverrides({
			'/api/lojas/resolver': {
				id: '920292999',
				nome: 'Le Botanic',
				nome_normalizado: 'lebotanic',
				marketplace: 'shopee',
				monitorada: false,
				imagem: null,
				seguidores: null,
				total_produtos: null,
				avaliacao: null
			}
		});
		await page.goto('/');
		await adicionarLojaViaOmnibox(page, 'https://s.shopee.com.br/8fQYnxWQqu');
		await expect(page.getByText('Le Botanic')).toBeVisible({ timeout: 10000 });
	});
});
