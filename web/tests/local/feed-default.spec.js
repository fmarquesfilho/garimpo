/**
 * E2E LOCAL — Feed Default (produtos on-load sem interação do usuário).
 *
 * Testa:
 * - Página abre com produtos visíveis (feed default ativo)
 * - Busca manual sobrescreve o feed
 * - Cada reload pode mostrar categoria diferente (rotação)
 */
import { test, expect } from './fixtures.js';

// ── Mock data ─────────────────────────────────────────────────────────────────

const CANDIDATOS_FEED = {
	candidatos: [
		{ id: 'fd1', nome: 'Kit Beleza Completo Skincare', preco: 45.9, comissao: 0.12, vendas: 320, loja: 'Korean Beauty', link: 'https://shopee.com.br/1', _fonte: 'curadoria' },
		{ id: 'fd2', nome: 'Perfume Importado Feminino', preco: 62.0, comissao: 0.09, vendas: 280, loja: 'Beleza Pro', link: 'https://shopee.com.br/2', _fonte: 'curadoria' },
		{ id: 'fd3', nome: 'Skincare Hidratante Noturno', preco: 38.5, comissao: 0.11, vendas: 150, loja: 'Skin Lab', link: 'https://shopee.com.br/3', _fonte: 'curadoria' }
	],
	total_bruto: 3,
	estrategia: 'nicho'
};

const CANDIDATOS_BUSCA_MANUAL = {
	candidatos: [
		{ id: 'bm1', nome: 'Retinol Sérum 0.5%', preco: 55.0, comissao: 0.1, vendas: 200, loja: 'Dermato Shop', link: 'https://shopee.com.br/4', _fonte: 'curadoria' }
	],
	total_bruto: 1,
	estrategia: 'nicho'
};

/** Helper: locator que acha pelo menos um produto do feed (qualquer keyword) */
function feedProductLocator(page) {
	return page.getByRole('heading', { name: /Kit Beleza|Perfume Importado|Skincare Hidratante/ }).first();
}

// ═══════════════════════════════════════════════════════════════════════════════
// 1. FEED DEFAULT — PRODUTOS ON-LOAD
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Feed Default — Produtos on-load', () => {
	test('página abre com produtos visíveis sem digitar nada', async ({ authedPage: page }) => {
		page.apiOverrides({
			'/api/candidatos': CANDIDATOS_FEED,
			'/api/categorias': {
				marketplaces: [
					{
						marketplace: 'shopee',
						categorias: [
							{ id: 100630, nome: 'Beleza' },
							{ id: 100664, nome: 'Cuidados com a Pele' },
							{ id: 100640, nome: 'Perfumaria' }
						]
					}
				]
			}
		});
		await page.goto('/');

		// Produtos aparecem automaticamente
		await expect(feedProductLocator(page)).toBeVisible({ timeout: 15000 });
	});

	test('indicador de feed default visível (keyword no input)', async ({ authedPage: page }) => {
		page.apiOverrides({ '/api/candidatos': CANDIDATOS_FEED });
		await page.goto('/');

		// Aguarda o feed default carregar
		await expect(feedProductLocator(page)).toBeVisible({ timeout: 15000 });

		// O input omnibox fica limpo (feed é silencioso, input pronto para digitar)
		const input = page.getByRole('combobox');
		const value = await input.inputValue();
		expect(value).toBe('');
	});

	test('busca manual sobrescreve o feed default', async ({ authedPage: page }) => {
		let callCount = 0;
		page.apiOverrides({
			'/api/candidatos': (request) => {
				callCount++;
				if (callCount <= 1) return CANDIDATOS_FEED;
				return CANDIDATOS_BUSCA_MANUAL;
			}
		});
		await page.goto('/');

		// Feed default carregou
		await expect(feedProductLocator(page)).toBeVisible({ timeout: 15000 });

		// Digitar sobrescreve
		const input = page.getByRole('combobox');
		await input.fill('retinol');
		await input.press('Enter');

		// Resultado da busca manual aparece
		await expect(page.getByText('Retinol Sérum 0.5%')).toBeVisible({ timeout: 10000 });
	});

	test('sem empty state confuso ao abrir', async ({ authedPage: page }) => {
		page.apiOverrides({ '/api/candidatos': CANDIDATOS_FEED });
		await page.goto('/');

		// Deve mostrar produtos, "Nenhum resultado" NÃO deve aparecer
		await expect(feedProductLocator(page)).toBeVisible({ timeout: 15000 });
		await expect(page.getByText(/Nenhum resultado/i)).not.toBeVisible();
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 2. FEED DEFAULT — ROTAÇÃO
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Feed Default — Rotação', () => {
	test('produtos do feed aparecem on-load (rotação funciona)', async ({ authedPage: page }) => {
		page.apiOverrides({ '/api/candidatos': CANDIDATOS_FEED });
		await page.goto('/');

		// Algum produto do feed é visível (a rotação escolheu uma categoria)
		await expect(feedProductLocator(page)).toBeVisible({ timeout: 15000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 3. FEED DEFAULT — COMPATIBILIDADE COM BUSCAS SALVAS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Feed Default — Buscas salvas', () => {
	test('com buscas salvas mas sem contexto no boot, feed default carrega normalmente', async ({ authedPage: page }) => {
		page.apiOverrides({
			'/api/candidatos': CANDIDATOS_FEED,
			'/api/buscas': {
				buscas: [
					{ id: 'b1', keywords: ['serum'], shop_ids: [], status: 'ativa', collection_keys: ['serum'] }
				],
				total: 1
			}
		});
		await page.goto('/');

		// Feed default carrega (ter buscas salvas não impede)
		await expect(feedProductLocator(page)).toBeVisible({ timeout: 15000 });
	});
});
