/**
 * Tests for omnibox-intencao — intent detection from free text.
 */
import { describe, it, expect } from 'vitest';
import { detectarIntencao, isUrl } from '$lib/omnibox-intencao.js';

describe('isUrl', () => {
	it('detects https URLs', () => {
		expect(isUrl('https://shopee.com.br/shop/12345')).toBe(true);
	});

	it('detects affiliate short links', () => {
		expect(isUrl('https://s.shopee.com.br/8fQYnxWQqu')).toBe(true);
		expect(isUrl('https://shp.ee/abc123')).toBe(true);
	});

	it('detects http URLs', () => {
		expect(isUrl('http://shopee.com.br/shop/12345')).toBe(true);
	});

	it('does not detect plain text', () => {
		expect(isUrl('glory of seoul')).toBe(false);
		expect(isUrl('serum')).toBe(false);
	});

	it('does not detect partial URLs', () => {
		expect(isUrl('shopee.com.br')).toBe(false);
	});

	it('handles empty/null', () => {
		expect(isUrl('')).toBe(false);
		expect(isUrl(null)).toBe(false);
		expect(isUrl(undefined)).toBe(false);
	});
});

describe('detectarIntencao', () => {
	const ctxBase = {
		categoriasDisponiveis: [
			{ nome: 'Beleza', marketplaces: ['shopee'] },
			{ nome: 'Eletrônicos', marketplaces: ['shopee', 'amazon'] },
			{ nome: 'Moda', marketplaces: ['shopee', 'mercado_livre'] }
		],
		marketplacesFiltro: [],
		shopIds: [],
		shopNomes: {}
	};

	describe('URL detection', () => {
		it('returns resolver_link as sole option for URLs', () => {
			const result = detectarIntencao('https://shopee.com.br/shop/123', ctxBase);
			expect(result).toHaveLength(1);
			expect(result[0].tipo).toBe('resolver_link');
			expect(result[0].payload.url).toBe('https://shopee.com.br/shop/123');
			expect(result[0].icone).toBe('🔗');
		});

		it('returns resolver_link for http URLs', () => {
			const result = detectarIntencao('http://example.com', ctxBase);
			expect(result).toHaveLength(1);
			expect(result[0].tipo).toBe('resolver_link');
		});

		it('URL option has accessible label', () => {
			const result = detectarIntencao('https://shopee.com.br/shop/123', ctxBase);
			expect(result[0].labelAcessivel).toContain('Resolver link');
		});
	});

	describe('text < 2 chars', () => {
		it('returns empty for single char', () => {
			expect(detectarIntencao('a', ctxBase)).toHaveLength(0);
		});

		it('returns empty for empty string', () => {
			expect(detectarIntencao('', ctxBase)).toHaveLength(0);
		});

		it('returns empty for whitespace only', () => {
			expect(detectarIntencao('   ', ctxBase)).toHaveLength(0);
		});
	});

	describe('base options (produtos + lojas)', () => {
		it('returns produtos and lojas for text >= 2 chars', () => {
			const result = detectarIntencao('se', ctxBase);
			const tipos = result.map((o) => o.tipo);
			expect(tipos).toContain('produtos');
			expect(tipos).toContain('lojas');
		});

		it('produtos option includes keyword in payload', () => {
			const result = detectarIntencao('serum coreano', ctxBase);
			const prod = result.find((o) => o.tipo === 'produtos');
			expect(prod.payload.keyword).toBe('serum coreano');
		});

		it('lojas option includes termo in payload', () => {
			const result = detectarIntencao('glory', ctxBase);
			const lojas = result.find((o) => o.tipo === 'lojas');
			expect(lojas.payload.termo).toBe('glory');
		});

		it('labels include the search text', () => {
			const result = detectarIntencao('serum', ctxBase);
			const prod = result.find((o) => o.tipo === 'produtos');
			expect(prod.label).toContain('serum');
			expect(prod.labelAcessivel).toContain('serum');
		});
	});

	describe('category matching', () => {
		it('includes categoria option when text matches', () => {
			const result = detectarIntencao('beleza', ctxBase);
			const catOpts = result.filter((o) => o.tipo === 'categoria');
			expect(catOpts.length).toBeGreaterThanOrEqual(1);
			expect(catOpts[0].payload.categoria).toBe('Beleza');
		});

		it('matches partially (substring)', () => {
			const result = detectarIntencao('bel', ctxBase);
			const catOpts = result.filter((o) => o.tipo === 'categoria');
			expect(catOpts.length).toBe(1);
			expect(catOpts[0].payload.categoria).toBe('Beleza');
		});

		it('case-insensitive matching', () => {
			const result = detectarIntencao('BELEZA', ctxBase);
			const catOpts = result.filter((o) => o.tipo === 'categoria');
			expect(catOpts.length).toBe(1);
		});

		it('respects maxCategorias limit', () => {
			const ctx = {
				...ctxBase,
				categoriasDisponiveis: Array.from({ length: 10 }, (_, i) => ({
					nome: `Cat${i}`,
					marketplaces: ['shopee']
				}))
			};
			const result = detectarIntencao('cat', ctx);
			const catOpts = result.filter((o) => o.tipo === 'categoria');
			expect(catOpts.length).toBeLessThanOrEqual(3);
		});
	});

	describe('context: marketplace filter', () => {
		it('shows marketplace name when single marketplace active', () => {
			const ctx = { ...ctxBase, marketplacesFiltro: ['shopee'] };
			const result = detectarIntencao('beleza', ctx);
			const catOpt = result.find((o) => o.tipo === 'categoria');
			expect(catOpt.label).toContain('shopee');
		});

		it('generic label with multiple marketplaces active', () => {
			const ctx = { ...ctxBase, marketplacesFiltro: ['shopee', 'amazon'] };
			const result = detectarIntencao('eletr', ctx);
			const catOpt = result.find((o) => o.tipo === 'categoria');
			// Multiple marketplaces → generic label
			expect(catOpt.label).toContain('#Eletrônicos');
		});
	});

	describe('context: lojas no escopo', () => {
		it('shows loja name when single loja active', () => {
			const ctx = {
				...ctxBase,
				shopIds: [123],
				shopNomes: { 123: 'Glory of Seoul' }
			};
			const result = detectarIntencao('beleza', ctx);
			const catOpt = result.find((o) => o.tipo === 'categoria');
			expect(catOpt.label).toContain('Glory of Seoul');
		});

		it('shows generic "nas lojas selecionadas" for multiple lojas', () => {
			const ctx = {
				...ctxBase,
				shopIds: [123, 456],
				shopNomes: { 123: 'Glory', 456: 'Le Botanic' }
			};
			const result = detectarIntencao('beleza', ctx);
			const catOpt = result.find((o) => o.tipo === 'categoria');
			expect(catOpt.label).toContain('lojas selecionadas');
		});

		it('produtos option includes context suffix for single loja', () => {
			const ctx = {
				...ctxBase,
				shopIds: [123],
				shopNomes: { 123: 'Glory of Seoul' }
			};
			const result = detectarIntencao('serum', ctx);
			const prod = result.find((o) => o.tipo === 'produtos');
			expect(prod.label).toContain('Glory of Seoul');
		});
	});

	describe('order', () => {
		it('follows configured order: produtos, lojas, categorias', () => {
			const result = detectarIntencao('beleza', ctxBase);
			const tipos = result.map((o) => o.tipo);
			const idxProdutos = tipos.indexOf('produtos');
			const idxLojas = tipos.indexOf('lojas');
			const idxCat = tipos.indexOf('categoria');
			expect(idxProdutos).toBeLessThan(idxLojas);
			expect(idxLojas).toBeLessThan(idxCat);
		});
	});

	describe('no ctx provided', () => {
		it('works with empty context', () => {
			const result = detectarIntencao('serum');
			expect(result.length).toBeGreaterThanOrEqual(2);
			expect(result[0].tipo).toBe('produtos');
			expect(result[1].tipo).toBe('lojas');
		});
	});
});
