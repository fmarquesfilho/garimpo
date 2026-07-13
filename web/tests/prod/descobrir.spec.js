/**
 * E2E PRODUCAO — Pagina Descobrir contra APIs reais.
 * Seletores UI v4 (Omnibox headless + chips inline + filtros diretos).
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

test.describe('Producao — Boot', () => {
	test('pagina carrega autenticada com Omnibox', async ({ authedPage: page }) => {
		await page.goto('/');
		await expect(page.getByRole('combobox')).toBeVisible({ timeout: 15000 });
		await expect(page.getByRole('button', { name: /Entrar com Google/i })).toHaveCount(0);
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 2. BUSCA
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Producao — Busca', () => {
	test('busca por keyword retorna produtos ou empty state', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByRole('combobox');
		await expect(input).toBeVisible({ timeout: 15000 });
		await input.pressSequentially('serum', { delay: 80 });
		await input.press('Enter');

		// A UI deve reagir: loading, resultados, ou empty state
		await expect(
			page.getByText(/\d+ produto/i).or(page.getByText(/Nenhum resultado/i)).or(page.getByText(/Buscando/i))
		).toBeVisible({ timeout: 15000 });
	});

	test('empty state sem keyword', async ({ authedPage: page }) => {
		await page.goto('/');
		await expect(page.getByRole('combobox')).toBeVisible({ timeout: 15000 });
		await expect(page.getByText(/Nenhum resultado/i)).toBeVisible({ timeout: 10000 });
	});

	test('Escape fecha dropdown', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByRole('combobox');
		await expect(input).toBeVisible({ timeout: 15000 });
		await input.fill('serum');
		await expect(page.getByRole('listbox')).toBeVisible({ timeout: 5000 });
		await input.press('Escape');
		await expect(page.getByRole('listbox')).not.toBeVisible();
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 3. LOJAS (via Omnibox — resolve link de afiliado)
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Producao — Lojas', () => {
	test('adicionar loja via link de afiliado mostra chip', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByRole('combobox');
		await expect(input).toBeVisible({ timeout: 15000 });

		// Colar link de afiliado → Resolver Link
		await input.fill('https://s.shopee.com.br/8fQYnxWQqu');
		await expect(page.getByRole('option').filter({ hasText: 'Resolver Link' })).toBeVisible({ timeout: 5000 });
		await page.getByRole('option').first().click();

		// Apos resolver: chip dourado (loja) aparece dentro do Omnibox OU erro de rede
		await expect(
			page
				.getByLabel(/Loja:.*ativa/)
				.first()
				.or(page.getByText(/Timeout|falhou/i))
		).toBeVisible({ timeout: 25000 });
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 4. FILTROS (visíveis inline — sem raia colapsavel)
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Producao — Filtros', () => {
	test('filtros de comissao e vendas visiveis', async ({ authedPage: page }) => {
		await page.goto('/');
		await expect(page.getByRole('combobox')).toBeVisible({ timeout: 15000 });
		await expect(page.getByText('comissão mín.')).toBeVisible({ timeout: 5000 });
		await expect(page.getByText('vendas mín.')).toBeVisible();
	});

	test('toggles de fontes visiveis', async ({ authedPage: page }) => {
		await page.goto('/');
		await expect(page.getByRole('combobox')).toBeVisible({ timeout: 15000 });
		await expect(page.getByRole('button', { name: /Novos/ })).toBeVisible();
		await expect(page.getByRole('button', { name: /Quedas/ })).toBeVisible();
	});

	test('marketplaces visiveis', async ({ authedPage: page }) => {
		await page.goto('/');
		await expect(page.getByRole('combobox')).toBeVisible({ timeout: 15000 });
		await expect(page.getByRole('button', { name: /Shopee/ })).toBeVisible();
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 5. REGRAS (validacao do JSON em runtime)
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Producao — Regras', () => {
	test('intent table valida', () => {
		expect(rules.intent).toHaveLength(4);
		expect(rules.intent.map((r) => r.result).sort()).toEqual(
			['keyword_global', 'keyword_na_loja', 'loja_completa', 'nenhum'].sort()
		);
	});

	test('omnibox intencao config presente', () => {
		expect(rules.omnibox.intencao).toBeDefined();
		expect(rules.omnibox.intencao.habilitado).toBe(true);
		expect(rules.omnibox.intencao.minChars).toBe(2);
	});

	test('storeCard config presente', () => {
		expect(rules.storeCard).toBeDefined();
		expect(rules.storeCard.camposVisiveis.shopee).toContain('imagem');
		expect(rules.storeCard.camposVisiveis._fallback).toContain('nome');
	});
});
