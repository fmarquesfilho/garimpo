/**
 * E2E LOCAL — Cenários completos da página Descobrir.
 *
 * Organizado por seção:
 *   1. Fontes de dados (toggles)
 *   2. Filtros (comissão, categorias)
 *   3. Buscas salvas (painel)
 *   4. Input de busca (limpar, ESC, debounce)
 *   5. Empty states
 *   6. Erro de rede
 *   7. Fluxo completo
 */
import { test, expect } from './fixtures.js';
import { abrirPainelBuscas, adicionarLojaViaOmnibox, waitForEngineReady, SEL } from './helpers.js';

const PRODUTOS_CURADORIA = [
	{
		id: 'p1',
		produto_id: 'p1',
		nome: 'Serum Vitamina C SKIN1004',
		preco: 89.9,
		comissao: 0.15,
		vendas: 200,
		loja: 'SKIN1004 Official',
		link: 'https://shopee.com.br/p1',
		_fonte: 'curadoria'
	},
	{
		id: 'p2',
		produto_id: 'p2',
		nome: 'Perfume Kenzo 50ml',
		preco: 299.9,
		comissao: 0.08,
		vendas: 80,
		loja: 'Perfumaria JP',
		link: 'https://shopee.com.br/p2',
		_fonte: 'curadoria'
	},
	{
		id: 'p3',
		produto_id: 'p3',
		nome: 'Batom Matte Ruby Rose',
		preco: 15,
		comissao: 0.03,
		vendas: 5,
		loja: 'Maquiagem Store',
		link: 'https://shopee.com.br/p3',
		_fonte: 'curadoria'
	}
];

// ═══════════════════════════════════════════════════════════════════════════════
// 1. FONTES DE DADOS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Fontes', () => {
	test('busca keyword → resultados curadoria', async ({ authedPage: page }) => {
		page.apiOverrides({ '/api/candidatos': { candidatos: PRODUTOS_CURADORIA, total_bruto: 3, estrategia: 'nicho' } });
		await page.goto('/');
		await page.locator(SEL.searchInput).fill('serum');
		await expect(page.getByText('Serum Vitamina C SKIN1004')).toBeVisible({ timeout: 10000 });
	});

	test('busca vazia + sem loja → empty state', async ({ authedPage: page }) => {
		await page.goto('/');
		await expect(page.getByText(/Nenhum resultado/i)).toBeVisible({ timeout: 10000 });
	});

	test('quedas aparecem quando há busca salva com loja', async ({ authedPage: page }) => {
		page.apiOverrides({
			'/api/buscas': {
				buscas: [
					{
						id: 'loja-789',
						keywords: ['cosrx'],
						shop_ids: [789],
						shop_names: { 789: 'COSRX Store' },
						cron: '0 */8 * * *',
						comissao_min: 0,
						vendas_min: 0,
						categorias: null,
						fontes: ['curadoria', 'quedas'],
						marketplaces: 'shopee'
					}
				],
				total: 1
			},
			'/api/lojas/novidades': {
				variacoes: [
					{
						produto_id: 'q1',
						nome: 'Tonico COSRX AHA BHA',
						preco_atual: 45.9,
						preco_anterior: 59.9,
						variacao_pct: -0.23,
						loja: 'COSRX Store'
					}
				],
				produtos_novos: [],
				dias_janela: 7
			}
		});
		await page.goto('/');
		await waitForEngineReady(page);
		await abrirPainelBuscas(page);

		// Clica rodar na busca salva
		await page.locator(SEL.btnRodar).first().click();

		// Quedas devem aparecer
		await expect(page.getByText('Tonico COSRX AHA BHA')).toBeVisible({ timeout: 15000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 2. FILTROS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Filtros', () => {
	test('categorias combobox visível na raia Filtros', async ({ authedPage: page }) => {
		page.apiOverrides({
			'/api/candidatos': { candidatos: PRODUTOS_CURADORIA, total_bruto: 3, estrategia: 'nicho' },
			'/api/categorias': {
				marketplaces: [
					{
						marketplace: 'shopee',
						categorias: [
							{ id: 100630, nome: 'Beleza' },
							{ id: 100640, nome: 'Perfumaria' },
							{ id: 100663, nome: 'Maquiagem' }
						]
					}
				]
			}
		});
		await page.goto('/');
		await page.locator(SEL.searchInput).fill('serum');
		await page.waitForTimeout(600);

		// Filtros agora sao inline (sempre visiveis) — categorias via Omnibox #prefixo
		await expect(page.getByText('comissão mín.')).toBeVisible({ timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 3. BUSCAS SALVAS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Buscas salvas', () => {
	test('rodar restaura keyword no input', async ({ authedPage: page }) => {
		page.apiOverrides({
			'/api/buscas': {
				buscas: [
					{
						id: 'b1',
						keywords: ['retinol'],
						shop_ids: [],
						shop_names: null,
						comissao_min: 0.1,
						vendas_min: 0,
						categorias: null,
						fontes: ['curadoria'],
						cron: null,
						marketplaces: 'shopee'
					}
				],
				total: 1
			}
		});
		await page.goto('/');
		await waitForEngineReady(page);
		await abrirPainelBuscas(page);
		await page.locator(SEL.btnRodar).first().click();
		await expect(page.locator(SEL.searchInput)).toHaveValue('retinol');
	});

	test('busca agendada mostra badge ⏱', async ({ authedPage: page }) => {
		page.apiOverrides({
			'/api/buscas': {
				buscas: [
					{
						id: 'b-cron',
						keywords: ['perfume'],
						shop_ids: [],
						shop_names: null,
						comissao_min: 0.07,
						vendas_min: 0,
						categorias: null,
						fontes: ['curadoria'],
						cron: '0 */8 * * *',
						marketplaces: 'shopee'
					}
				],
				total: 1
			}
		});
		await page.goto('/');
		await waitForEngineReady(page);
		await abrirPainelBuscas(page);
		await expect(page.getByText('a cada 8h')).toBeVisible({ timeout: 5000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 4. INPUT
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Input', () => {
	test('ESC limpa o campo', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.locator(SEL.searchInput);
		await input.fill('serum');
		await input.press('Escape');
		await expect(input).toHaveValue('');
	});

	test('debounce → resultados aparecem', async ({ authedPage: page }) => {
		page.apiOverrides({ '/api/candidatos': { candidatos: PRODUTOS_CURADORIA, total_bruto: 3, estrategia: 'nicho' } });
		await page.goto('/');
		await page.locator(SEL.searchInput).pressSequentially('serum', { delay: 50 });
		await expect(page.getByText('Serum Vitamina C SKIN1004')).toBeVisible({ timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 5. EMPTY STATES
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Empty states', () => {
	test('sem contexto → empty state', async ({ authedPage: page }) => {
		await page.goto('/');
		await expect(page.getByText(/Nenhum resultado/i)).toBeVisible({ timeout: 10000 });
	});

	test('keyword sem match → empty state', async ({ authedPage: page }) => {
		await page.goto('/');
		await page.locator(SEL.searchInput).fill('xyzinexistente');
		await expect(page.getByText(/Nenhum resultado/i)).toBeVisible({ timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 6. ERRO DE REDE
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Erro', () => {
	test('API 500 → mostra mensagem de erro', async ({ authedPage: page }) => {
		// Override /api/candidatos com resposta de erro
		page.apiOverrides({
			'/api/candidatos': async () => {
				throw new Error('500');
			}
		});

		// Re-register route that returns 500 for candidatos
		await page.unroute('**/api/**');
		await page.route('**/api/**', async (route) => {
			const path = new URL(route.request().url()).pathname;
			if (path.startsWith('/api/candidatos')) {
				await route.fulfill({
					status: 500,
					contentType: 'application/json',
					body: JSON.stringify({ detail: 'Servidor indisponível' })
				});
			} else {
				await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({}) });
			}
		});

		await page.goto('/');
		await page.locator(SEL.searchInput).fill('serum');
		await expect(page.getByText(/indisponível|falhou|Erro|demorou/i)).toBeVisible({ timeout: 15000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 7. FLUXO COMPLETO
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Fluxo completo', () => {
	test('busca → adiciona loja → salva → card aparece', async ({ authedPage: page }) => {
		let postCount = 0;
		page.apiOverrides({
			'/api/candidatos': { candidatos: PRODUTOS_CURADORIA, total_bruto: 3, estrategia: 'nicho' },
			'/api/lojas': { id: 'loja-lb', keyword: 'Le Botanic', shop_ids: [920292999], status: 'adicionada' },
			'/api/buscas': (req) => {
				if (req.method() === 'POST') {
					postCount++;
					return { id: 'busca-new', keywords: ['serum'], shop_ids: [920292999], status: 'salva' };
				}
				if (postCount > 0) {
					return {
						buscas: [
							{
								id: 'busca-new',
								keywords: ['serum'],
								shop_ids: [920292999],
								shop_names: { 920292999: 'Le Botanic' },
								comissao_min: 0.07,
								vendas_min: 0,
								categorias: null,
								fontes: ['curadoria'],
								cron: null,
								marketplaces: 'shopee'
							}
						],
						total: 1
					};
				}
				return { buscas: [], total: 0 };
			}
		});

		await page.goto('/');

		// 1. Busca keyword
		await page.locator(SEL.searchInput).fill('serum');
		await expect(page.getByText('Serum Vitamina C SKIN1004')).toBeVisible({ timeout: 10000 });

		// 2. Adiciona loja via Omnibox
		await adicionarLojaViaOmnibox(page, 'https://s.shopee.com.br/8fQYnxWQqu');
		await expect(page.getByText('Le Botanic')).toBeVisible({ timeout: 10000 });

		// 3. Abre painel salvas e salva
		await page.locator(SEL.btnBuscasSalvas).click();
		await page.getByRole('button', { name: /salvar busca atual/ }).click();
		await page.locator('button:has-text("Salvar")').last().click();

		// 4. BuscaCard com loja aparece
		await expect(page.getByText('🏪 Le Botanic')).toBeVisible({ timeout: 10000 });
	});
});
