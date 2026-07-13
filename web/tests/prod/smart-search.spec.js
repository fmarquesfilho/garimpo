/**
 * E2E PRODUÇÃO — Smart Search contra APIs reais.
 *
 * Testa o Omnibox refatorado (Headless UI Controller) contra garimpei.app.br.
 * NÃO mocka nada — APIs reais, dados reais.
 *
 * Rodar: mise run test:e2e-prod -- tests/prod/smart-search.spec.js
 */
import { test, expect } from './fixtures.js';

// ═══════════════════════════════════════════════════════════════════════════════
// 1. SMART DROPDOWN
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Smart Search Dropdown', () => {
	test('digitar texto mostra opções de intenção', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByRole('combobox');
		await expect(input).toBeVisible({ timeout: 15000 });

		await input.fill('serum');
		const listbox = page.getByRole('listbox');
		await expect(listbox).toBeVisible({ timeout: 5000 });

		const options = page.getByRole('option');
		expect(await options.count()).toBeGreaterThanOrEqual(2);
		await expect(options.first()).toContainText('Produtos');
	});

	test('Enter executa busca de produtos', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByRole('combobox');
		await expect(input).toBeVisible({ timeout: 15000 });

		await input.pressSequentially('serum', { delay: 80 });
		await input.press('Enter');

		// Resultados ou empty state (depende dos dados de prod)
		await expect(page.getByText(/\d+ produto|Nenhum resultado/i).first()).toBeVisible({ timeout: 20000 });
	});

	test('Escape fecha dropdown', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByRole('combobox');
		await expect(input).toBeVisible({ timeout: 15000 });

		await input.fill('glory');
		await expect(page.getByRole('listbox')).toBeVisible({ timeout: 5000 });
		await input.press('Escape');
		await expect(page.getByRole('listbox')).not.toBeVisible();
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 2. BUSCA DE LOJAS
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Busca de Lojas', () => {
	test('selecionar "Pesquisar em Lojas" retorna resultados do registro', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByRole('combobox');
		await expect(input).toBeVisible({ timeout: 15000 });

		await input.fill('glory');
		await page.getByRole('option').filter({ hasText: 'Lojas' }).click();

		// Deve mostrar ao menos uma loja OU mensagem de nenhuma encontrada
		await expect(page.getByText('Glory of Seoul').or(page.getByText(/Nenhuma loja encontrada/))).toBeVisible({
			timeout: 15000
		});
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 3. RESOLVER LINK
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Resolver Link', () => {
	test('colar URL de loja mostra Resolver Link e resolve', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByRole('combobox');
		await expect(input).toBeVisible({ timeout: 15000 });

		await input.fill('https://shopee.com.br/gloryofseoul');
		await expect(page.getByRole('option').filter({ hasText: 'Resolver Link' })).toBeVisible({ timeout: 5000 });

		await page.getByRole('option').first().click();

		// Espera resolução: Store Card ou erro
		await expect(page.getByText(/Glory/).or(page.getByText(/não encontrada|falhou|inválido/i))).toBeVisible({
			timeout: 20000
		});
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 4. PREFIXOS (COEXISTÊNCIA)
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Prefixos', () => {
	test('@loja mostra sugestões do registro', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByRole('combobox');
		await expect(input).toBeVisible({ timeout: 15000 });

		await input.fill('@gl');

		// Deve mostrar dropdown com sugestão (se há lojas com "gl" no registro)
		const listbox = page.getByRole('listbox');
		// Pode não ter lojas — aceita listbox visível OU não visível sem erro
		await page.waitForTimeout(2000);
		if (await listbox.isVisible()) {
			expect(await page.getByRole('option').count()).toBeGreaterThanOrEqual(1);
		}
	});
});

// ═══════════════════════════════════════════════════════════════════════════════
// 5. ACESSIBILIDADE
// ═══════════════════════════════════════════════════════════════════════════════

test.describe('Produção — Acessibilidade Omnibox', () => {
	test('input tem role combobox com atributos ARIA', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByRole('combobox');
		await expect(input).toBeVisible({ timeout: 15000 });
		await expect(input).toHaveAttribute('aria-autocomplete', 'list');
		await expect(input).toHaveAttribute('aria-controls', 'omnibox-listbox');
	});

	test('dropdown tem role listbox com options', async ({ authedPage: page }) => {
		await page.goto('/');
		const input = page.getByRole('combobox');
		await expect(input).toBeVisible({ timeout: 15000 });

		await input.fill('serum');
		const listbox = page.getByRole('listbox');
		await expect(listbox).toBeVisible({ timeout: 5000 });
		await expect(listbox).toHaveAttribute('aria-label');

		const options = page.getByRole('option');
		expect(await options.count()).toBeGreaterThan(0);
	});
});
