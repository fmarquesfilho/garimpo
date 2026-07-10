/**
 * E2E LOCAL — Contract tests visuais (T-0056 regression).
 *
 * Valida que shop_names, BuscaCard compacto e busca shop-only funcionam.
 * Usa dados reduzidos (1-2 buscas) para evitar timeout da engine.
 */
import { test, expect, FIXTURE_API_BUSCAS, FIXTURE_COLLECTOR_FETCHSHOP } from './fixtures.js';
import { abrirPainelBuscas, SEL } from './helpers.js';

// Buscas individuais dos fixtures (evita carregar todas de uma vez)
const BUSCA_GLORY = FIXTURE_API_BUSCAS.buscas.find((b) => b.id === 'busca-loja-glory');
const BUSCA_MULTI = FIXTURE_API_BUSCAS.buscas.find((b) => b.id === 'busca-loja-multi');
const BUSCA_KEYWORD = FIXTURE_API_BUSCAS.buscas.find((b) => b.id === 'busca-keyword-serum');

// ═══════════════════════════════════════════════════════════════════════════════
// 1. SHOP_NAMES
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Contract — shop_names', () => {
	test('BuscaCard mostra nome da loja, não ID numérico', async ({ authedPage: page }) => {
		page.apiOverrides({ '/api/buscas': { buscas: [BUSCA_GLORY], total: 1 } });
		await page.goto('/');
		await abrirPainelBuscas(page);

		await expect(page.getByText('🏪 Glory of Seoul')).toBeVisible({ timeout: 5000 });
		await expect(page.getByText('🏪 920292999')).toHaveCount(0);
	});

	test('busca multi-loja mostra todos os nomes', async ({ authedPage: page }) => {
		page.apiOverrides({ '/api/buscas': { buscas: [BUSCA_MULTI], total: 1 } });
		await page.goto('/');
		await abrirPainelBuscas(page);

		await expect(page.getByText('🏪 Le Botanic')).toBeVisible({ timeout: 5000 });
		await expect(page.getByText('🏪 COSRX Official')).toBeVisible();
		await expect(page.getByText('🏪 282170857')).toHaveCount(0);
	});

	test('busca keyword-only não mostra badge de loja', async ({ authedPage: page }) => {
		page.apiOverrides({ '/api/buscas': { buscas: [BUSCA_KEYWORD], total: 1 } });
		await page.goto('/');
		await abrirPainelBuscas(page);

		await expect(page.getByText('serum').first()).toBeVisible({ timeout: 5000 });
		// Sem 🏪 no card
		const panel = page.locator('.flex.flex-wrap.gap-2\\.5');
		await expect(panel.getByText('🏪')).toHaveCount(0);
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 2. BUSCA CARD COMPACTO
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Contract — BuscaCard compacto', () => {
	test('BuscaCard sem título bold de display', async ({ authedPage: page }) => {
		page.apiOverrides({ '/api/buscas': { buscas: [BUSCA_GLORY], total: 1 } });
		await page.goto('/');
		await abrirPainelBuscas(page);

		const oldTitle = page.locator('[class*="font-bold"][class*="text-base"][class*="display"]');
		await expect(oldTitle).toHaveCount(0);
	});

	test('BuscaCard mostra cron badge inline', async ({ authedPage: page }) => {
		page.apiOverrides({ '/api/buscas': { buscas: [BUSCA_GLORY], total: 1 } });
		await page.goto('/');
		await abrirPainelBuscas(page);

		await expect(page.getByText('a cada 8h')).toBeVisible({ timeout: 5000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 3. BUSCA SHOP-ONLY
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Contract — Busca shop-only', () => {
	// TODO: este teste precisa de investigação do flow completo engine → resultados → ProductCard
	// A engine executa a busca corretamente mas o resultado não renderiza como ProductCard.
	// Possível causa: o resultado vai para resultado.lojas (não curadoria) e o componente
	// espera campos específicos (id vs produto_id).
	test.fixme('rodar busca com loja mostra produtos', async ({ authedPage: page }) => {
		page.apiOverrides({
			'/api/buscas': { buscas: [BUSCA_GLORY], total: 1 },
			'/api/candidatos': {
				candidatos: FIXTURE_COLLECTOR_FETCHSHOP.products.map((p) => ({
					...p,
					id: p.produto_id,
					_fonte: 'curadoria'
				})),
				total_bruto: FIXTURE_COLLECTOR_FETCHSHOP.total_found,
				estrategia: 'nicho'
			}
		});
		await page.goto('/');
		await abrirPainelBuscas(page);

		await page.locator(SEL.btnRodar).first().click();

		// Após rodar, a engine executa nova busca — aguarda resultado
		await expect(page.getByText('Sérum Facial Vitamina C 30ml')).toBeVisible({ timeout: 25000 });
	});
});
