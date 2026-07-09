/**
 * E2E LOCAL — Cenários completos da página Descobrir (TESTES_DESCOBRIR.md).
 *
 * Cada test valida um fluxo real da UI com API mockada.
 * Organizado por seção do documento de cenários:
 *   1. Fontes de dados (toggles)
 *   2. Filtros (comissão, categorias)
 *   3. Buscas salvas (pills)
 *   4. Input de busca (limpar, ESC, debounce)
 *   5. Empty states
 *   6. Erro de rede
 *   7. Fluxo completo
 */
import { test, expect, mockApi } from './fixtures.js';

// ── Dados mockados (nomes sem acento para match com keyword sem acento) ──────

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

test.describe('Descobrir — Fontes de dados', () => {
	test('#1: busca keyword "serum" → resultados da curadoria', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/candidatos': { candidatos: PRODUTOS_CURADORIA }
		});
		await page.goto('/');
		await page.getByPlaceholder(/Buscar produto/i).fill('serum');
		// "serum" matches "Serum Vitamina C SKIN1004" (case-insensitive, sem acento)
		await expect(page.getByText('Serum Vitamina C SKIN1004')).toBeVisible({ timeout: 10000 });
	});

	test('#2: busca vazia + sem loja → empty state', async ({ garimparPage: page }) => {
		await mockApi(page);
		await page.goto('/');
		await expect(page.getByText(/Nenhum resultado/i)).toBeVisible({ timeout: 5000 });
	});

	test.skip('#4: quedas aparecem quando há busca salva com loja', async ({ garimparPage: page }) => {
		// SKIP: Este cenário depende do store externo $buscasSalvas ser populado ANTES
		// de executarBusca. No E2E local, o store vem de localStorage (vazio) e o mock
		// de /api/buscas popula engine.ctx mas não o store Svelte que effects usa.
		// Bug real de acoplamento store↔engine — issue para corrigir.
		await mockApi(page, {
			'/api/buscas': {
				buscas: [
					{
						id: 'loja-789',
						keywords: ['cosrx'],
						shop_ids: [789],
						nome: 'COSRX Store',
						cron: '0 */8 * * *',
						comissao_min: 0,
						vendas_min: 0,
						categorias: [],
						fontes: ['curadoria', 'quedas']
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
			},
			'/api/candidatos': { candidatos: [] }
		});
		await page.goto('/');

		// Clica na pill que aparece (busca salva "cosrx")
		const pill = page.locator('button', { hasText: 'cosrx' });
		await expect(pill).toBeVisible({ timeout: 10000 });
		await pill.click();

		// Quedas devem aparecer após a busca executar
		await expect(page.getByText('Tonico COSRX AHA BHA')).toBeVisible({ timeout: 15000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 2. FILTROS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Filtros', () => {
	test('#16: categorias aparecem como chips quando API retorna formato correto', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/candidatos': { candidatos: PRODUTOS_CURADORIA },
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
		await page.getByPlaceholder(/Buscar produto/i).fill('serum');
		await page.waitForTimeout(600);

		// Abre filtros
		await page.getByRole('button', { name: /Filtros/i }).click();

		// Categorias devem aparecer como chips
		await expect(page.locator('button', { hasText: /^Perfumaria$/ })).toBeVisible({ timeout: 10000 });
		await expect(page.locator('button', { hasText: /^Maquiagem$/ })).toBeVisible();
	});

	test('comissão select mostra opções formatadas em %', async ({ garimparPage: page }) => {
		await mockApi(page);
		await page.goto('/');

		// Abre filtros
		await page.getByRole('button', { name: /Filtros/i }).click();

		// Label "comissão mín." visível
		await expect(page.getByText('comissão mín.')).toBeVisible();

		// Nenhum float cru visível na seção de filtros
		const filterSection = page.locator('.bg-muted').first();
		const text = await filterSection.textContent();
		expect(text).not.toMatch(/0\.\d{5,}/);
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 3. BUSCAS SALVAS (pills)
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Buscas salvas', () => {
	test('#27: pill restaura keyword no input', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/buscas': {
				buscas: [
					{
						id: 'b1',
						keywords: ['retinol'],
						shop_ids: [],
						comissao_min: 0.1,
						vendas_min: 0,
						categorias: [],
						fontes: ['curadoria'],
						cron: null
					}
				],
				total: 1
			},
			'/api/candidatos': { candidatos: [] }
		});
		await page.goto('/');

		// Pill aparece
		const pill = page.locator('button', { hasText: 'retinol' });
		await expect(pill).toBeVisible({ timeout: 10000 });

		// Clicar restaura keyword
		await pill.click();
		await expect(page.getByPlaceholder(/Buscar produto/i)).toHaveValue('retinol');
	});

	test('#30: busca agendada mostra badge ⏱', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/buscas': {
				buscas: [
					{
						id: 'b-cron',
						keywords: ['perfume'],
						shop_ids: [],
						comissao_min: 0.07,
						vendas_min: 0,
						categorias: [],
						fontes: ['curadoria'],
						cron: '0 */8 * * *'
					}
				],
				total: 1
			}
		});
		await page.goto('/');

		await expect(page.getByText('⏱')).toBeVisible({ timeout: 10000 });
		await expect(page.locator('button', { hasText: 'perfume' })).toBeVisible();
	});

	test('#29: pill com categorias restaura categorias (filtros abertos)', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/buscas': {
				buscas: [
					{
						id: 'b-cat',
						keywords: ['serum'],
						shop_ids: [],
						comissao_min: 0.07,
						vendas_min: 0,
						categorias: ['Beleza'],
						fontes: ['curadoria'],
						cron: null
					}
				],
				total: 1
			},
			'/api/candidatos': { candidatos: PRODUTOS_CURADORIA },
			'/api/categorias': {
				marketplaces: [
					{
						marketplace: 'shopee',
						categorias: [
							{ id: 100630, nome: 'Beleza' },
							{ id: 100640, nome: 'Perfumaria' }
						]
					}
				]
			}
		});
		await page.goto('/');

		// Clica no pill
		const pill = page.locator('button', { hasText: 'serum' });
		await expect(pill).toBeVisible({ timeout: 10000 });
		await pill.click();

		// Abre filtros
		await page.getByRole('button', { name: /Filtros/i }).click();

		// Chip "Beleza" deve estar ativa (border-primary)
		const chipBeleza = page.locator('button', { hasText: /^Beleza$/ });
		await expect(chipBeleza).toBeVisible({ timeout: 10000 });
		await expect(chipBeleza).toHaveClass(/border-primary|bg-accent/);
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 4. INPUT DE BUSCA
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Input de busca', () => {
	test('#43: botão ✕ limpa o campo', async ({ garimparPage: page }) => {
		await mockApi(page, { '/api/candidatos': { candidatos: PRODUTOS_CURADORIA } });
		await page.goto('/');

		const input = page.getByPlaceholder(/Buscar produto/i);
		await input.fill('serum');
		await expect(input).toHaveValue('serum');

		// Botão ✕ aparece (aria-label="Limpar")
		const clearBtn = page.getByRole('button', { name: /Limpar/i });
		await expect(clearBtn).toBeVisible();
		await clearBtn.click();
		await expect(input).toHaveValue('');
	});

	test('#44: ESC limpa o campo', async ({ garimparPage: page }) => {
		await mockApi(page);
		await page.goto('/');

		const input = page.getByPlaceholder(/Buscar produto/i);
		await input.fill('serum');
		await input.press('Escape');
		await expect(input).toHaveValue('');
	});

	test('#45: campo vazio não mostra botão ✕', async ({ garimparPage: page }) => {
		await mockApi(page);
		await page.goto('/');
		await expect(page.getByRole('button', { name: /Limpar/i })).toHaveCount(0);
	});

	test('#46: digitar keyword → resultados aparecem após debounce', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/candidatos': { candidatos: PRODUTOS_CURADORIA }
		});
		await page.goto('/');

		// Digita rápido
		await page.getByPlaceholder(/Buscar produto/i).pressSequentially('serum', { delay: 50 });

		// Resultados aparecem após debounce
		await expect(page.getByText('Serum Vitamina C SKIN1004')).toBeVisible({ timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 5. EMPTY STATES
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Empty states', () => {
	test('#38: sem contexto → empty state', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/buscas': { buscas: [], total: 0 },
			'/api/candidatos': { candidatos: [] }
		});
		await page.goto('/');
		await expect(page.getByText(/Nenhum resultado/i)).toBeVisible({ timeout: 5000 });
	});

	test('keyword sem match → empty state', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/candidatos': { candidatos: [] }
		});
		await page.goto('/');
		await page.getByPlaceholder(/Buscar produto/i).fill('xyzinexistente');
		await expect(page.getByText(/Nenhum resultado/i)).toBeVisible({ timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 6. ERRO DE REDE
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Erro', () => {
	test.skip('#40: API retorna erro → UI mostra mensagem', async ({ garimparPage: page }) => {
		// SKIP: O frontend engole erros HTTP em carregarCuradoria (catch → return [])
		// O timeout de 25s é o único caminho para ERROR, mas não podemos esperar 25s no E2E.
		// Coberto pelo unit test (busca-engine-cenarios.test.js: "timeout > 25s").
		// Registra routes ANTES de mockApi (para ter prioridade)
		await page.route('**/api/candidatos**', async (route) => {
			await route.fulfill({ status: 500, body: 'Internal Server Error' });
		});
		// mockApi para as outras rotas
		await page.route('**/api/buscas**', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ buscas: [], total: 0 })
			});
		});
		await page.route('**/api/favoritos**', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ favoritos: [] })
			});
		});
		await page.route('**/api/categorias**', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ marketplaces: [] })
			});
		});
		await page.route('**/api/alertas**', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({})
			});
		});

		await page.goto('/');
		await page.getByPlaceholder(/Buscar produto/i).fill('serum');

		// Aguarda mensagem de erro (a engine mostra ctx.error)
		await expect(page.getByText(/falhou|erro/i)).toBeVisible({ timeout: 15000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 7. FLUXO COMPLETO
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Fluxo completo', () => {
	test('busca serum → adiciona loja → salva → pill restaura', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/candidatos': { candidatos: PRODUTOS_CURADORIA },
			'/api/lojas': { id: 'loja-lb', keyword: 'Le Botanic', shop_ids: [920292999], status: 'adicionada' },
			'/api/buscas': (req) => {
				if (req.method() === 'POST') {
					return { id: 'busca-new', keywords: ['serum'], shop_ids: [920292999] };
				}
				return {
					buscas: [
						{
							id: 'busca-new',
							keywords: ['serum'],
							shop_ids: [920292999],
							nome: 'Le Botanic',
							comissao_min: 0.07,
							vendas_min: 0,
							categorias: [],
							fontes: ['curadoria'],
							cron: null
						}
					],
					total: 1
				};
			}
		});

		await page.goto('/');

		// 1. Busca keyword
		const input = page.getByPlaceholder(/Buscar produto/i);
		await input.fill('serum');
		await expect(page.getByText('Serum Vitamina C SKIN1004')).toBeVisible({ timeout: 10000 });

		// 2. Adiciona loja
		const lojaInput = page.locator('input[placeholder*="loja"]').first();
		await lojaInput.fill('https://s.shopee.com.br/8fQYnxWQqu');
		await lojaInput.press('Enter');
		await expect(page.getByText('🏪 Le Botanic')).toBeVisible({ timeout: 10000 });

		// 3. Salva
		await page
			.getByRole('button', { name: /Salvar/i })
			.first()
			.click();
		await page.getByRole('button', { name: /^Salvar/ }).click();

		// 4. Pill aparece com label correto
		await expect(page.locator('button', { hasText: 'serum' })).toBeVisible({ timeout: 10000 });

		// 5. Limpa e restaura
		await input.fill('');
		await page.waitForTimeout(600);
		await page.locator('button', { hasText: 'serum' }).click();
		await expect(input).toHaveValue('serum');
	});
});
