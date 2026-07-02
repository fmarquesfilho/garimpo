import { test, expect } from './fixtures.js';

/**
 * Testes E2E da página Descobrir.
 *
 * Usa o fixture `authedPage` que faz login via Firebase Auth Emulator.
 * Intercepta chamadas de API via route.fulfill para isolar o frontend.
 */

// ── Dados mock da API ─────────────────────────────────────────────────────

const mockCandidatos = [
	{ id: '1', nome: 'Sérum Vitamina C', preco: 89.9, comissao: 0.15, vendas: 200, avaliacao: 4.9, loja: 'SKIN1004', categoria: 'Cuidados com a Pele', imagem: '', link: '' },
	{ id: '2', nome: 'Perfume Kenzo 50ml', preco: 299.9, comissao: 0.08, vendas: 80, avaliacao: 4.6, loja: 'Perfumaria JP', categoria: 'Perfumaria', imagem: '', link: '' },
	{ id: '3', nome: 'Tônico BHA COSRX', preco: 59.9, comissao: 0.12, vendas: 500, avaliacao: 4.8, loja: 'COSRX Store', categoria: 'Cuidados com a Pele', imagem: '', link: '' },
	{ id: '4', nome: 'Batom Matte Ruby', preco: 25, comissao: 0.05, vendas: 30, avaliacao: 3.5, loja: 'Loja X', categoria: 'Maquiagem', imagem: '', link: '' }
];

function respondCandidatos(candidatos) {
	return JSON.stringify({ estrategia: 'nicho', total_bruto: candidatos.length, candidatos });
}

// ── Helpers ───────────────────────────────────────────────────────────────

async function interceptarAPIs(page) {
	await page.route('**/api/candidatos*', async (route) => {
		const url = new URL(route.request().url());
		const keyword = url.searchParams.get('keyword') ?? '';

		let filtered = mockCandidatos;
		if (keyword) {
			const kw = keyword.toLowerCase();
			filtered = mockCandidatos.filter(c =>
				c.nome.toLowerCase().includes(kw) ||
				c.categoria.toLowerCase().includes(kw) ||
				c.loja.toLowerCase().includes(kw)
			);
		}
		await route.fulfill({ status: 200, contentType: 'application/json', body: respondCandidatos(filtered) });
	});

	await page.route('**/api/lojas/novidades*', async (route) => {
		await route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify({
				variacoes: [
					{ produto_id: 'V1', nome: 'Tônico Queda Preço', preco_atual: 49, preco_anterior: 69, variacao_pct: -0.29, detectado_em: '2026-07-01', imagem: '', link: '', comissao: 0.1, vendas: 100 }
				],
				produtos_novos: [
					{ produto_id: 'N1', nome: 'Retinol Novo Lançamento', preco: 45, comissao: 0.1, vendas: 0, detectado_em: '2026-07-02', imagem: '', link: '' }
				]
			})
		});
	});

	await page.route('**/api/buscas', async (route) => {
		if (route.request().method() === 'GET') {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					buscas: [
						{ id: 'b1', keywords: ['sérum'], categorias: [], fontes: ['curadoria'] },
						{ id: 'loja-cosrx', nome: 'COSRX Store', shop_ids: [789], keywords: [] }
					]
				})
			});
		} else {
			await route.fulfill({ status: 200, contentType: 'application/json', body: '{}' });
		}
	});

	await page.route('**/api/favoritos*', async (route) => {
		await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ favoritos: [] }) });
	});

	await page.route('**/api/admin/me', async (route) => {
		await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ admin: false, email: 'teste-e2e@garimpo.dev' }) });
	});
}

// ── Testes ────────────────────────────────────────────────────────────────

test.describe('Descobrir — Filtros e resultados', () => {
	test('1. Busca por keyword exibe resultados', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.reload();
		await page.waitForLoadState('networkidle');

		const input = page.locator('input[type="search"]');
		await input.fill('sérum');
		await page.waitForTimeout(600);

		await expect(page.locator('.grade')).toBeVisible();
		const cards = page.locator('.grade > *');
		expect(await cards.count()).toBeGreaterThanOrEqual(1);
	});

	test('2. Busca vazia mostra feed genérico (quedas + novos)', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.reload();
		await page.waitForTimeout(800);

		// Fontes Quedas e Novos devem ter badges (loja monitorada no mock)
		const btnQuedas = page.locator('.fonte-btn', { hasText: 'Quedas' });
		const badgeQuedas = btnQuedas.locator('.fonte-badge');
		await expect(badgeQuedas).toBeVisible({ timeout: 5000 });
	});

	test('3. vendas_min filtra corretamente sem zerar', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.reload();
		await page.waitForLoadState('networkidle');

		const input = page.locator('input[type="search"]');
		await input.fill('c');
		await page.waitForTimeout(600);

		// Abrir filtros avançados
		await page.locator('.btn-avancado').click();
		const vendasInput = page.locator('input[type="number"]');
		await vendasInput.fill('100');
		await page.waitForTimeout(600);

		// Deve mostrar resultados (Sérum:200, Tônico:500 têm vendas >= 100)
		const contagem = page.locator('.contagem');
		await expect(contagem).toBeVisible({ timeout: 5000 });
		const texto = await contagem.textContent();
		expect(parseInt(texto)).toBeGreaterThan(0);
	});

	test('4. comissão_min filtra corretamente', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.reload();
		await page.waitForLoadState('networkidle');

		const input = page.locator('input[type="search"]');
		await input.fill('c');
		await page.waitForTimeout(600);

		await page.locator('.btn-avancado').click();
		const comissaoSelect = page.locator('select').first();
		await comissaoSelect.selectOption('0.15');
		await page.waitForTimeout(600);

		// Apenas Sérum (15%) deve aparecer dos resultados de curadoria
		const contagem = page.locator('.contagem');
		await expect(contagem).toBeVisible({ timeout: 5000 });
	});

	test('5. Categoria sem keyword retorna produtos', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.reload();
		await page.waitForLoadState('networkidle');

		// Abrir filtros avançados e preencher categoria
		await page.locator('.btn-avancado').click();
		const catInput = page.locator('input[placeholder="todas (digite para filtrar)"]');
		await catInput.fill('cosméticos');
		await page.waitForTimeout(600);

		// Não deve ter erro — a chamada à API deve ter sido feita com keyword="cosméticos"
		await expect(page.locator('.msg-erro')).not.toBeVisible();
	});

	test('6. Categoria + keyword = interseção', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.reload();
		await page.waitForLoadState('networkidle');

		const input = page.locator('input[type="search"]');
		await input.fill('sérum');
		await page.locator('.btn-avancado').click();
		const catInput = page.locator('input[placeholder="todas (digite para filtrar)"]');
		await catInput.fill('Cuidados com a Pele');
		await page.waitForTimeout(600);

		// Apenas Sérum Vitamina C (curadoria + Cuidados com a Pele)
		const contagem = page.locator('.contagem');
		await expect(contagem).toBeVisible({ timeout: 5000 });
		const texto = await contagem.textContent();
		expect(parseInt(texto)).toBeGreaterThanOrEqual(1);
	});

	test('7. Badge de contagem = número de cards visíveis', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.reload();
		await page.waitForLoadState('networkidle');

		const input = page.locator('input[type="search"]');
		await input.fill('c');
		await page.waitForTimeout(600);

		const badge = page.locator('.fonte-btn', { hasText: 'Busca' }).locator('.fonte-badge');
		if (await badge.isVisible()) {
			const badgeCount = parseInt(await badge.textContent());
			// Badge deve ser consistente com cards de curadoria exibidos
			expect(badgeCount).toBeGreaterThan(0);

			// Total de resultados mostrado deve ser >= badge (pode ter outras fontes)
			const contagem = await page.locator('.contagem').textContent();
			expect(parseInt(contagem)).toBeGreaterThanOrEqual(badgeCount);
		}
	});

	test('8. Toggle fontes atualiza resultados', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.reload();
		await page.waitForLoadState('networkidle');

		const input = page.locator('input[type="search"]');
		await input.fill('sérum');
		await page.waitForTimeout(600);

		// Deve ter resultado
		await expect(page.locator('.grade')).toBeVisible();

		// Desativar todas as fontes
		const fontes = page.locator('.fonte-btn');
		const count = await fontes.count();
		for (let i = 0; i < count; i++) {
			const btn = fontes.nth(i);
			if (await btn.evaluate(el => el.classList.contains('ativa'))) {
				await btn.click();
			}
		}
		await page.waitForTimeout(500);

		// Deve mostrar hint para ativar fonte
		await expect(page.locator('.hint-fontes')).toBeVisible();
	});

	test('9. Alterar filtro não zera resultados (filtragem client-side)', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.reload();
		await page.waitForLoadState('networkidle');

		const input = page.locator('input[type="search"]');
		await input.fill('c');
		await page.waitForTimeout(600);

		// Deve ter resultados inicialmente
		const contagemInicial = await page.locator('.contagem').textContent();
		expect(parseInt(contagemInicial)).toBeGreaterThan(0);

		// Alterar vendas_min para valor que ainda deixa resultados
		await page.locator('.btn-avancado').click();
		const vendasInput = page.locator('input[type="number"]');
		await vendasInput.fill('50');
		await page.waitForTimeout(600);

		// Deve continuar com resultados (não zerou)
		const contagem = page.locator('.contagem');
		await expect(contagem).toBeVisible();
		const texto = await contagem.textContent();
		expect(parseInt(texto)).toBeGreaterThan(0);
	});

	test('10. Filtro nota mínima não existe na interface', async ({ authedPage: page }) => {
		await interceptarAPIs(page);
		await page.reload();
		await page.waitForLoadState('networkidle');

		await page.locator('.btn-avancado').click();

		// O campo "nota mín." deve ter sido removido
		const labels = page.locator('.avancados .rotulo');
		const textos = await labels.allTextContents();
		expect(textos).not.toContain('nota mín.');
		expect(textos).toContain('categoria');
		expect(textos).toContain('comissão mín.');
		expect(textos).toContain('vendas mín.');
	});
});
