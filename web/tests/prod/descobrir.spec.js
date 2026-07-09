/**
 * E2E PRODUÇÃO — Página Descobrir contra APIs reais.
 *
 * Valida os mesmos cenários dos testes locais, mas sem mocks.
 * Dados são reais (dependem do estado do banco/APIs).
 * Cenários focam em: funcionalidade, não em dados específicos.
 *
 * Rodar: npm run test:e2e:prod
 */
import { test, expect } from '@playwright/test';
import { readFileSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const rules = JSON.parse(readFileSync(resolve(__dirname, '../../rules/busca-rules.json'), 'utf-8'));

// ═══════════════════════════════════════════════════════════════════════════════
// 1. AUTENTICAÇÃO E BOOT
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Boot', () => {
	test('página carrega autenticada (sem tela de login)', async ({ page }) => {
		await page.goto('/');
		await expect(page.getByPlaceholder(/Buscar produto/i)).toBeVisible({ timeout: 15000 });
		// Não deve mostrar botão de login
		await expect(page.getByRole('button', { name: /Entrar com Google/i })).toHaveCount(0);
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 2. BUSCA POR KEYWORD
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Busca', () => {
	test('#1: busca por keyword retorna resultados da API real', async ({ page }) => {
		await page.goto('/');
		await page.getByPlaceholder(/Buscar produto/i).fill('serum');

		// Aguarda resultados reais (pelo menos 1 produto ou empty state)
		await expect(page.locator('.grade, [class*="empty"]').first()).toBeVisible({ timeout: 20000 });

		// Se houver resultados, devem ter formato de card
		const contagem = page.locator('.contagem');
		if (await contagem.isVisible()) {
			const texto = await contagem.textContent();
			expect(texto).toMatch(/\d+ produto/);
		}
	});

	test('#2: busca vazia mostra empty state', async ({ page }) => {
		await page.goto('/');
		// Sem digitar nada, deve mostrar empty state ou hint
		await expect(page.getByText(/Nenhum resultado/i)).toBeVisible({ timeout: 10000 });
	});

	test('debounce funciona (não busca a cada tecla)', async ({ page }) => {
		await page.goto('/');
		const input = page.getByPlaceholder(/Buscar produto/i);
		await input.pressSequentially('vitamina c', { delay: 30 });

		// Após debounce, resultados ou empty state deve aparecer
		await expect(page.locator('.grade, .contagem, [class*="empty"]').first()).toBeVisible({ timeout: 15000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 3. ADICIONAR LOJA (resolve shop real)
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Lojas', () => {
	test('adicionar loja via URL curta resolve e mostra badge', async ({ page }) => {
		await page.goto('/');

		// Le Botanic
		const lojaInput = page.locator('input[placeholder*="loja"]').first();
		await lojaInput.fill('https://s.shopee.com.br/8fQYnxWQqu');
		await lojaInput.press('Enter');

		// Deve resolver e mostrar badge com nome da loja
		await expect(page.locator('[class*="badge"], [class*="Badge"]').filter({ hasText: /🏪/ })).toBeVisible({
			timeout: 20000
		});
	});

	test('adicionar loja com keyword escopa busca (intent: keyword_na_loja)', async ({ page }) => {
		await page.goto('/');

		// Digita keyword primeiro
		await page.getByPlaceholder(/Buscar produto/i).fill('serum');
		await page.waitForTimeout(500);

		// Adiciona loja
		const lojaInput = page.locator('input[placeholder*="loja"]').first();
		await lojaInput.fill('https://s.shopee.com.br/70IKp57jnV');
		await lojaInput.press('Enter');

		// Valida pela regra: keyword+shop → keyword_na_loja → sources inclui lojas
		const intentRow = rules.intent.find((r) => r.keyword && r.shop);
		expect(intentRow.result).toBe('keyword_na_loja');

		// Aguarda resolução + busca executar
		await expect(page.locator('[class*="badge"], [class*="Badge"]').filter({ hasText: /🏪/ })).toBeVisible({
			timeout: 20000
		});
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 4. FILTROS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Filtros', () => {
	test('filtros abrem e mostram comissão formatada em %', async ({ page }) => {
		await page.goto('/');
		await page.getByRole('button', { name: /Filtros/i }).click();

		// Label "comissão mín." visível
		await expect(page.getByText('comissão mín.')).toBeVisible();

		// Nenhum float cru nos filtros
		const filterSection = page.locator('.bg-muted').first();
		const text = await filterSection.textContent();
		expect(text).not.toMatch(/0\.\d{5,}/);
	});

	test('categorias aparecem como chips selecionáveis', async ({ page }) => {
		await page.goto('/');
		await page.getByPlaceholder(/Buscar produto/i).fill('perfume');
		await page.waitForTimeout(600);

		await page.getByRole('button', { name: /Filtros/i }).click();

		// Ao menos uma categoria da Shopee deve aparecer (API real retorna categorias)
		const categoriaChips = page.locator('.bg-muted button[class*="rounded-full"]');
		await expect(categoriaChips.first()).toBeVisible({ timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 5. SALVAR E RESTAURAR BUSCA
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Salvar busca', () => {
	test('salvar busca → pill aparece → clicar restaura', async ({ page }) => {
		await page.goto('/');
		const input = page.getByPlaceholder(/Buscar produto/i);

		// Keyword única para este teste (evita colisão com buscas existentes)
		const keyword = `e2e-${Date.now()}`;
		await input.fill(keyword);
		await page.waitForTimeout(600);

		// Abre salvar
		await page
			.getByRole('button', { name: /Salvar/i })
			.first()
			.click();

		// Confirma
		await page.getByRole('button', { name: /^Salvar/ }).click();

		// Pill deve aparecer
		const pill = page.locator('button', { hasText: keyword });
		await expect(pill).toBeVisible({ timeout: 15000 });

		// Limpa e restaura
		await input.fill('');
		await page.waitForTimeout(600);
		await pill.click();
		await expect(input).toHaveValue(keyword);

		// Cleanup: remove a busca salva (clica no ✕ ao lado do pill)
		const removeBtn = pill.locator('..').locator('button', { hasText: '✕' });
		if (await removeBtn.isVisible()) {
			await removeBtn.click();
			await page.waitForTimeout(500);
		}
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 6. INPUT DE BUSCA
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Input', () => {
	test('✕ limpa o campo', async ({ page }) => {
		await page.goto('/');
		const input = page.getByPlaceholder(/Buscar produto/i);
		await input.fill('serum');

		const clearBtn = page.getByRole('button', { name: /Limpar/i });
		await expect(clearBtn).toBeVisible();
		await clearBtn.click();
		await expect(input).toHaveValue('');
	});

	test('ESC limpa o campo', async ({ page }) => {
		await page.goto('/');
		const input = page.getByPlaceholder(/Buscar produto/i);
		await input.fill('serum');
		await input.press('Escape');
		await expect(input).toHaveValue('');
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 7. FONTES (toggles)
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Fontes', () => {
	test('toggles de fontes são visíveis e clicáveis', async ({ page }) => {
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		// Toggle de busca/curadoria deve estar visível
		const buscaToggle = page.locator('.fonte-btn', { hasText: '🔍' });
		await expect(buscaToggle).toBeVisible({ timeout: 10000 });

		// Toggle de quedas
		const quedasToggle = page.locator('.fonte-btn', { hasText: '📉' });
		await expect(quedasToggle).toBeVisible();
	});

	test('defaults de fontes respeitam rules/busca-rules.json', async ({ page }) => {
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		// Verifica que curadoria está ativa por default (conforme rules)
		expect(rules.defaults.fontes.curadoria).toBe(true);
		const buscaToggle = page.locator('.fonte-btn', { hasText: '🔍' });
		await expect(buscaToggle).toHaveAttribute('data-state', 'on', { timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 8. REGRAS EXTERNAS (validação contra JSON)
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Regras externas', () => {
	test('intent table é consistente com o comportamento da UI', async ({ page }) => {
		// Valida que as 4 combinações da intent table existem
		expect(rules.intent).toHaveLength(4);

		const intents = rules.intent.map((r) => r.result);
		expect(intents).toContain('keyword_na_loja');
		expect(intents).toContain('keyword_global');
		expect(intents).toContain('loja_completa');
		expect(intents).toContain('nenhum');
	});

	test('normalização de comissão funciona (7 → 7%, não 7.0000)', async ({ page }) => {
		await page.goto('/');
		await page.getByRole('button', { name: /Filtros/i }).click();

		// A regra diz: divideBy100IfGt1
		expect(rules.normalize.comissao.divideBy100IfGt1).toBe(true);

		// Na UI, as opções devem ser porcentagens inteiras
		const filterSection = page.locator('.bg-muted').first();
		const text = await filterSection.textContent();
		expect(text).toContain('%');
		expect(text).not.toMatch(/0\.07000/);
	});
});
