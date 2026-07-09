/**
 * E2E LOCAL da página Garimpar — roda sem Firebase/emulador/backend.
 * Prova que o harness (bypass de auth + API mockada) funciona e cobre os
 * fluxos básicos reportados. Novos cenários (agendamento, escopo de loja,
 * coleta→estatísticas) entram aqui como testes de regressão.
 */
import { test, expect, mockApi } from './fixtures.js';

test.describe('Garimpar — E2E local', () => {
	test('carrega autenticado (bypass), sem tela de login', async ({ garimparPage: page }) => {
		await mockApi(page);
		await page.goto('/');
		await expect(page.getByPlaceholder(/Buscar produto/i)).toBeVisible();
		await expect(page.getByRole('button', { name: /Entrar com Google/i })).toHaveCount(0);
	});

	test('busca por palavra-chave renderiza resultados', async ({ garimparPage: page }) => {
		await mockApi(page, {
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
				]
			}
		});
		await page.goto('/');
		await page.getByPlaceholder(/Buscar produto/i).fill('serum');
		await expect(page.getByText('Serum Vitamina C')).toBeVisible({ timeout: 10000 });
	});

	test('adicionar loja mostra o badge da loja resolvida', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/lojas': { id: 'loja-1', keyword: 'Le Botanic', shop_ids: [920292999], status: 'adicionada' }
		});
		await page.goto('/');
		const inputLoja = page.getByPlaceholder(/Adicionar loja/i);
		await inputLoja.fill('https://s.shopee.com.br/8fQYnxWQqu');
		await inputLoja.press('Enter');
		await expect(page.getByText('Le Botanic')).toBeVisible({ timeout: 10000 });
	});
});
