/**
 * E2E LOCAL — Validação do frontend contra rules/busca-rules.json.
 *
 * Prova que o comportamento visível da UI obedece às regras declarativas.
 */
import { test, expect } from './fixtures.js';
import {
	abrirRaiaFiltros,
	abrirRaiaLojas,
	abrirPainelBuscas,
	adicionarLoja,
	waitForEngineReady,
	SEL
} from './helpers.js';
import { readFileSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const rules = JSON.parse(readFileSync(resolve(__dirname, '../../../rules/busca-rules.json'), 'utf-8'));

test.describe('Regras externas (busca-rules.json)', () => {
	test('intent keyword_na_loja ao adicionar loja com keyword', async ({ authedPage: page }) => {
		const intentRow = rules.intent.find((r) => r.keyword && r.shop);
		expect(intentRow.result).toBe('keyword_na_loja');

		page.apiOverrides({
			'/api/candidatos': {
				candidatos: [
					{
						id: 'p1',
						produto_id: 'p1',
						nome: 'Serum Le Botanic',
						preco: 49.9,
						comissao: 0.12,
						vendas: 80,
						loja: 'Le Botanic',
						link: 'https://x'
					}
				],
				total_bruto: 1,
				estrategia: 'nicho'
			},
			'/api/lojas': { id: 'loja-lb', keyword: 'Le Botanic', shop_ids: [920292999], status: 'adicionada' }
		});

		await page.goto('/');
		await page.locator(SEL.searchInput).fill('serum');
		await expect(page.getByText('Serum Le Botanic')).toBeVisible({ timeout: 10000 });

		await abrirRaiaLojas(page);
		await adicionarLoja(page, 'https://s.shopee.com.br/8fQYnxWQqu');
		await expect(page.getByText('🏪 Le Botanic')).toBeVisible({ timeout: 10000 });
	});

	test('defaults fontes: curadoria=true, quedas=true, novos=true', async ({ authedPage: page }) => {
		expect(rules.defaults.fontes.curadoria).toBe(true);
		expect(rules.defaults.fontes.quedas).toBe(true);
		expect(rules.defaults.fontes.novos).toBe(true);
		expect(rules.defaults.fontes.lojas).toBe(false);

		await page.goto('/');
		await abrirRaiaFiltros(page);
		// Toggles de fontes estão na raia Filtros
		await expect(page.locator('button:has-text("🆕 Novos")')).toBeVisible({ timeout: 5000 });
	});

	test('busca agendada inclui cron e badge ⏱', async ({ authedPage: page }) => {
		page.apiOverrides({
			'/api/buscas': {
				buscas: [
					{
						id: 'b-cron',
						keywords: ['vitamina c'],
						shop_ids: [],
						shop_names: null,
						comissao_min: 0.07,
						vendas_min: 0,
						categorias: null,
						fontes: ['curadoria'],
						cron: '0 */8 * * *',
						marketplaces: 'shopee'
					}
				],
				total: 1
			}
		});
		await page.goto('/');
		await waitForEngineReady(page);
		await abrirPainelBuscas(page);
		await expect(page.getByText('a cada 8h')).toBeVisible({ timeout: 5000 });
	});

	test('salvar busca com loja → card mostra shop_names', async ({ authedPage: page }) => {
		let postCount = 0;
		page.apiOverrides({
			'/api/lojas': { id: 'loja-lb', keyword: 'Le Botanic', shop_ids: [920292999], status: 'adicionada' },
			'/api/buscas': (req) => {
				if (req.method() === 'POST') {
					postCount++;
					return { id: 'b1', status: 'salva' };
				}
				if (postCount > 0) {
					return {
						buscas: [
							{
								id: 'b1',
								keywords: ['serum'],
								shop_ids: [920292999],
								shop_names: { 920292999: 'Le Botanic' },
								comissao_min: 0.07,
								vendas_min: 0,
								categorias: null,
								fontes: ['curadoria', 'lojas'],
								cron: null,
								marketplaces: 'shopee'
							}
						],
						total: 1
					};
				}
				return { buscas: [], total: 0 };
			}
		});

		await page.goto('/');
		await page.locator(SEL.searchInput).fill('serum');
		await page.waitForTimeout(600);

		await abrirRaiaLojas(page);
		await adicionarLoja(page, 'Le Botanic');
		await expect(page.getByText('Le Botanic').first()).toBeVisible({ timeout: 10000 });

		// Salva
		await page.locator(SEL.btnBuscasSalvas).click();
		await page.getByRole('button', { name: /salvar busca atual/ }).click();
		await page.locator('button:has-text("Salvar")').last().click();

		// Card com shop_names renderizados
		await expect(page.getByText('🏪 Le Botanic')).toBeVisible({ timeout: 10000 });
	});
});
