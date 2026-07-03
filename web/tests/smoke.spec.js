import { test, expect } from '@playwright/test';

// ── Testes de smoke: verificam que as páginas carregam sem crash ──────────

test.describe('Landing page (não logado)', () => {
	test('mostra botão de login e não mostra menu', async ({ page }) => {
		await page.goto('/');
		// Landing page deve ter botão de entrar
		await expect(page.locator('text=Entrar com Google')).toBeVisible();
		// Não deve mostrar o menu hamburger
		await expect(page.locator('.hamburguer')).not.toBeVisible();
	});

	test('mostra o nome do app', async ({ page }) => {
		await page.goto('/');
		await expect(page.locator('.marca')).toContainText('Garimpei');
	});

	test('mostra features na landing', async ({ page }) => {
		await page.goto('/');
		await expect(page.locator('.hero-features')).toBeVisible();
		await expect(page.locator('.feature')).toHaveCount(4);
	});

	test('não mostra conteúdo do app sem login', async ({ page }) => {
		await page.goto('/');
		// A barra de filtros NÃO deve estar visível
		await expect(page.locator('.filtros')).not.toBeVisible();
	});
});

test.describe('Páginas protegidas (sem login)', () => {
	const rotas = [
		'/lojas',
		'/publicar',
		'/publicacoes',
		'/coletas',
		'/estatisticas',
		'/canais',
		'/admin',
		'/oportunidades'
	];

	for (const rota of rotas) {
		test(`${rota} não mostra conteúdo sem login`, async ({ page }) => {
			await page.goto(rota);
			// Deve mostrar landing (botão entrar) em vez do conteúdo
			await expect(page.locator('text=Entrar com Google')).toBeVisible();
		});
	}
});

test.describe('Build estático', () => {
	test('CSS carrega (design tokens presentes)', async ({ page }) => {
		await page.goto('/');
		// Verifica que o CSS foi carregado checando uma propriedade custom
		const bg = await page.evaluate(() => getComputedStyle(document.body).backgroundColor);
		// porcelana #f6f1ef = rgb(246, 241, 239)
		expect(bg).not.toBe('rgba(0, 0, 0, 0)');
	});

	test('nenhum erro JS no console', async ({ page }) => {
		const errors = [];
		page.on('pageerror', (err) => errors.push(err.message));
		await page.goto('/');
		await page.waitForTimeout(1000);
		expect(errors).toHaveLength(0);
	});
});
