/**
 * E2E PRODUÇÃO — Página Descobrir contra APIs reais.
 * Seletores atualizados para UI v3 (raias, BuscaCards).
 *
 * Rodar: mise run test:e2e-prod
 */
import { test, expect } from './fixtures.js';
import { readFileSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const rules = JSON.parse(readFileSync(resolve(__dirname, '../../../rules/busca-rules.json'), 'utf-8'));

// ═══════════════════════════════════════════════════════════════════════════════
// 1. BOOT
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Boot', () => {
	test('página carrega autenticada', async ({ authedPage: page }) => {
		await page.goto('/');
		await expect(page.locator('input[type="search"]')).toBeVisible({ timeout: 15000 });
		await expect(page.getByRole('button', { name: /Entrar com Google/i })).toHaveCount(0);
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 2. BUSCA
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Busca', () => {
	test('busca por keyword retorna produtos', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.locator('input[type="search"]');
		await input.fill('serum');

		// Com o fix: init não executa busca automática, então não há race condition.
		// A busca é disparada pelo debounce do DIGITAR → deve completar em <10s.
		await expect(page.getByText(/\d+ produto/i).first()).toBeVisible({ timeout: 15000 });
	});

	test('empty state sem keyword', async ({ authedPage: page }) => {
		await page.goto('/');
		// Sem keyword e sem buscas salvas → empty state após engine inicializar
		await expect(page.getByText('Buscando produtos')).not.toBeVisible({ timeout: 30000 });
		await expect(page.getByText(/Nenhum resultado/i)).toBeVisible({ timeout: 5000 });
	});

	test('ESC limpa keyword', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.locator('input[type="search"]');
		await input.fill('serum');
		await input.press('Escape');
		await expect(input).toHaveValue('');
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 3. LOJAS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Lojas', () => {
	test('adicionar loja via URL resolve e mostra badge', async ({ authedPage: page }) => {
		await page.goto('/');
		await page.locator('button:has-text("🏪 Lojas")').click();
		const lojaInput = page.locator('input[placeholder*="loja"]').first();
		await lojaInput.fill('https://s.shopee.com.br/8fQYnxWQqu');
		await lojaInput.press('Enter');
		await expect(page.getByText('🏪').first()).toBeVisible({ timeout: 20000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 4. FILTROS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Filtros', () => {
	test('raia filtros abre e mostra comissão', async ({ authedPage: page }) => {
		await page.goto('/');
		await page.locator('button:has-text("⚙️ Filtros")').click();
		await expect(page.getByText('comissão mín.')).toBeVisible({ timeout: 5000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 5. REGRAS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Regras', () => {
	test('intent table valida', () => {
		expect(rules.intent).toHaveLength(4);
		expect(rules.intent.map((r) => r.result).sort()).toEqual(
			['keyword_global', 'keyword_na_loja', 'loja_completa', 'nenhum'].sort()
		);
	});
});
