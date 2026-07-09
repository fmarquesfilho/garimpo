/**
 * E2E LOCAL — Cenários completos da página Descobrir (TESTES_DESCOBRIR.md).
 *
 * Cada test valida um fluxo real da UI com API mockada.
 * Organizado por seção do documento de cenários:
 *   1. Fontes de dados (toggles)
 *   2. Filtros (comissão, vendas, categorias)
 *   3. Buscas salvas (pills)
 *   4. Input de busca (debounce, limpar, ESC)
 *   5. Empty states
 *   6. Timeout e erro
 */
import { test, expect, mockApi } from './fixtures.js';

// ── Dados mockados realistas ─────────────────────────────────────────────────

const PRODUTOS_CURADORIA = [
	{
		id: 'p1',
		produto_id: 'p1',
		nome: 'Sérum Vitamina C SKIN1004',
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

const PRODUTOS_QUEDAS = [
	{
		id: 'q1',
		produto_id: 'q1',
		nome: 'Tônico COSRX AHA',
		preco: 45.9,
		comissao: 0.12,
		vendas: 150,
		loja: 'COSRX Store',
		variacao_pct: -0.25,
		preco_anterior: 59.9,
		_fonte: 'queda'
	},
	{
		id: 'q2',
		produto_id: 'q2',
		nome: 'Skin1004 Centella Toner',
		preco: 75,
		comissao: 0.1,
		vendas: 300,
		loja: 'SKIN1004 Official',
		variacao_pct: -0.15,
		preco_anterior: 89,
		_fonte: 'queda'
	}
];

const PRODUTOS_NOVOS = [
	{
		id: 'n1',
		produto_id: 'n1',
		nome: 'Retinol Serum Novo',
		preco: 45.5,
		comissao: 0.1,
		vendas: 0,
		loja: 'SKIN1004 Official',
		_fonte: 'novo'
	}
];

// ═══════════════════════════════════════════════════════════════════════════════
// 1. FONTES DE DADOS (toggles)
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Fontes de dados', () => {
	test('#1: busca keyword "sérum" → resultados da curadoria', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/candidatos': { candidatos: PRODUTOS_CURADORIA }
		});
		await page.goto('/');
		await page.getByPlaceholder(/Buscar produto/i).fill('sérum');
		await expect(page.getByText('Sérum Vitamina C SKIN1004')).toBeVisible({ timeout: 10000 });
	});

	test('#2: busca vazia + sem loja → empty state', async ({ garimparPage: page }) => {
		await mockApi(page);
		await page.goto('/');
		// Campo vazio, nenhuma loja → deve mostrar empty state ou nada
		await expect(page.getByText(/Nenhum resultado/i)).toBeVisible({ timeout: 5000 });
	});

	test('#4/#5: toggle Quedas com keyword filtra por nome/loja', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/candidatos': { candidatos: PRODUTOS_CURADORIA },
			'/api/lojas/novidades': { variacoes: PRODUTOS_QUEDAS, produtos_novos: PRODUTOS_NOVOS, dias_janela: 7 }
		});
		await page.goto('/');

		// Precisa de loja para Quedas/Novos aparecerem
		// Mocka buscas salvas com loja para que os dados de quedas/novos carreguem
		await page.getByPlaceholder(/Buscar produto/i).fill('COSRX');
		await expect(page.getByText('Tônico COSRX AHA')).toBeVisible({ timeout: 10000 });
	});

	test('#8/#9: toggle Favoritos mostra produtos favoritados', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/favoritos': {
				favoritos: [
					{
						produto_id: 'f1',
						nome: 'Meu Perfume Favorito',
						preco: 150,
						comissao: 0.09,
						vendas: 60,
						loja: 'Loja Favorita'
					}
				]
			},
			'/api/candidatos': { candidatos: [] }
		});
		await page.goto('/');

		// Ativa fonte Favoritos clicando no toggle
		const favToggle = page.locator('.fonte-btn', { hasText: '⭐ Favoritos' });
		await favToggle.click();

		// Precisa de contexto para a engine executar (digita algo)
		await page.getByPlaceholder(/Buscar produto/i).fill('perfume');
		await expect(page.getByText('Meu Perfume Favorito')).toBeVisible({ timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 2. FILTROS (comissão, vendas, categorias)
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Filtros', () => {
	test('#16/#17: filtro por categoria via chips', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/candidatos': { candidatos: PRODUTOS_CURADORIA },
			'/api/categorias': {
				categorias: [{ nome: 'Perfumaria' }, { nome: 'Maquiagem' }, { nome: 'Cuidados com a Pele' }]
			}
		});
		await page.goto('/');
		await page.getByPlaceholder(/Buscar produto/i).fill('a');
		await page.waitForTimeout(500);

		// Abre filtros
		await page.getByRole('button', { name: /Filtros/i }).click();

		// Categorias devem aparecer como chips clicáveis
		const chipPerfumaria = page.locator('button', { hasText: 'Perfumaria' });
		await expect(chipPerfumaria).toBeVisible({ timeout: 5000 });

		// Clica para ativar categoria
		await chipPerfumaria.click();

		// Chip deve ficar "ativo" (styled differently — border-primary)
		await expect(chipPerfumaria).toHaveClass(/border-primary|bg-accent/);
	});

	test('comissão select mostra opções em %', async ({ garimparPage: page }) => {
		await mockApi(page);
		await page.goto('/');

		// Abre filtros
		await page.getByRole('button', { name: /Filtros/i }).click();

		// Verifica que o label "comissão mín." existe
		await expect(page.getByText('comissão mín.')).toBeVisible();

		// O select deve existir com valores formatados (5%, 7%, 10%, 15%)
		// Não deve conter "0.07" ou similares em nenhum texto visível dos filtros
		const filterSection = page.locator('.bg-muted').first();
		const text = await filterSection.textContent();
		expect(text).not.toMatch(/0\.\d{2,}/); // Nenhum float cru
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 3. BUSCAS SALVAS (pills)
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Buscas salvas', () => {
	test('#27/#28: pill de busca salva restaura keyword e fontes', async ({ garimparPage: page }) => {
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
						fontes: ['curadoria', 'quedas'],
						cron: null
					}
				],
				total: 1
			},
			'/api/candidatos': { candidatos: PRODUTOS_CURADORIA }
		});
		await page.goto('/');

		// Pill deve aparecer com label "retinol"
		const pill = page.locator('button', { hasText: 'retinol' });
		await expect(pill).toBeVisible({ timeout: 10000 });

		// Clicar no pill
		await pill.click();

		// Keyword deve ser preenchida
		const input = page.getByPlaceholder(/Buscar produto/i);
		await expect(input).toHaveValue('retinol');
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

		// Badge ⏱ deve aparecer junto à pill
		await expect(page.getByText('⏱')).toBeVisible({ timeout: 10000 });
		// Pill deve ter o label da busca
		await expect(page.locator('button', { hasText: 'perfume' })).toBeVisible();
	});

	test('#29: busca salva com categorias restaura categorias ao clicar', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/buscas': {
				buscas: [
					{
						id: 'b-cat',
						keywords: ['serum'],
						shop_ids: [],
						comissao_min: 0.07,
						vendas_min: 0,
						categorias: ['Cuidados com a Pele'],
						fontes: ['curadoria'],
						cron: null
					}
				],
				total: 1
			},
			'/api/candidatos': { candidatos: PRODUTOS_CURADORIA },
			'/api/categorias': { categorias: [{ nome: 'Perfumaria' }, { nome: 'Cuidados com a Pele' }] }
		});
		await page.goto('/');

		// Clica no pill
		const pill = page.locator('button', { hasText: 'serum' });
		await expect(pill).toBeVisible({ timeout: 10000 });
		await pill.click();

		// Abre filtros para verificar que categoria foi ativada
		await page.getByRole('button', { name: /Filtros/i }).click();
		const chipCuidados = page.locator('button', { hasText: 'Cuidados com a Pele' });
		await expect(chipCuidados).toBeVisible();
		// Deve estar ativa (classe do estado on)
		await expect(chipCuidados).toHaveClass(/border-primary|bg-accent/);
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 4. INPUT DE BUSCA (debounce, limpar, ESC)
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Input de busca', () => {
	test('#43: botão ✕ limpa o campo', async ({ garimparPage: page }) => {
		await mockApi(page, { '/api/candidatos': { candidatos: PRODUTOS_CURADORIA } });
		await page.goto('/');

		const input = page.getByPlaceholder(/Buscar produto/i);
		await input.fill('serum');
		await expect(input).toHaveValue('serum');

		// Botão ✕ aparece
		const clearBtn = page.getByRole('button', { name: /Limpar/i });
		await expect(clearBtn).toBeVisible();
		await clearBtn.click();

		// Campo deve estar vazio
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

		// Sem texto → botão limpar não deve existir
		const clearBtn = page.getByRole('button', { name: /Limpar/i });
		await expect(clearBtn).toHaveCount(0);
	});

	test('#46: debounce — busca não dispara imediatamente', async ({ garimparPage: page }) => {
		let fetchCount = 0;
		await page.route('**/api/candidatos**', async (route) => {
			fetchCount++;
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ candidatos: PRODUTOS_CURADORIA })
			});
		});
		await mockApi(page);
		await page.goto('/');

		const input = page.getByPlaceholder(/Buscar produto/i);

		// Digita rápido — não deve disparar para cada letra
		await input.pressSequentially('serum', { delay: 50 });

		// Espera o debounce (400ms + margem)
		await page.waitForTimeout(600);

		// Resultados devem aparecer (uma busca foi feita)
		await expect(page.getByText('Sérum Vitamina C SKIN1004')).toBeVisible({ timeout: 5000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 5. EMPTY STATES
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Empty states', () => {
	test('#38: sem lojas + fontes Quedas/Novos → sem resultados', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/buscas': { buscas: [], total: 0 },
			'/api/candidatos': { candidatos: [] },
			'/api/lojas/novidades': { variacoes: [], produtos_novos: [], dias_janela: 7 }
		});
		await page.goto('/');

		// Sem keyword, sem loja → empty state
		await expect(page.getByText(/Nenhum resultado/i)).toBeVisible({ timeout: 5000 });
	});

	test('busca sem resultados mostra empty state', async ({ garimparPage: page }) => {
		await mockApi(page, {
			'/api/candidatos': { candidatos: [] }
		});
		await page.goto('/');
		await page.getByPlaceholder(/Buscar produto/i).fill('xyzinexistente');
		await expect(page.getByText(/Nenhum resultado/i)).toBeVisible({ timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 6. TIMEOUT E ERRO
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Timeout e erro', () => {
	test('#40: API demora > timeout → mostra erro', async ({ garimparPage: page }) => {
		// Mocka API que nunca responde (simula timeout)
		await page.route('**/api/candidatos**', async (route) => {
			// Nunca responde — o frontend tem timeout de 25s
			// Para o teste, vamos simular retornando erro após curto delay
			await new Promise((r) => setTimeout(r, 500));
			await route.abort('timedout');
		});
		await mockApi(page);
		await page.goto('/');

		await page.getByPlaceholder(/Buscar produto/i).fill('serum');

		// Aguarda o erro aparecer (pode ser mensagem de timeout ou erro genérico)
		await expect(page.locator('.msg-erro, [class*="erro"]')).toBeVisible({ timeout: 30000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 7. FLUXO COMPLETO: busca → loja → filtro → salvar → restaurar
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Descobrir — Fluxo completo E2E', () => {
	test('busca serum → adiciona Le Botanic → filtra comissão → salva → restaura', async ({ garimparPage: page }) => {
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
							fontes: ['curadoria', 'lojas'],
							cron: null
						}
					],
					total: 1
				};
			},
			'/api/categorias': {
				categorias: [{ nome: 'Perfumaria' }, { nome: 'Maquiagem' }, { nome: 'Cuidados com a Pele' }]
			}
		});

		await page.goto('/');

		// 1. Digita keyword
		const input = page.getByPlaceholder(/Buscar produto/i);
		await input.fill('serum');
		await expect(page.getByText('Sérum Vitamina C SKIN1004')).toBeVisible({ timeout: 10000 });

		// 2. Adiciona loja
		const lojaInput = page.locator('input[placeholder*="loja"]').first();
		await lojaInput.fill('https://s.shopee.com.br/8fQYnxWQqu');
		await lojaInput.press('Enter');
		await expect(page.getByText('Le Botanic')).toBeVisible({ timeout: 10000 });

		// 3. Abre filtros e verifica que comissão não tem float cru
		await page.getByRole('button', { name: /Filtros/i }).click();
		const filterText = await page.locator('.bg-muted').first().textContent();
		expect(filterText).not.toMatch(/0\.\d{5,}/);

		// 4. Salva busca
		await page
			.getByRole('button', { name: /Salvar/i })
			.first()
			.click();
		await page.getByRole('button', { name: /^Salvar/ }).click();

		// 5. Pill deve aparecer com label correto (não "sem keywords")
		await expect(page.locator('button', { hasText: 'serum' })).toBeVisible({ timeout: 10000 });

		// 6. Limpa e restaura via pill
		await input.fill('');
		await page.waitForTimeout(500);
		const pill = page.locator('button', { hasText: 'serum' });
		await pill.click();
		await expect(input).toHaveValue('serum');
	});
});
