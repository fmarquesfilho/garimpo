/**
 * E2E LOCAL — Smart Search (Omnibox refatorado).
 *
 * Testa os fluxos do Omnibox como Headless UI Controller:
 * - Detecção de intenção (texto livre → Smart Dropdown)
 * - Busca de lojas por nome (Store Cards)
 * - Resolução de link (URL colada)
 * - Monitoramento inline (Store Card → botão Monitorar)
 * - Coexistência com prefixos (@, #, !)
 * - Navegação por teclado
 */
import { test, expect } from './fixtures.js';

// ── Mock data ─────────────────────────────────────────────────────────────────

const LOJAS_REGISTRO = [
	{
		id: '920292999',
		nome: 'Glory of Seoul',
		nome_normalizado: 'gloryofseoul',
		marketplace: 'shopee',
		cron: '0 */8 * * *',
		origem: '🇰🇷',
		monitorada: true,
		imagem: 'https://img.test/glory.jpg',
		seguidores: 12000,
		total_produtos: 340,
		avaliacao: 4.8
	},
	{
		id: '281000111',
		nome: 'Le Botanic Beauty',
		nome_normalizado: 'lebotanicbeauty',
		marketplace: 'shopee',
		cron: null,
		origem: '🇧🇷',
		monitorada: false,
		imagem: null,
		seguidores: null,
		total_produtos: null,
		avaliacao: null
	}
];

const LOJAS_BUSCAR_RESULT = {
	lojas: [
		{
			id: '920292999',
			nome: 'Glory of Seoul',
			nome_normalizado: 'gloryofseoul',
			marketplace: 'shopee',
			monitorada: true,
			origem: '🇰🇷',
			imagem: 'https://img.test/glory.jpg',
			seguidores: 12000,
			total_produtos: 340,
			avaliacao: 4.8
		}
	],
	total: 1
};

const RESOLVER_RESULT = {
	id: '999888777',
	nome: 'New Resolved Shop',
	nome_normalizado: 'newresolvedshop',
	marketplace: 'shopee',
	cron: null,
	origem: null,
	monitorada: false,
	imagem: 'https://img.test/new.jpg',
	seguidores: 500,
	total_produtos: 42,
	avaliacao: 4.2,
	localizacao: 'SP'
};

// ── Setup ─────────────────────────────────────────────────────────────────────

function setupSmartSearch(page, overrides = {}) {
	page.apiOverrides({
		'/api/lojas/registro': { lojas: LOJAS_REGISTRO, total: LOJAS_REGISTRO.length },
		'/api/lojas/buscar': LOJAS_BUSCAR_RESULT,
		'/api/lojas/resolver': RESOLVER_RESULT,
		'/api/categorias': {
			marketplaces: [
				{
					marketplace: 'shopee',
					categorias: [
						{ id: 100630, nome: 'Beleza' },
						{ id: 100664, nome: 'Cuidados com a Pele' }
					]
				}
			]
		},
		...overrides
	});
}

// ═══════════════════════════════════════════════════════════════════════════════
// 1. SMART DROPDOWN — DETECÇÃO DE INTENÇÃO
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Smart Search — Detecção de Intenção', () => {
	test('digitar 2+ chars mostra opções de intenção', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('serum');

		// Smart Dropdown abre com opções
		const listbox = page.getByRole('listbox');
		await expect(listbox).toBeVisible({ timeout: 5000 });

		// Deve ter pelo menos Produtos e Lojas
		const options = page.getByRole('option');
		expect(await options.count()).toBeGreaterThanOrEqual(2);
	});

	test('opção "Pesquisar em Produtos" aparece primeiro', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('serum');

		const firstOption = page.getByRole('option').first();
		await expect(firstOption).toContainText('Produtos');
	});

	test('opção "Pesquisar em Lojas" aparece', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('serum');

		await expect(page.getByRole('option').filter({ hasText: 'Lojas' })).toBeVisible();
	});

	test('texto < 2 chars não mostra dropdown', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('s');

		await expect(page.getByRole('listbox')).not.toBeVisible();
	});

	test('match de categoria mostra opção adicional', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('beleza');

		// Deve ter opção de categoria
		await expect(page.getByRole('option').filter({ hasText: '#Beleza' })).toBeVisible({ timeout: 5000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 2. URL DETECTION — RESOLVER LINK
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Smart Search — Resolver Link', () => {
	test('colar URL mostra opção "Resolver Link" exclusiva', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('https://shopee.com.br/shop/123456');

		const options = page.getByRole('option');
		expect(await options.count()).toBe(1);
		await expect(options.first()).toContainText('Resolver Link');
	});

	test('link de afiliado (s.shopee.com.br) tambem mostra Resolver Link', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('https://s.shopee.com.br/8fQYnxWQqu');

		const options = page.getByRole('option');
		expect(await options.count()).toBe(1);
		await expect(options.first()).toContainText('Resolver Link');
	});

	test('selecionar Resolver Link resolve e mostra chip da loja', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('https://shopee.com.br/shop/123456');

		await page.getByRole('option').first().click();

		// Apos resolver: chip dourado aparece com nome da loja resolvida
		await expect(page.getByLabel(/Loja:.*New Resolved Shop/)).toBeVisible({ timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 3. BUSCA DE LOJAS — STORE CARDS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Smart Search — Busca de Lojas', () => {
	test('selecionar "Pesquisar em Lojas" mostra Store Cards', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('glory');

		// Seleciona opção "Pesquisar em Lojas"
		await page.getByRole('option').filter({ hasText: 'Lojas' }).click();

		// Store Card com dados da loja
		await expect(page.getByText('Glory of Seoul')).toBeVisible({ timeout: 10000 });
		await expect(page.getByText(/12.*000.*seguidores/)).toBeVisible();
		await expect(page.getByText(/340 produtos/)).toBeVisible();
	});

	test('Store Card mostra bandeira de origem', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('glory');
		await page.getByRole('option').filter({ hasText: 'Lojas' }).click();

		await expect(page.getByTitle('Origem')).toContainText('🇰🇷');
	});

	test('nenhum resultado mostra mensagem de fallback', async ({ authedPage: page }) => {
		setupSmartSearch(page, { '/api/lojas/buscar': { lojas: [], total: 0 } });
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('inexistente');
		await page.getByRole('option').filter({ hasText: 'Lojas' }).click();

		await expect(page.getByText(/Nenhuma loja encontrada/)).toBeVisible({ timeout: 10000 });
		await expect(page.getByText(/colar o link/)).toBeVisible();
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 4. MONITORAMENTO INLINE
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Smart Search — Monitorar Loja', () => {
	test('botão Monitorar visível em loja não monitorada', async ({ authedPage: page }) => {
		setupSmartSearch(page, {
			'/api/lojas/buscar': {
				lojas: [
					{
						id: '281',
						nome: 'Le Botanic',
						marketplace: 'shopee',
						monitorada: false,
						origem: '🇧🇷',
						imagem: null,
						seguidores: null,
						total_produtos: null,
						avaliacao: null
					}
				],
				total: 1
			}
		});
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('botanic');
		await page.getByRole('option').filter({ hasText: 'Lojas' }).click();

		await expect(page.getByRole('button', { name: /monitorar/i })).toBeVisible({ timeout: 10000 });
	});

	test('loja monitorada mostra indicador (sem botão)', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('glory');
		await page.getByRole('option').filter({ hasText: 'Lojas' }).click();

		// Glory of Seoul é monitorada → não deve ter botão
		await expect(page.getByTitle('Monitorada')).toBeVisible({ timeout: 10000 });
		await expect(page.getByRole('button', { name: /monitorar/i })).not.toBeVisible();
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 5. COEXISTÊNCIA COM PREFIXOS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Smart Search — Prefixos', () => {
	test('@loja mostra sugestões de loja (modo prefixo)', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('@gl');

		// Dropdown com sugestão de loja (nome)
		await expect(page.getByRole('listbox')).toBeVisible({ timeout: 5000 });
		await expect(page.getByText('Glory of Seoul')).toBeVisible();
	});

	test('#categoria mostra sugestões de categoria', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('#bel');

		await expect(page.getByRole('listbox')).toBeVisible({ timeout: 5000 });
		await expect(page.getByText('Beleza')).toBeVisible();
	});

	test('selecionar @loja adiciona ao escopo', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('@gl');

		await page.getByText('Glory of Seoul').click();

		// Card de loja no escopo deve aparecer
		await expect(page.getByText('Glory of Seoul')).toBeVisible({ timeout: 5000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 6. TECLADO
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Smart Search — Teclado', () => {
	test('Enter sem seleção executa primeira opção (Produtos)', async ({ authedPage: page }) => {
		setupSmartSearch(page, {
			'/api/candidatos': {
				candidatos: [
					{ id: 'p1', nome: 'Serum X', preco: 50, comissao: 0.1, vendas: 100, loja: 'L', link: '', _fonte: 'curadoria' }
				],
				total_bruto: 1,
				estrategia: 'nicho'
			}
		});
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('serum');
		await input.press('Enter');

		// Dropdown fecha e busca executa
		await expect(page.getByRole('listbox')).not.toBeVisible();
		await expect(page.getByText('Serum X')).toBeVisible({ timeout: 10000 });
	});

	test('ArrowDown + Enter seleciona opção destacada', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('serum');

		// ArrowDown → segunda opção (Lojas)
		await input.press('ArrowDown');
		await input.press('ArrowDown');
		await input.press('Enter');

		// Deve estar em modo lojas
		await expect(page.getByText('Glory of Seoul')).toBeVisible({ timeout: 10000 });
	});

	test('Escape fecha dropdown', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('serum');

		await expect(page.getByRole('listbox')).toBeVisible({ timeout: 5000 });
		await input.press('Escape');
		await expect(page.getByRole('listbox')).not.toBeVisible();
	});

	test('aria-activedescendant atualiza com navegação', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('serum');

		await input.press('ArrowDown');
		await expect(input).toHaveAttribute('aria-activedescendant', 'omnibox-opt-0');
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 7. TRANSIÇÃO LOJAS → PRODUTOS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Smart Search — Transição de modo', () => {
	test('após buscar lojas, digitar keyword volta para produtos', async ({ authedPage: page }) => {
		setupSmartSearch(page, {
			'/api/candidatos': {
				candidatos: [
					{
						id: 'p1',
						nome: 'Retinol ABC',
						preco: 40,
						comissao: 0.08,
						vendas: 50,
						loja: 'X',
						link: '',
						_fonte: 'curadoria'
					}
				],
				total_bruto: 1,
				estrategia: 'nicho'
			}
		});
		await page.goto('/');
		const input = page.getByRole('combobox');

		// Primeiro: busca lojas
		await input.fill('glory');
		await page.getByRole('option').filter({ hasText: 'Lojas' }).click();
		await expect(page.getByText('Glory of Seoul')).toBeVisible({ timeout: 10000 });

		// Depois: digita keyword → volta para produtos
		await input.fill('retinol');
		await input.press('Enter');
		await expect(page.getByText('Retinol ABC')).toBeVisible({ timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 8. MOBILE — BLUR SEM ENTER
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Smart Search — Mobile (blur sem Enter)', () => {
	test('digitar + clicar fora fecha dropdown e busca executa via debounce', async ({ authedPage: page }) => {
		setupSmartSearch(page, {
			'/api/candidatos': {
				candidatos: [
					{
						id: 'p1',
						nome: 'Serum Mobile',
						preco: 50,
						comissao: 0.1,
						vendas: 100,
						loja: 'L',
						link: '',
						_fonte: 'curadoria'
					}
				],
				total_bruto: 1,
				estrategia: 'nicho'
			}
		});
		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('serum');

		// Dropdown esta aberto
		await expect(page.getByRole('listbox')).toBeVisible({ timeout: 5000 });

		// Clicar fora (simula unfocus mobile)
		await page.locator('h1').click();

		// Dropdown fechou
		await expect(page.getByRole('listbox')).not.toBeVisible();

		// Busca executou via debounce (sem Enter)
		await expect(page.getByText('Serum Mobile')).toBeVisible({ timeout: 10000 });
	});

	test('chip X funciona por tap (sem teclado)', async ({ authedPage: page }) => {
		setupSmartSearch(page);
		await page.goto('/');
		const input = page.getByRole('combobox');

		// Adiciona loja via @prefixo
		await input.fill('@gl');
		await page.getByText('Glory of Seoul').click();

		// Chip dourado aparece
		await expect(page.getByLabel(/Loja: Glory of Seoul/)).toBeVisible({ timeout: 5000 });

		// Tap no X do chip
		await page.getByLabel('Remover loja Glory of Seoul').click();

		// Chip removido
		await expect(page.getByLabel(/Loja: Glory of Seoul/)).not.toBeVisible();
	});
});
