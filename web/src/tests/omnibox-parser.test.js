import { describe, it, expect } from 'vitest';
import { parsearInput, serializarTokens, tokensParaContexto } from '$lib/omnibox-parser.js';

describe('omnibox-parser — parsearInput', () => {
	it('keyword simples (incompleta)', () => {
		expect(parsearInput('serum')).toEqual([{ tipo: 'keyword', valor: 'serum', completo: false }]);
	});

	it('multi-keyword com espaço final → ambas completas', () => {
		expect(parsearInput('serum vitamina ')).toEqual([
			{ tipo: 'keyword', valor: 'serum', completo: true },
			{ tipo: 'keyword', valor: 'vitamina', completo: true }
		]);
	});

	it('multi-keyword sem espaço final → última incompleta', () => {
		const t = parsearInput('serum vitamina');
		expect(t[0].completo).toBe(true);
		expect(t[1].completo).toBe(false);
	});

	it('prefixo loja @', () => {
		expect(parsearInput('@glory')).toEqual([{ tipo: 'loja', valor: 'glory', completo: false }]);
	});

	it('prefixo categoria #', () => {
		expect(parsearInput('#beleza')).toEqual([{ tipo: 'categoria', valor: 'beleza', completo: false }]);
	});

	it('prefixo marketplace !', () => {
		expect(parsearInput('!shopee')).toEqual([{ tipo: 'marketplace', valor: 'shopee', completo: false }]);
	});

	it('misto: keyword + loja + categoria', () => {
		const t = parsearInput('serum @lebotanic #beleza');
		expect(t.map((x) => x.tipo)).toEqual(['keyword', 'loja', 'categoria']);
		expect(t.map((x) => x.valor)).toEqual(['serum', 'lebotanic', 'beleza']);
		expect(t.map((x) => x.completo)).toEqual([true, true, false]);
	});

	it('prefixo inválido ($) é tratado como keyword', () => {
		expect(parsearInput('$texto')).toEqual([{ tipo: 'keyword', valor: '$texto', completo: false }]);
	});

	it('prefixo sozinho (@) → token de loja com valor vazio', () => {
		expect(parsearInput('@')).toEqual([{ tipo: 'loja', valor: '', completo: false }]);
	});

	it('input vazio → []', () => {
		expect(parsearInput('')).toEqual([]);
		expect(parsearInput('   ')).toEqual([]);
	});

	it('ignora espaços internos extras e leading', () => {
		expect(parsearInput('  serum   @glory ')).toEqual([
			{ tipo: 'keyword', valor: 'serum', completo: true },
			{ tipo: 'loja', valor: 'glory', completo: true }
		]);
	});

	it('não-string → []', () => {
		expect(parsearInput(null)).toEqual([]);
		expect(parsearInput(undefined)).toEqual([]);
	});
});

describe('omnibox-parser — serializarTokens (round-trip)', () => {
	const casos = ['serum', 'serum vitamina ', '@glory', '#beleza', 'serum @lebotanic #beleza', '@lebotanic !shopee', ''];
	for (const raw of casos) {
		it(`round-trip: ${JSON.stringify(raw)}`, () => {
			const t = parsearInput(raw);
			expect(parsearInput(serializarTokens(t))).toEqual(t);
		});
	}

	it('serializa tokens vazio → string vazia', () => {
		expect(serializarTokens([])).toBe('');
		expect(serializarTokens(null)).toBe('');
	});

	it('token completo preserva espaço final', () => {
		expect(serializarTokens([{ tipo: 'keyword', valor: 'serum', completo: true }])).toBe('serum ');
	});
});

describe('omnibox-parser — tokensParaContexto', () => {
	const ctx = {
		lojasMonitoradas: [
			{ id: '920', nome: 'Le Botanic', marketplace: 'shopee' },
			{ id: '281', nome: 'Glory of Seoul', marketplace: 'shopee' }
		],
		categoriasDisponiveis: [{ nome: 'Beleza' }, { nome: 'Casa' }],
		marketplaces: ['shopee', 'mercado_livre', 'amazon']
	};

	it('resolve keyword + loja + categoria + marketplace (match parcial de token)', () => {
		// token não contém espaço → resolve por match parcial ("botanic" ⊂ "Le Botanic")
		const r = tokensParaContexto(parsearInput('serum @botanic #beleza !shopee'), ctx);
		expect(r.keyword).toBe('serum');
		expect(r.shopIds).toEqual(['920']);
		expect(r.categorias).toEqual(['Beleza']);
		expect(r.marketplacesFiltro).toEqual(['shopee']);
		expect(r.lojasResolvidas[0].id).toBe('920');
	});

	it('concatena múltiplas keywords', () => {
		const r = tokensParaContexto(parsearInput('serum vitamina c'), ctx);
		expect(r.keyword).toBe('serum vitamina c');
	});

	it('ignora loja sem match no contexto', () => {
		const r = tokensParaContexto(parsearInput('@inexistente'), ctx);
		expect(r.shopIds).toEqual([]);
	});

	it('estado zero (ctx vazio) → só keyword', () => {
		const r = tokensParaContexto(parsearInput('serum @glory #beleza'), {});
		expect(r.keyword).toBe('serum');
		expect(r.shopIds).toEqual([]);
		expect(r.categorias).toEqual([]);
	});

	it('não duplica shopIds repetidos', () => {
		const r = tokensParaContexto(parsearInput('@botanic @botanic'), ctx);
		expect(r.shopIds).toEqual(['920']);
	});
});
