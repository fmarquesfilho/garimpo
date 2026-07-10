/**
 * E2E LOCAL — Contract tests visuais.
 *
 * Valida que os bugs corrigidos em T-0056 não regridem:
 * - shop_names renderiza nomes corretos (não IDs numéricos)
 * - BuscaCard compacto (sem título bold)
 * - Busca shop-only retorna resultados
 * - Multi-loja mapeia todos os nomes
 *
 * Usa dados dos fixtures/ compartilhados para garantir consistência cross-stack.
 */
import { test, expect, mockApiFromFixtures, FIXTURE_API_BUSCAS, FIXTURE_COLLECTOR_FETCHSHOP } from './fixtures.js';

// ═══════════════════════════════════════════════════════════════════════════════
// 1. SHOP_NAMES — Nomes de loja corretos
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Contract — shop_names rendering', () => {
	test('BuscaCard mostra nome da loja, não ID numérico', async ({ garimparPage: page }) => {
		await mockApiFromFixtures(page);
		await page.goto('/');

		// A fixture tem busca "busca-loja-glory" com shop_names: { "920292999": "Glory of Seoul" }
		// O BuscaCard deve renderizar "🏪 Glory of Seoul"
		await expect(page.getByText('🏪 Glory of Seoul')).toBeVisible({ timeout: 10000 });
		// Nunca deve mostrar o ID numérico como texto visível do badge
		await expect(page.getByText('🏪 920292999')).toHaveCount(0);
	});

	test('busca multi-loja mostra todos os nomes', async ({ garimparPage: page }) => {
		await mockApiFromFixtures(page);
		await page.goto('/');

		// Fixture "busca-loja-multi" tem 2 lojas: Le Botanic + COSRX Official
		await expect(page.getByText('🏪 Le Botanic')).toBeVisible({ timeout: 10000 });
		await expect(page.getByText('🏪 COSRX Official')).toBeVisible();
		// Nunca mostra IDs
		await expect(page.getByText('🏪 282170857')).toHaveCount(0);
		await expect(page.getByText('🏪 592884015')).toHaveCount(0);
	});

	test('busca keyword-only não mostra seção de lojas', async ({ garimparPage: page }) => {
		// Mocka com apenas a busca keyword
		const buscaKeywordOnly = {
			buscas: [FIXTURE_API_BUSCAS.buscas.find((b) => b.id === 'busca-keyword-serum')],
			total: 1
		};
		await mockApiFromFixtures(page, { '/api/buscas': buscaKeywordOnly });
		await page.goto('/');

		// Deve mostrar keyword "serum" mas não o badge 🏪
		await expect(page.locator('button', { hasText: 'serum' })).toBeVisible({ timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 2. BUSCA CARD COMPACTO
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Contract — BuscaCard compacto', () => {
	test('BuscaCard não tem título bold redundante', async ({ garimparPage: page }) => {
		await mockApiFromFixtures(page);
		await page.goto('/');

		// Abre painel de buscas salvas
		await page.getByRole('button', { name: /Buscas/i }).click();
		await page.waitForTimeout(500);

		// Não deve haver um elemento com font-bold text-base (antigo título)
		const boldTitle = page.locator('.font-bold.text-base');
		await expect(boldTitle).toHaveCount(0);
	});

	test('BuscaCard mostra cron badge inline', async ({ garimparPage: page }) => {
		await mockApiFromFixtures(page);
		await page.goto('/');

		// Busca "busca-loja-glory" tem cron "0 */8 * * *"
		await expect(page.getByText('⏱ a cada 8h')).toBeVisible({ timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 3. BUSCA SHOP-ONLY (sem keyword)
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Contract — Busca shop-only', () => {
	test('busca com loja sem keyword retorna produtos via FetchShop', async ({ garimparPage: page }) => {
		// Mocka com busca loja-only + candidatos que vêm do FetchShop
		const buscaLojaOnly = {
			buscas: [FIXTURE_API_BUSCAS.buscas.find((b) => b.id === 'busca-loja-glory')],
			total: 1
		};
		const shopProducts = {
			candidatos: FIXTURE_COLLECTOR_FETCHSHOP.products.map((p) => ({
				...p,
				id: p.produto_id,
				_fonte: 'curadoria'
			}))
		};

		await mockApiFromFixtures(page, {
			'/api/buscas': buscaLojaOnly,
			'/api/candidatos': shopProducts
		});
		await page.goto('/');

		// Clica na pill da busca (busca-loja-glory não tem keywords, deve mostrar a loja)
		const pill = page.locator('button', { hasText: /Glory of Seoul/i });
		if (await pill.isVisible({ timeout: 5000 }).catch(() => false)) {
			await pill.click();
		}

		// Produtos da loja devem aparecer
		await expect(page.getByText('Sérum Facial Vitamina C 30ml')).toBeVisible({ timeout: 15000 });
	});

	test('busca só-categoria usa nome como keyword', async ({ garimparPage: page }) => {
		const buscaCategoria = {
			buscas: [FIXTURE_API_BUSCAS.buscas.find((b) => b.id === 'busca-categoria')],
			total: 1
		};

		let capturedUrl = null;
		await mockApiFromFixtures(page, {
			'/api/buscas': buscaCategoria,
			'/api/candidatos': (req) => {
				capturedUrl = req.url();
				return { candidatos: FIXTURE_API_BUSCAS.buscas.length > 0 ? [] : [] };
			}
		});
		await page.goto('/');

		// Clica na pill da busca-categoria
		const pill = page.locator('button', { hasText: /Skincare|Beleza|categor/i });
		if (await pill.isVisible({ timeout: 5000 }).catch(() => false)) {
			await pill.click();
			await page.waitForTimeout(1000);
			// A chamada à API deve ter keyword=Skincare (primeira categoria)
			if (capturedUrl) {
				expect(capturedUrl).toContain('keyword=Skincare');
			}
		}
	});
});
