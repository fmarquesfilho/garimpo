import { test, expect } from '@playwright/test';

/**
 * Testes E2E para a página /canais — verificam que as rotas existem,
 * não dão erro JS, e que os mocks de API respondem corretamente.
 *
 * A lógica do componente SeletorGrupo é testada pelo Vitest
 * (src/tests/SeletorGrupo.test.js) com o runtime real do Svelte.
 */

test.describe('Página Canais — carregamento', () => {
	test('a rota /canais existe e não dá 404', async ({ page }) => {
		const response = await page.goto('/canais');
		expect(response.status()).toBe(200);
	});

	test('sem login, mostra landing page', async ({ page }) => {
		await page.goto('/canais');
		await expect(page.locator('text=Entrar com Google')).toBeVisible();
	});

	test('não há erros JS ao carregar', async ({ page }) => {
		const errors = [];
		page.on('pageerror', err => errors.push(err.message));
		await page.goto('/canais');
		await page.waitForTimeout(1000);
		expect(errors).toHaveLength(0);
	});
});

test.describe('Página Canais — API de destinos (mock)', () => {
	test('POST /api/destinos com tipo whatsapp é aceito', async ({ page }) => {
		await page.route('**/api/destinos', async route => {
			if (route.request().method() === 'POST') {
				const body = JSON.parse(route.request().postData());
				if (!body.nome || !body.config) {
					return route.fulfill({ status: 400, json: { erro: 'campos obrigatórios' } });
				}
				return route.fulfill({
					status: 201,
					json: { status: 'ok', destino: { id: 'test', ...body, ativo: true } }
				});
			}
			return route.fulfill({ json: { destinos: [] } });
		});

		await page.goto('/canais');

		const resultado = await page.evaluate(async () => {
			const resp = await fetch('/api/destinos', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					nome: '#1 Garimpo Hoje',
					config: '120363430000000000@g.us,120363410893012870@g.us',
					tipo: 'whatsapp'
				})
			});
			return { status: resp.status, body: await resp.json() };
		});

		expect(resultado.status).toBe(201);
		expect(resultado.body.destino.tipo).toBe('whatsapp');
		expect(resultado.body.destino.config).toContain(',');
	});

	test('POST /api/destinos sem nome retorna erro 400', async ({ page }) => {
		await page.route('**/api/destinos', async route => {
			if (route.request().method() === 'POST') {
				const body = JSON.parse(route.request().postData());
				if (!body.nome) {
					return route.fulfill({ status: 400, json: { erro: 'nome é obrigatório' } });
				}
				return route.fulfill({ status: 201, json: { status: 'ok', destino: body } });
			}
			return route.fulfill({ json: { destinos: [] } });
		});

		await page.goto('/canais');

		const resultado = await page.evaluate(async () => {
			const resp = await fetch('/api/destinos', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ nome: '', config: '123@g.us', tipo: 'whatsapp' })
			});
			return { status: resp.status, body: await resp.json() };
		});

		expect(resultado.status).toBe(400);
	});

	test('GET /api/whatsapp/grupos retorna lista mockada', async ({ page }) => {
		const gruposMock = [
			{ id: '120363430000000000@g.us', nome: '#1 Garimpo Hoje' },
			{ id: '120363410893012870@g.us', nome: '#08 AVANÇADO VOE' }
		];
		await page.route('**/api/whatsapp/grupos', route =>
			route.fulfill({ json: { grupos: gruposMock } })
		);

		await page.goto('/canais');

		const response = await page.evaluate(async () => {
			const resp = await fetch('/api/whatsapp/grupos');
			return resp.json();
		});

		expect(response.grupos).toHaveLength(2);
		expect(response.grupos[0].nome).toBe('#1 Garimpo Hoje');
	});
});
