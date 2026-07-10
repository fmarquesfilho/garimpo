import { describe, it, expect } from 'vitest';
import {
	configToPayload,
	payloadToConfig,
	gerarResumo,
	contarFiltrosAtivos,
	cronLabel,
	gerarLabelBusca
} from '$lib/busca-unificada-logic.js';

describe('busca-unificada-logic — configToPayload', () => {
	it('converte config completa para payload', () => {
		const p = configToPayload({
			keywords: ['sérum', 'vitamina c'],
			shopIds: [920292999],
			comissaoMin: 0.1,
			vendasMin: 50,
			categorias: ['cosméticos'],
			fontes: ['curadoria', 'quedas'],
			cron: '0 */8 * * *',
			marketplaces: 'shopee'
		});
		expect(p.keywords).toEqual(['sérum', 'vitamina c']);
		expect(p.shop_ids).toEqual([920292999]);
		expect(p.comissao_min).toBe(0.1);
		expect(p.vendas_min).toBe(50);
		expect(p.categorias).toEqual(['cosméticos']);
		expect(p.fontes).toEqual(['curadoria', 'quedas']);
		expect(p.cron).toBe('0 */8 * * *');
	});

	it('omite campos vazios/default', () => {
		const p = configToPayload({ keywords: [], shopIds: [], comissaoMin: 0, vendasMin: 0, categorias: [] });
		expect(p.keywords).toBeUndefined();
		expect(p.shop_ids).toBeUndefined();
		expect(p.comissao_min).toBeUndefined();
		expect(p.vendas_min).toBeUndefined();
	});

	it('filtra keywords em branco', () => {
		const p = configToPayload({ keywords: ['sérum', '', '  '], shopIds: [] });
		expect(p.keywords).toEqual(['sérum']);
	});
});

describe('busca-unificada-logic — payloadToConfig', () => {
	it('converte response da API para config', () => {
		const c = payloadToConfig({
			id: 'uuid-1',
			keywords: ['sérum'],
			shop_ids: [123],
			shop_names: { 123: 'SKIN1004' },
			cron: '0 */8 * * *',
			comissao_min: 0.1,
			vendas_min: 50,
			categorias: ['cosméticos'],
			fontes: ['curadoria'],
			marketplaces: 'shopee'
		});
		expect(c.id).toBe('uuid-1');
		expect(c.keywords).toEqual(['sérum']);
		expect(c.shopIds).toEqual([123]);
		expect(c.shopNomes).toEqual({ 123: 'SKIN1004' });
		expect(c.cron).toBe('0 */8 * * *');
		expect(c.comissaoMin).toBe(0.1);
	});

	it('trata campos ausentes com defaults', () => {
		const c = payloadToConfig({ id: 'x' });
		expect(c.keywords).toEqual([]);
		expect(c.shopIds).toEqual([]);
		expect(c.comissaoMin).toBe(0);
		expect(c.vendasMin).toBe(0);
		expect(c.categorias).toEqual([]);
		expect(c.fontes).toEqual([]);
		expect(c.cron).toBeNull();
	});
});

describe('busca-unificada-logic — gerarResumo', () => {
	it('gera resumo completo', () => {
		const r = gerarResumo({
			keywords: ['sérum'],
			shopIds: [1, 2],
			comissaoMin: 0.1,
			vendasMin: 50,
			categorias: ['cosméticos'],
			cron: '0 */8 * * *'
		});
		expect(r).toContain('"sérum"');
		expect(r).toContain('2 lojas');
		expect(r).toContain('3 filtros');
		expect(r).toContain('⏱ a cada 8h');
	});

	it('retorna fallback sem dados', () => {
		expect(gerarResumo({ keywords: [], shopIds: [] })).toBe('Nenhum filtro ativo');
	});

	it('singular para 1 loja', () => {
		const r = gerarResumo({ keywords: [], shopIds: [1] });
		expect(r).toContain('1 loja');
	});
});

describe('busca-unificada-logic — contarFiltrosAtivos', () => {
	it('0 quando tudo default', () => {
		expect(contarFiltrosAtivos({ comissaoMin: 0.07, vendasMin: 0, categorias: [] })).toBe(0);
	});

	it('conta cada filtro não-default', () => {
		expect(contarFiltrosAtivos({ comissaoMin: 0.1, vendasMin: 50, categorias: ['x'] })).toBe(3);
	});
});

describe('busca-unificada-logic — cronLabel', () => {
	it('converte crons conhecidos', () => {
		expect(cronLabel('0 */8 * * *')).toBe('a cada 8h');
		expect(cronLabel('0 */12 * * *')).toBe('a cada 12h');
		expect(cronLabel('0 9 * * *')).toBe('diária 9h');
	});

	it('retorna raw para desconhecidos', () => {
		expect(cronLabel('*/5 * * * *')).toBe('*/5 * * * *');
	});

	it('vazio para null', () => {
		expect(cronLabel(null)).toBe('');
	});
});

describe('busca-unificada-logic — gerarLabelBusca', () => {
	it('keywords + lojas', () => {
		const l = gerarLabelBusca({ keywords: ['sérum', 'retinol'], shopIds: [1, 2] });
		expect(l).toBe('sérum, retinol + 2 lojas');
	});

	it('sem keywords', () => {
		const l = gerarLabelBusca({ keywords: [], shopIds: [1] });
		expect(l).toBe('(sem keywords) + 1 loja');
	});

	it('só keywords', () => {
		const l = gerarLabelBusca({ keywords: ['perfume'], shopIds: [] });
		expect(l).toBe('perfume');
	});
});
