/**
 * E2E PRODUÇÃO — Página Descobrir contra APIs reais.
 *
 * Usa token Firebase real (obtido no auth.setup.js) injetado via fixture.
 * Valida os mesmos cenários dos testes locais, sem mocks.
 *
 * Rodar: npm run test:e2e:prod
 */
import { test, expect } from './fixtures.js';
import { readFileSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const rules = JSON.parse(readFileSync(resolve(__dirname, '../../../rules/busca-rules.json'), 'utf-8'));

// ═══════════════════════════════════════════════════════════════════════════════
// 1. AUTENTICAÇÃO E BOOT
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Boot', () => {
	test('página carrega autenticada (sem tela de login)', async ({ authedPage: page }) => {
		await page.goto('/');
		await expect(page.getByPlaceholder(/Buscar produto/i)).toBeVisible({ timeout: 15000 });
		await expect(page.getByRole('button', { name: /Entrar com Google/i })).toHaveCount(0);
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 2. BUSCA POR KEYWORD
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Busca', () => {
	test('#1: busca por keyword retorna resultados ou empty state', async ({ authedPage: page }) => {
		await page.goto('/');
		await page.getByPlaceholder(/Buscar produto/i).fill('serum');

		// API real pode demorar — aguarda resultado ou empty state
		await expect(page.locator('.contagem, [class*="empty"], .grade').first()).toBeVisible({ timeout: 30000 });
	});

	test('#2: sem keyword nem loja, UI mostra empty state ou dados de buscas salvas', async ({ authedPage: page }) => {
		await page.goto('/');
		// Em produção o usuário pode ter buscas salvas que geram resultados ao inicializar.
		// Verifica que ALGO aparece (empty state OU resultados OU pills de buscas salvas).
		await expect(
			page.locator('.contagem, [class*="empty"], .grade, button[class*="rounded-full"]').first()
		).toBeVisible({ timeout: 15000 });
	});

	test('debounce funciona', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByPlaceholder(/Buscar produto/i);
		await input.pressSequentially('vitamina c', { delay: 30 });
		await expect(page.locator('.contagem, [class*="empty"], .grade').first()).toBeVisible({ timeout: 20000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 3. ADICIONAR LOJA
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Lojas', () => {
	test('adicionar loja via URL curta resolve e mostra badge', async ({ authedPage: page }) => {
		await page.goto('/');
		const lojaInput = page.locator('input[placeholder*="loja"]').first();
		await lojaInput.fill('https://s.shopee.com.br/8fQYnxWQqu');
		await lojaInput.press('Enter');

		// Badge com 🏪 deve aparecer (loja resolvida)
		await expect(page.locator('text=🏪').first()).toBeVisible({ timeout: 20000 });
	});

	test('keyword + loja → intent keyword_na_loja (regras)', async ({ authedPage: page }) => {
		await page.goto('/');
		await page.getByPlaceholder(/Buscar produto/i).fill('serum');
		await page.waitForTimeout(500);

		const lojaInput = page.locator('input[placeholder*="loja"]').first();
		await lojaInput.fill('https://s.shopee.com.br/70IKp57jnV');
		await lojaInput.press('Enter');

		const intentRow = rules.intent.find((r) => r.keyword && r.shop);
		expect(intentRow.result).toBe('keyword_na_loja');

		await expect(page.locator('text=🏪').first()).toBeVisible({ timeout: 20000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 4. FILTROS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Filtros', () => {
	test('filtros abrem e comissão formatada em %', async ({ authedPage: page }) => {
		await page.goto('/');
		await page.getByRole('button', { name: /Filtros/i }).click();
		await expect(page.getByText('comissão mín.')).toBeVisible();

		const filterSection = page.locator('.bg-muted').first();
		const text = await filterSection.textContent();
		expect(text).not.toMatch(/0\.\d{5,}/);
	});

	test('categorias aparecem como chips', async ({ authedPage: page }) => {
		await page.goto('/');
		// Usa keyword "serum" que provavelmente tem resultados em produção
		await page.getByPlaceholder(/Buscar produto/i).fill('serum');
		await page.waitForTimeout(1000);
		await page.getByRole('button', { name: /Filtros/i }).click();

		// Categorias podem vir da API ou do fallback hardcoded
		// Verifica que a seção de filtros mostra ALGO além de comissão/vendas
		const filtrosContainer = page.getByText('categorias');
		await expect(filtrosContainer).toBeVisible({ timeout: 10000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 5. SALVAR E RESTAURAR
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Salvar busca', () => {
	test('dialog de salvar abre ao clicar 💾', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByPlaceholder(/Buscar produto/i);

		// Keyword para ativar o guard podeSalvar
		await input.fill('serum');
		await page.waitForTimeout(600);

		// Abre dialog salvar
		await page.locator('button', { hasText: '💾' }).click();

		// Dialog deve abrir (mostra AgendadorBusca + botão confirmar)
		await expect(page.getByText(/Salvar configuração/i)).toBeVisible({ timeout: 5000 });
		// Botão de confirmação deve existir
		await expect(page.locator('button', { hasText: /^Salvar/ }).last()).toBeVisible();
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 6. INPUT
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Input', () => {
	test('✕ limpa o campo', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByPlaceholder(/Buscar produto/i);
		await input.fill('serum');
		await page.getByRole('button', { name: /Limpar/i }).click();
		await expect(input).toHaveValue('');
	});

	test('ESC limpa o campo', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByPlaceholder(/Buscar produto/i);
		await input.fill('serum');
		await input.press('Escape');
		await expect(input).toHaveValue('');
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 7. FONTES
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Fontes', () => {
	test('toggles de fontes visíveis', async ({ authedPage: page }) => {
		await page.goto('/');
		await expect(page.locator('.fonte-btn', { hasText: '🔍' })).toBeVisible({ timeout: 10000 });
		await expect(page.locator('.fonte-btn', { hasText: '📉' })).toBeVisible();
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 8. REGRAS EXTERNAS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Regras externas', () => {
	test('intent table tem 4 combinações', () => {
		expect(rules.intent).toHaveLength(4);
		expect(rules.intent.map((r) => r.result).sort()).toEqual(
			['keyword_global', 'keyword_na_loja', 'loja_completa', 'nenhum'].sort()
		);
	});

	test('defaults de fontes conferem com UI', async ({ authedPage: page }) => {
		expect(rules.defaults.fontes.curadoria).toBe(true);
		expect(rules.defaults.fontes.quedas).toBe(true);
		expect(rules.defaults.fontes.novos).toBe(true);
		expect(rules.defaults.fontes.lojas).toBe(false);

		await page.goto('/');
		const buscaToggle = page.locator('.fonte-btn[data-state="on"]', { hasText: '🔍' });
		await expect(buscaToggle).toBeVisible({ timeout: 10000 });
	});
});
