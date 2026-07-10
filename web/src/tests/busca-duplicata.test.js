import { describe, it, expect } from 'vitest';
import { fingerprint, buscarDuplicada, buscaSalvaToCtx } from '$lib/busca-config.js';

/**
 * Testes isolados para fingerprint() e buscarDuplicada().
 * Funções puras — testáveis sem engine, sem DOM, sem API.
 */

describe('fingerprint — normalização de contexto', () => {
	it('keyword normaliza para lowercase trimmed', () => {
		const fp1 = fingerprint({ keyword: 'Serum', shopIds: [], categorias: [], marketplacesFiltro: [] });
		const fp2 = fingerprint({ keyword: '  serum  ', shopIds: [], categorias: [], marketplacesFiltro: [] });
		expect(fp1).toBe(fp2);
	});

	it('shopIds são sorted (ordem não importa)', () => {
		const fp1 = fingerprint({ keyword: '', shopIds: [3, 1, 2], categorias: [], marketplacesFiltro: [] });
		const fp2 = fingerprint({ keyword: '', shopIds: [1, 2, 3], categorias: [], marketplacesFiltro: [] });
		expect(fp1).toBe(fp2);
	});

	it('categorias são sorted e lowercase', () => {
		const fp1 = fingerprint({
			keyword: '',
			shopIds: [],
			categorias: ['Perfumaria', 'Beleza'],
			marketplacesFiltro: []
		});
		const fp2 = fingerprint({
			keyword: '',
			shopIds: [],
			categorias: ['beleza', 'perfumaria'],
			marketplacesFiltro: []
		});
		expect(fp1).toBe(fp2);
	});

	it('marketplacesFiltro são sorted', () => {
		const fp1 = fingerprint({
			keyword: '',
			shopIds: [],
			categorias: [],
			marketplacesFiltro: ['amazon', 'shopee']
		});
		const fp2 = fingerprint({
			keyword: '',
			shopIds: [],
			categorias: [],
			marketplacesFiltro: ['shopee', 'amazon']
		});
		expect(fp1).toBe(fp2);
	});

	it('contextos diferentes produzem fingerprints diferentes', () => {
		const fp1 = fingerprint({ keyword: 'serum', shopIds: [], categorias: [], marketplacesFiltro: [] });
		const fp2 = fingerprint({ keyword: 'retinol', shopIds: [], categorias: [], marketplacesFiltro: [] });
		expect(fp1).not.toBe(fp2);
	});

	it('campos null/undefined são tratados como vazios', () => {
		const fp1 = fingerprint({ keyword: '', shopIds: [], categorias: [], marketplacesFiltro: [] });
		const fp2 = fingerprint({ keyword: null, shopIds: undefined, categorias: null, marketplacesFiltro: undefined });
		expect(fp1).toBe(fp2);
	});

	it('keyword vazia com shopIds é diferente de keyword com shopIds vazio', () => {
		const fp1 = fingerprint({ keyword: '', shopIds: [123], categorias: [], marketplacesFiltro: [] });
		const fp2 = fingerprint({ keyword: 'serum', shopIds: [], categorias: [], marketplacesFiltro: [] });
		expect(fp1).not.toBe(fp2);
	});
});

describe('buscaSalvaToCtx — conversão para formato comparável', () => {
	it('extrai keyword do primeiro item de keywords[]', () => {
		const ctx = buscaSalvaToCtx({ keywords: ['serum', 'vitamina c'], shopIds: [] });
		expect(ctx.keyword).toBe('serum');
	});

	it('keywords vazio produz keyword vazia', () => {
		const ctx = buscaSalvaToCtx({ keywords: [], shopIds: [123] });
		expect(ctx.keyword).toBe('');
	});

	it('marketplaces string é ignorada (só arrays contam para filtro)', () => {
		const ctx = buscaSalvaToCtx({ keywords: [], shopIds: [], marketplaces: 'shopee' });
		expect(ctx.marketplacesFiltro).toEqual([]);
	});

	it('marketplaces array é preservado', () => {
		const ctx = buscaSalvaToCtx({ keywords: [], shopIds: [], marketplaces: ['shopee', 'amazon'] });
		expect(ctx.marketplacesFiltro).toEqual(['shopee', 'amazon']);
	});
});

describe('buscarDuplicada — detecção de busca existente', () => {
	const salvas = [
		{ id: 'b1', keywords: ['serum'], shopIds: [], categorias: [], marketplaces: 'shopee' },
		{ id: 'b2', keywords: ['retinol'], shopIds: [123], categorias: ['Beleza'], marketplaces: ['shopee'] },
		{ id: 'b3', keywords: [], shopIds: [456, 789], categorias: [], marketplaces: ['amazon', 'shopee'] }
	];

	it('encontra duplicata por keyword exata', () => {
		const ctx = { keyword: 'serum', shopIds: [], categorias: [], marketplacesFiltro: [] };
		const dup = buscarDuplicada(ctx, salvas);
		expect(dup).not.toBeNull();
		expect(dup.id).toBe('b1');
	});

	it('encontra duplicata com keyword case-insensitive', () => {
		const ctx = { keyword: 'SERUM', shopIds: [], categorias: [], marketplacesFiltro: [] };
		const dup = buscarDuplicada(ctx, salvas);
		expect(dup).not.toBeNull();
		expect(dup.id).toBe('b1');
	});

	it('encontra duplicata com shopIds em ordem diferente', () => {
		const ctx = { keyword: '', shopIds: [789, 456], categorias: [], marketplacesFiltro: ['shopee', 'amazon'] };
		const dup = buscarDuplicada(ctx, salvas);
		expect(dup).not.toBeNull();
		expect(dup.id).toBe('b3');
	});

	it('não encontra duplicata para configuração nova', () => {
		const ctx = { keyword: 'vitamina c', shopIds: [], categorias: [], marketplacesFiltro: [] };
		const dup = buscarDuplicada(ctx, salvas);
		expect(dup).toBeNull();
	});

	it('exclui a busca em edição (excluirId)', () => {
		const ctx = { keyword: 'serum', shopIds: [], categorias: [], marketplacesFiltro: [] };
		const dup = buscarDuplicada(ctx, salvas, 'b1');
		expect(dup).toBeNull();
	});

	it('sem buscas salvas retorna null', () => {
		const ctx = { keyword: 'serum', shopIds: [], categorias: [], marketplacesFiltro: [] };
		expect(buscarDuplicada(ctx, [], null)).toBeNull();
		expect(buscarDuplicada(ctx, null, null)).toBeNull();
	});

	it('categorias case-insensitive e sorted', () => {
		const ctx = { keyword: 'retinol', shopIds: [123], categorias: ['beleza'], marketplacesFiltro: ['shopee'] };
		const dup = buscarDuplicada(ctx, salvas);
		expect(dup).not.toBeNull();
		expect(dup.id).toBe('b2');
	});
});
