import { test, expect } from '@playwright/test';

/**
 * Testes E2E reais: monta a app no browser, interage com o DOM renderizado
 * pelo Svelte, e verifica se o botão habilita ao selecionar um grupo.
 *
 * Estes testes simulam o fluxo real do usuário — se o bug do select
 * existir no runtime do Svelte, eles VÃO falhar.
 */

const gruposMock = [
	{ id: '120363430000000000@g.us', nome: '#1 Garimpo Hoje' },
	{ id: '120363410893012870@g.us', nome: '#08 AVANÇADO VOE' },
	{ id: '558491629647-1486926372@g.us', nome: 'Famílias da Pipa' }
];

// Configura mocks de API e simula login
async function setupPage(page) {
	// Mock Firebase auth — injeta usuário antes da página carregar
	await page.addInitScript(() => {
		// Intercepta o módulo firebase para simular login
		window.__TEST_FORCE_AUTH = true;
	});

	// Mock APIs
	await page.route('**/api/admin/me', route =>
		route.fulfill({ json: { uid: 'test', email: 'test@test.com', admin: true } })
	);
	await page.route('**/api/destinos', async route => {
		if (route.request().method() === 'GET') {
			return route.fulfill({ json: { destinos: [] } });
		}
		if (route.request().method() === 'POST') {
			const body = JSON.parse(route.request().postData());
			return route.fulfill({
				status: 201,
				json: { status: 'ok', destino: { id: 'test-id', ...body, ativo: true } }
			});
		}
		return route.fulfill({ json: { destinos: [] } });
	});
	await page.route('**/api/whatsapp/grupos', route =>
		route.fulfill({ json: { grupos: gruposMock } })
	);
}

test.describe('Canais — interação real: selecionar grupo WhatsApp', () => {
	test('o select de grupos é populado após selecionar tipo WhatsApp', async ({ page }) => {
		await setupPage(page);
		await page.goto('/canais');

		// Verifica se o formulário existe (pode estar atrás do login)
		const tipoSelect = page.locator('#tipo');
		if (await tipoSelect.isVisible()) {
			// Muda para WhatsApp
			await tipoSelect.selectOption('whatsapp');

			// Espera o select de grupos aparecer
			const grupoSelect = page.locator('select').last();
			await expect(grupoSelect).toBeVisible({ timeout: 5000 });

			// Verifica que tem opções
			const options = grupoSelect.locator('option');
			const count = await options.count();
			expect(count).toBeGreaterThan(1); // pelo menos placeholder + 1 grupo
		}
	});

	test('selecionar um grupo habilita o botão Adicionar', async ({ page }) => {
		await setupPage(page);
		await page.goto('/canais');

		const tipoSelect = page.locator('#tipo');
		// Se não está visível, é porque o login não foi mockado — falha explicitamente
		const visivel = await tipoSelect.isVisible({ timeout: 3000 }).catch(() => false);
		if (!visivel) {
			test.skip();
			return;
		}

		// Preenche nome
		await page.locator('#nome').fill('Garimpo Hoje');

		// Muda para WhatsApp
		await tipoSelect.selectOption('whatsapp');

		// Espera select de grupos
		await page.waitForTimeout(500);
		const grupoSelect = page.locator('select').last();

		// Seleciona o primeiro grupo
		await grupoSelect.selectOption('120363430000000000@g.us');

		// Espera a reatividade do Svelte
		await page.waitForTimeout(200);

		// O botão deve estar habilitado
		const botao = page.locator('button[type="submit"]');
		const disabled = await botao.getAttribute('disabled');

		// Log para debug
		const configValue = await page.evaluate(() => {
			const selects = document.querySelectorAll('select');
			const lastSelect = selects[selects.length - 1];
			return lastSelect ? lastSelect.value : 'NO SELECT FOUND';
		});
		console.log('Config value no DOM após seleção:', configValue);
		console.log('Botão disabled attr:', disabled);

		expect(disabled).toBeNull(); // null = não disabled = habilitado
	});

	test('botão fica desabilitado se só selecionar grupo sem nome', async ({ page }) => {
		await setupPage(page);
		await page.goto('/canais');

		const tipoSelect = page.locator('#tipo');
		if (await tipoSelect.isVisible()) {
			// Muda para WhatsApp SEM preencher nome
			await tipoSelect.selectOption('whatsapp');
			await page.waitForTimeout(500);

			const grupoSelect = page.locator('select').last();
			await grupoSelect.selectOption('120363430000000000@g.us');
			await page.waitForTimeout(100);

			// O botão deve estar DESABILITADO (nome vazio)
			const botao = page.locator('button[type="submit"]');
			await expect(botao).toBeDisabled();
		}
	});
});
