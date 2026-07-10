/**
 * E2E Local — Helpers e seletores da UI v3.
 *
 * Centraliza seletores para manutenção. Quando a UI mudar,
 * basta atualizar este arquivo.
 */
import { expect } from '@playwright/test';

// ── Seletores ────────────────────────────────────────────────────────────────

export const SEL = {
	searchInput: 'input[type="search"]',
	lojaInput: 'input[placeholder*="loja"]',

	btnFiltros: 'button:has-text("⚙️ Filtros")',
	btnLojas: 'button:has-text("🏪 Lojas")',
	btnBuscasSalvas: 'button:has-text("💾")',
	btnLimparTudo: 'button:has-text("✕ limpar tudo")',

	btnRodar: 'button:has-text("rodar")',
	btnEditar: '[aria-label="Editar busca"]',
	btnRemover: '[aria-label="Remover busca"]'
};

// ── Page helpers ─────────────────────────────────────────────────────────────

/**
 * Espera a engine terminar a inicialização (buscas e categorias carregadas).
 * NÃO espera o executarBusca terminar — isso pode ser lento com muitas lojas.
 */
export async function waitForEngineReady(page) {
	// A engine está pronta quando o botão 💾 mostra o contador
	const btn = page.locator(SEL.btnBuscasSalvas);
	await expect(btn).toBeVisible({ timeout: 10000 });
}

/**
 * Abre o painel de buscas salvas.
 * Pré-condição: engine inicializada (buscas carregadas do mock).
 */
export async function abrirPainelBuscas(page) {
	const btn = page.locator(SEL.btnBuscasSalvas);
	await expect(btn).toBeVisible({ timeout: 10000 });
	await expect(btn).not.toContainText('0 salvas', { timeout: 10000 });
	await btn.click();
	// Painel renderiza BuscaCards
	await expect(page.getByText(/coleta periódica|busca manual salva/).first()).toBeVisible({ timeout: 5000 });
}

/**
 * Abre a raia de Lojas.
 */
export async function abrirRaiaLojas(page) {
	await page.locator(SEL.btnLojas).click();
}

/**
 * Abre a raia de Filtros.
 */
export async function abrirRaiaFiltros(page) {
	await page.locator(SEL.btnFiltros).click();
}

/**
 * Preenche o input de loja e submete (raia Lojas precisa estar aberta).
 */
export async function adicionarLoja(page, valor) {
	const input = page.locator(SEL.lojaInput).first();
	await input.fill(valor);
	await input.press('Enter');
}
