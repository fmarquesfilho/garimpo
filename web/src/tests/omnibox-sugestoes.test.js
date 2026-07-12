import { describe, it, expect } from 'vitest';
import { gerarSugestoes } from '$lib/omnibox-sugestoes.js';
import { parsearInput } from '$lib/omnibox-parser.js';

const CONFIG = { minChars: 2, maxSugestoes: 7, matchBuscaSalva: true };

const ctx = {
	lojasMonitoradas: [
		{ id: '920', nome: 'Glory of Seoul', marketplace: 'shopee' },
		{ id: '281', nome: 'Le Botanic', marketplace: 'shopee' }
	],
	categoriasDisponiveis: [{ nome: 'Beleza' }, { nome: 'Casa' }, { nome: 'Bebês' }],
	marketplaces: ['shopee', 'mercado_livre', 'amazon'],
	buscasSalvas: [{ keywords: ['serum vitamina c'], shopNomes: { 920: 'Glory of Seoul' } }]
};

/**
 * Helper: gera sugestões para o último token de um raw input.
 * @param {string} raw
 * @param {any} [c]
 * @param {any} [cfg]
 */
function sugerir(raw, c = ctx, cfg = CONFIG) {
	const tokens = parsearInput(raw);
	return gerarSugestoes(tokens[tokens.length - 1] ?? { tipo: 'keyword', valor: '', completo: false }, c, cfg);
}

describe('omnibox-sugestoes — match sem prefixo (inferência)', () => {
	it('match loja parcial (case-insensitive)', () => {
		const m = sugerir('glo');
		expect(m.get('loja')?.map((s) => s.label)).toEqual(['Glory of Seoul']);
	});

	it('match categoria', () => {
		const m = sugerir('bel');
		expect(m.get('categoria')?.map((s) => s.label)).toContain('Beleza');
	});

	it('match marketplace (startsWith)', () => {
		const m = sugerir('sho');
		expect(m.get('marketplace')?.map((s) => s.label)).toEqual(['shopee']);
	});

	it('keyword sem prefixo busca em todos os tipos', () => {
		const m = sugerir('se'); // "se" casa Seoul (loja) e serum (busca salva)
		expect([...m.keys()]).toEqual(expect.arrayContaining(['loja', 'busca_salva']));
	});
});

describe('omnibox-sugestoes — prefixos filtram por tipo', () => {
	it('@ filtra só lojas', () => {
		const m = sugerir('@glo');
		expect([...m.keys()]).toEqual(['loja']);
	});

	it('loja retorna sugestão para nomes normalizados (ex: @gloryofseoul -> Glory of Seoul)', () => {
		const ctx = {
			lojasMonitoradas: [
				{ id: 1, nome: 'Glory of Seoul', nome_normalizado: 'gloryofseoul' },
				{ id: 2, nome: 'Le Botanic', nome_normalizado: 'lebotanic' }
			]
		};

		/** @type {import('$lib/omnibox-parser.js').Token} */
		const t1 = { tipo: 'loja', valor: '@gloryofseoul', completo: false };
		// note que matchLojas no registry ignora '@', the query sent is 'gloryofseoul' 
		// wait, no! The query sent to gerarSugestoes is `valor.toLowerCase()`.
		// so the query is '@gloryofseoul'. normalizarNome('@gloryofseoul') becomes 'gloryofseoul'
		const r1 = gerarSugestoes(t1, ctx);
		expect(r1.get('loja')[0].label).toBe('Glory of Seoul');

		/** @type {import('$lib/omnibox-parser.js').Token} */
		const t2 = { tipo: 'loja', valor: '@le', completo: false };
		const r2 = gerarSugestoes(t2, ctx);
		expect(r2.get('loja')[0].label).toBe('Le Botanic');
	});

	it('# filtra só categorias', () => {
		const m = sugerir('#bel');
		expect([...m.keys()]).toEqual(['categoria']);
	});

	it('! filtra só marketplaces', () => {
		const m = sugerir('!sho');
		expect([...m.keys()]).toEqual(['marketplace']);
	});
});

describe('omnibox-sugestoes — regras de config', () => {
	it('< minChars → Map vazio', () => {
		expect(sugerir('s').size).toBe(0);
		expect(sugerir('@g').size).toBeGreaterThanOrEqual(0); // '@g' → valor 'g' (1 char) → vazio
		expect(sugerir('@g').size).toBe(0);
	});

	it('respeita maxSugestoes por grupo', () => {
		const muitas = { categoriasDisponiveis: Array.from({ length: 10 }, (_, i) => ({ nome: `Cat${i}` })) };
		const m = sugerir('cat', muitas, { minChars: 2, maxSugestoes: 7 });
		expect(m.get('categoria').length).toBe(7);
	});

	it('buscas salvas aparecem primeiro (ordem do Map)', () => {
		const m = sugerir('serum'); // casa busca salva "serum vitamina c"
		expect([...m.keys()][0]).toBe('busca_salva');
	});

	it('matchBuscaSalva=false não retorna buscas salvas', () => {
		const m = sugerir('serum', ctx, { minChars: 2, maxSugestoes: 7, matchBuscaSalva: false });
		expect(m.has('busca_salva')).toBe(false);
	});
});

describe('omnibox-sugestoes — degradação graceful (Req 8.4)', () => {
	it('sem lojas nem categorias → só keyword global (Map vazio de sugestões locais)', () => {
		const m = sugerir('serum', { marketplaces: [], lojasMonitoradas: [], categoriasDisponiveis: [] });
		expect(m.size).toBe(0);
	});

	it('categorias ausentes não quebram match de loja', () => {
		const m = sugerir('glo', { lojasMonitoradas: ctx.lojasMonitoradas });
		expect(m.get('loja')?.length).toBe(1);
	});
});
