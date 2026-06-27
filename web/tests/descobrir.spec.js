import { test, expect } from '@playwright/test';

// ── Mock de autenticação ─────────────────────────────────────────────────
async function mockAuth(page) {
	await page.addInitScript(() => {
		window.__MOCK_USER = { uid: 'test', email: 'test@test.com', nome: 'Test User' };
	});
}

// ── Mock das APIs ────────────────────────────────────────────────────────
const produtosCuradoria = {
	candidatos: [
		{ id: 'P1', nome: 'Sérum Vitamina C SKIN1004', preco: 89.9, comissao: 0.12, vendas: 150, avaliacao: 4.8, imagem: 'https://img.test/1.jpg', link: 'https://shopee.com.br/p1', loja: 'SKIN1004 Official', categoria: 'skincare', score: 0.85 },
		{ id: 'P2', nome: 'Perfume Kenzo 50ml', preco: 299.9, comissao: 0.08, vendas: 80, avaliacao: 4.6, imagem: 'https://img.test/2.jpg', link: 'https://shopee.com.br/p2', loja: 'Perfumaria JP', categoria: 'perfumaria', score: 0.72 }
	]
};

const novidadesLoja = {
	busca_id: 'loja-123', dias_janela: 7, total_atual: 5,
	variacoes: [
		{ produto_id: 'V1', nome: 'Tônico COSRX', preco_anterior: 79.9, preco_atual: 59.9, variacao_pct: -0.25, detectado_em: '2026-06-27T10:00:00Z', imagem: 'https://img.test/v1.jpg', link: 'https://shopee.com.br/v1', loja: 'COSRX Store' },
		{ produto_id: 'V2', nome: 'Skin1004 Centella', preco_anterior: 120, preco_atual: 95, variacao_pct: -0.21, detectado_em: '2026-06-27T08:00:00Z', imagem: '', link: '', loja: '' }
	],
	produtos_novos: [
		{ produto_id: 'N1', nome: 'Retinol Serum Novo', preco: 45.5, comissao: 0.15, vendas: 0, nota: 0, detectado_em: '2026-06-27T12:00:00Z', imagem: 'https://img.test/n1.jpg', link: 'https://shopee.com.br/n1', loja: 'SKIN1004 Official' }
	]
};

async function mockAPIs(page) {
	await page.route('**/api/admin/me', route =>
		route.fulfill({ json: { uid: 'test', email: 'test@test.com', admin: false } })
	);
	await page.route('**/api/buscas', route =>
		route.fulfill({ json: { buscas: [
			{ id: 'skincare', keywords: ['sérum', 'skin1004'], cron: '0 8 * * *', ativo: true, fontes: ['curadoria', 'quedas'], shop_ids: [123] }
		] } })
	);
	await page.route('**/api/candidatos*', route =>
		route.fulfill({ json: produtosCuradoria })
	);
	await page.route('**/api/lojas/novidades*', route =>
		route.fulfill({ json: novidadesLoja })
	);
	await page.route('**/api/favoritos', route => {
		if (route.request().method() === 'GET') {
			route.fulfill({ json: { favoritos: [] } });
		} else {
			route.fulfill({ json: { status: 'ok' } });
		}
	});
}

// ── Testes ───────────────────────────────────────────────────────────────

test.describe('Descobrir — Estrutura da página', () => {
	test('mostra título e fonte toggles quando logado', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await expect(page.locator('h1')).toContainText('O que publicar hoje?');
		await expect(page.locator('.fonte-btn')).toHaveCount(4);
	});

	test('mostra input de busca com botão limpar', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await expect(page.locator('.busca-input')).toBeVisible();
		// Sem texto, botão X não aparece
		await expect(page.locator('.btn-limpar')).not.toBeVisible();
	});

	test('sem erros JS', async ({ page }) => {
		const errors = [];
		page.on('pageerror', err => errors.push(err.message));
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.waitForTimeout(2000);
		expect(errors).toHaveLength(0);
	});
});

test.describe('Descobrir — Fontes de dados', () => {
	test('Busca por keyword retorna resultados da curadoria', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.fill('.busca-input', 'sérum');
		await page.waitForTimeout(600); // debounce
		await expect(page.locator('.grade')).toBeVisible();
		await expect(page.locator('text=Sérum Vitamina C SKIN1004')).toBeVisible();
	});

	test('Quedas mostra variações de preço das lojas', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.waitForTimeout(600);
		// Quedas está ativo por default — deve mostrar resultados
		await expect(page.locator('text=Tônico COSRX')).toBeVisible();
	});

	test('Novos mostra produtos novos detectados', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.waitForTimeout(600);
		await expect(page.locator('text=Retinol Serum Novo')).toBeVisible();
	});

	test('keyword filtra resultados de quedas por nome', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.fill('.busca-input', 'Skin1004');
		await page.waitForTimeout(600);
		// Skin1004 Centella deve aparecer (contém "Skin1004")
		await expect(page.locator('text=Skin1004 Centella')).toBeVisible();
		// Tônico COSRX não contém "Skin1004" — não deve aparecer
		await expect(page.locator('text=Tônico COSRX')).not.toBeVisible();
	});

	test('keyword filtra por nome de loja', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.fill('.busca-input', 'COSRX');
		await page.waitForTimeout(600);
		// Tônico COSRX tem loja "COSRX Store" — deve aparecer
		await expect(page.locator('text=Tônico COSRX')).toBeVisible();
	});

	test('desativar todas as fontes mostra hint', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		// Desativa todas as fontes (busca, quedas, novos já ativos por default)
		const btns = page.locator('.fonte-btn');
		for (let i = 0; i < 4; i++) {
			const btn = btns.nth(i);
			if (await btn.evaluate(el => el.classList.contains('ativa'))) {
				await btn.click();
			}
		}
		await page.waitForTimeout(500);
		await expect(page.locator('.hint-fontes')).toBeVisible();
	});

	test('Favoritos mostra mensagem vazia quando não tem favoritos', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		// Desativa tudo, ativa só Favoritos
		const btns = page.locator('.fonte-btn');
		for (let i = 0; i < 4; i++) {
			const btn = btns.nth(i);
			if (await btn.evaluate(el => el.classList.contains('ativa'))) {
				await btn.click();
			}
		}
		await btns.nth(3).click(); // Favoritos
		await page.waitForTimeout(600);
		// Sem favoritos, deve estar vazio
		await expect(page.locator('text=0 produtos')).toBeVisible();
	});
});

test.describe('Descobrir — Input de busca', () => {
	test('botão X aparece quando há texto', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.fill('.busca-input', 'teste');
		await expect(page.locator('.btn-limpar')).toBeVisible();
	});

	test('botão X limpa o campo', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.fill('.busca-input', 'teste');
		await page.click('.btn-limpar');
		await expect(page.locator('.busca-input')).toHaveValue('');
		await expect(page.locator('.btn-limpar')).not.toBeVisible();
	});

	test('ESC limpa o campo', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.fill('.busca-input', 'teste');
		await page.press('.busca-input', 'Escape');
		await expect(page.locator('.busca-input')).toHaveValue('');
	});
});

test.describe('Descobrir — Buscas salvas (atalhos)', () => {
	test('mostra pills de keywords das buscas salvas', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.waitForTimeout(500);
		await expect(page.locator('.kw-pill')).toHaveCount(2); // sérum, skin1004
	});

	test('busca agendada mostra ícone ⏱', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.waitForTimeout(500);
		await expect(page.locator('.atalho-icone[title="Busca agendada"]')).toBeVisible();
	});

	test('clicar pill aplica keyword no input', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.waitForTimeout(500);
		await page.click('.kw-pill >> text=sérum');
		await expect(page.locator('.busca-input')).toHaveValue('sérum');
	});
});

test.describe('Descobrir — Badges de contagem', () => {
	test('fonte Quedas mostra badge com número', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.waitForTimeout(1000);
		// 2 quedas no mock
		await expect(page.locator('.fonte-badge.queda')).toContainText('2');
	});

	test('fonte Novos mostra badge com número', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.waitForTimeout(1000);
		// 1 novo no mock
		await expect(page.locator('.fonte-badge.novo')).toContainText('1');
	});
});

test.describe('Descobrir — ProductCard', () => {
	test('card de curadoria mostra imagem e posição', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.fill('.busca-input', 'sérum');
		await page.waitForTimeout(600);
		// Card deve ter imagem
		await expect(page.locator('.thumb').first()).toBeVisible();
		// Deve mostrar posição #1
		await expect(page.locator('text=#1')).toBeVisible();
	});

	test('card mostra botão de favoritar (☆)', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.fill('.busca-input', 'sérum');
		await page.waitForTimeout(600);
		await expect(page.locator('text=☆').first()).toBeVisible();
	});

	test('card mostra botão publicar', async ({ page }) => {
		await mockAuth(page);
		await mockAPIs(page);
		await page.goto('/');
		await page.fill('.busca-input', 'sérum');
		await page.waitForTimeout(600);
		await expect(page.locator('text=📤 Publicar').first()).toBeVisible();
	});
});
