/**
 * Contract test: valida que payloadToConfig transforma o formato da API
 * (api-buscas.json) no formato esperado pelo frontend (frontend-ctx.json).
 *
 * Se este teste quebra, significa que o contrato API ↔ Frontend mudou.
 * Atualize os fixtures em fixtures/respostas/ para refletir a mudança.
 */
import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { resolve } from 'path';
import { payloadToConfig } from '$lib/busca-unificada-logic.js';

const fixturesDir = resolve(__dirname, '../../../fixtures/respostas');

const apiBuscas = JSON.parse(readFileSync(resolve(fixturesDir, 'api-buscas.json'), 'utf-8'));
const frontendCtx = JSON.parse(readFileSync(resolve(fixturesDir, 'frontend-ctx.json'), 'utf-8'));

describe('fixtures-contract: payloadToConfig vs golden', () => {
	it('deve transformar cada busca da API no config esperado pelo frontend', () => {
		const configs = apiBuscas.buscas.map(payloadToConfig);

		expect(configs.length).toBe(frontendCtx.configs.length);

		for (let i = 0; i < configs.length; i++) {
			const actual = configs[i];
			const expected = frontendCtx.configs[i];

			expect(actual.id).toBe(expected.id);
			expect(actual.keywords).toEqual(expected.keywords);
			expect(actual.shopIds).toEqual(expected.shopIds);
			expect(actual.shopNomes).toEqual(expected.shopNomes);
			expect(actual.comissaoMin).toBe(expected.comissaoMin);
			expect(actual.vendasMin).toBe(expected.vendasMin);
			expect(actual.categorias).toEqual(expected.categorias);
			expect(actual.fontes).toEqual(expected.fontes);
			expect(actual.cron).toBe(expected.cron);
			expect(actual.marketplaces).toBe(expected.marketplaces);
		}
	});

	it('busca com shop_names deve mapear todos os IDs para nomes (não apenas o primeiro)', () => {
		const multishop = apiBuscas.buscas.find((b) => b.id === 'busca-loja-multi');
		expect(multishop).toBeDefined();

		const config = payloadToConfig(multishop);
		// Deve ter 2 lojas mapeadas, não apenas a primeira
		expect(Object.keys(config.shopNomes).length).toBe(2);
		expect(config.shopNomes['282170857']).toBe('Le Botanic');
		expect(config.shopNomes['592884015']).toBe('COSRX Official');
	});

	it('busca keyword-only não deve ter shopNomes', () => {
		const kwOnly = apiBuscas.buscas.find((b) => b.id === 'busca-keyword-serum');
		const config = payloadToConfig(kwOnly);
		expect(config.shopNomes).toEqual({});
		expect(config.shopIds).toEqual([]);
	});

	it('busca sem keyword com loja deve preservar shopNomes e shopIds', () => {
		const lojaOnly = apiBuscas.buscas.find((b) => b.id === 'busca-loja-glory');
		const config = payloadToConfig(lojaOnly);
		expect(config.shopIds).toEqual([920292999]);
		expect(config.shopNomes).toEqual({ 920292999: 'Glory of Seoul' });
		expect(config.keywords).toEqual([]);
	});
});
