/**
 * E2E Local — Helpers e seletores da UI v4 (Omnibox headless + chips inline).
 */
import { expect } from '@playwright/test';

// ── Seletores ────────────────────────────────────────────────────────────────

export const SEL = {
	omnibox: '[role="combobox"]',
	searchInput: '[role="combobox"]',
	btnBuscasSalvas: 'button:has-text("💾")',
	btnLimparTudo: 'button:has-text("✕ limpar tudo")',
	btnRodar: 'button:has-text("rodar")',
	btnEditar: '[aria-label="Editar busca"]',
	btnRemover: '[aria-label="Remover busca"]'
};

// ── Page helpers ─────────────────────────────────────────────────────────────

/**
 * Espera a engine terminar a inicializacao (buscas e categorias carregadas).
 */
export async function waitForEngineReady(page) {
	const btn = page.locator(SEL.btnBuscasSalvas);
	await expect(btn).toBeVisible({ timeout: 10000 });
}

/**
 * Abre o painel de buscas salvas.
 */
export async function abrirPainelBuscas(page) {
	const btn = page.locator(SEL.btnBuscasSalvas);
	await expect(btn).toBeVisible({ timeout: 10000 });
	await expect(btn).not.toContainText('0 salvas', { timeout: 10000 });
	await btn.click();
	await expect(page.getByText(/coleta periódica|busca manual salva/).first()).toBeVisible({ timeout: 5000 });
}

/**
 * Adiciona loja via Omnibox (resolve link). Usa o Smart Search flow.
 */
export async function adicionarLojaViaOmnibox(page, url) {
	const input = page.getByRole('combobox');
	await input.fill(url);
	await expect(page.getByRole('option').filter({ hasText: 'Resolver Link' })).toBeVisible({ timeout: 5000 });
	await page.getByRole('option').first().click();
}
