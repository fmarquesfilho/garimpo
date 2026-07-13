/**
 * E2E LOCAL — Validacao do frontend contra rules/busca-rules.json.
 * Prova que o comportamento visivel da UI obedece as regras declarativas.
 */
import { test, expect } from './fixtures.js';
import { adicionarLojaViaOmnibox, abrirPainelBuscas, waitForEngineReady, SEL } from './helpers.js';
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
			'/api/lojas/resolver': {
				id: '920292999',
				nome: 'Le Botanic',
				nome_normalizado: 'lebotanic',
				marketplace: 'shopee',
				monitorada: false,
				imagem: null,
				seguidores: null,
				total_produtos: null,
				avaliacao: null
			}
		});

		await page.goto('/');
		const input = page.getByRole('combobox');
		await input.fill('serum');
		await input.press('Enter');
		await expect(page.getByText('Serum Le Botanic')).toBeVisible({ timeout: 10000 });

		// Adiciona loja via Omnibox (resolve link)
		await adicionarLojaViaOmnibox(page, 'https://s.shopee.com.br/8fQYnxWQqu');
		await expect(page.getByLabel(/Loja:.*Le Botanic/)).toBeVisible({ timeout: 10000 });
	});

	test('defaults fontes: curadoria=true, quedas=true, novos=true', async ({ authedPage: page }) => {
		expect(rules.defaults.fontes.curadoria).toBe(true);
		expect(rules.defaults.fontes.quedas).toBe(true);
		expect(rules.defaults.fontes.novos).toBe(true);
		expect(rules.defaults.fontes.lojas).toBe(false);

		await page.goto('/');
		// Toggles de fontes sao inline (sempre visiveis)
		await expect(page.getByRole('button', { name: /Novos/ })).toBeVisible({ timeout: 5000 });
		await expect(page.getByRole('button', { name: /Quedas/ })).toBeVisible();
	});

	test('busca agendada inclui cron e badge', async ({ authedPage: page }) => {
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

	test('omnibox intencao config aplicada', async ({ authedPage: page }) => {
		expect(rules.omnibox.intencao.habilitado).toBe(true);
		expect(rules.omnibox.intencao.minChars).toBe(2);

		await page.goto('/');
		const input = page.getByRole('combobox');

		// < minChars: sem dropdown
		await input.fill('s');
		await expect(page.getByRole('listbox')).not.toBeVisible();

		// >= minChars: dropdown com opcoes
		await input.fill('se');
		await expect(page.getByRole('listbox')).toBeVisible({ timeout: 5000 });
	});
});
