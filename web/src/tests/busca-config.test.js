import { describe, it, expect } from 'vitest';
import {
	DEFAULTS,
	normalizarComissao,
	normalizarVendas,
	checarGuard,
	intentBusca,
	comissaoPercentLabel
} from '$lib/busca-config.js';

describe('busca-config — normalização', () => {
	it('comissão > 1 é interpretada como porcentagem (7 → 0.07)', () => {
		expect(normalizarComissao(7)).toBe(0.07);
		expect(normalizarComissao(15)).toBe(0.15);
	});

	it('comissão decimal é mantida e arredondada a 4 casas', () => {
		expect(normalizarComissao(0.15)).toBe(0.15);
		expect(normalizarComissao(0.07000001)).toBe(0.07);
	});

	it('comissão é limitada a [0,1] e trata valores inválidos', () => {
		expect(normalizarComissao(2.5)).toBe(0.025); // 2.5 > 1 → /100
		expect(normalizarComissao(-3)).toBe(0);
		expect(normalizarComissao(NaN)).toBe(DEFAULTS.comissaoMin);
		expect(normalizarComissao('x')).toBe(DEFAULTS.comissaoMin);
	});

	it('vendas viram inteiro >= 0', () => {
		expect(normalizarVendas(50.7)).toBe(50);
		expect(normalizarVendas(-5)).toBe(0);
		expect(normalizarVendas('12')).toBe(12);
		expect(normalizarVendas(undefined)).toBe(0);
	});
});

describe('busca-config — guards declarativos', () => {
	it('temContextoBusca exige keyword OU loja', () => {
		expect(checarGuard('temContextoBusca', { keyword: 'serum', shopIds: [] })).toBe(true);
		expect(checarGuard('temContextoBusca', { keyword: '', shopIds: [123] })).toBe(true);
		expect(checarGuard('temContextoBusca', { keyword: '  ', shopIds: [] })).toBe(false);
	});

	it('guard desconhecido não bloqueia', () => {
		expect(checarGuard('inexistente', {})).toBe(true);
	});
});

describe('busca-config — intent de busca (tabela de decisão)', () => {
	it('keyword + loja → escopa na loja', () => {
		expect(intentBusca({ keyword: 'serum', shopIds: [123] })).toBe('keyword_na_loja');
	});
	it('keyword sem loja → busca global', () => {
		expect(intentBusca({ keyword: 'serum', shopIds: [] })).toBe('keyword_global');
	});
	it('loja sem keyword → loja completa', () => {
		expect(intentBusca({ keyword: '', shopIds: [123] })).toBe('loja_completa');
	});
	it('sem contexto → nenhum', () => {
		expect(intentBusca({ keyword: '', shopIds: [] })).toBe('nenhum');
	});
});

describe('busca-config — label de comissão', () => {
	it('formata sem lixo de ponto flutuante (0.07 → "7%")', () => {
		expect(comissaoPercentLabel(0.07)).toBe('7%');
		expect(comissaoPercentLabel(0.15)).toBe('15%');
		expect(comissaoPercentLabel(0)).toBe('0%');
	});
});
