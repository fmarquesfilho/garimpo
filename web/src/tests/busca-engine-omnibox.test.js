/**
 * Tests for BuscaEngine OMNIBOX_* handlers and Smart Search.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { BuscaEngine, STATES } from '$lib/busca-engine.svelte.js';
import { criarUIInicial } from '$lib/busca-engine-state.js';

function criarMockEffects() {
	return {
		carregarBuscasSalvas: vi.fn().mockResolvedValue([]),
		sincronizarStoreExterno: vi.fn().mockResolvedValue(undefined),
		carregarCategorias: vi.fn().mockResolvedValue([
			{ nome: 'Beleza', marketplaces: ['shopee'] },
			{ nome: 'Moda', marketplaces: ['shopee', 'mercado_livre'] }
		]),
		carregarRegistroLojas: vi.fn().mockResolvedValue([
			{ id: '100', nome: 'Glory of Seoul', nome_normalizado: 'gloryofseoul', marketplace: 'shopee' }
		]),
		executarBusca: vi.fn().mockResolvedValue({ curadoria: [], quedas: [], novos: [], lojas: [], favoritos: [] }),
		resolverLoja: vi.fn().mockResolvedValue({ id: '999', nome: 'New Shop', marketplace: 'shopee' }),
		salvarBusca: vi.fn().mockResolvedValue({}),
		removerBusca: vi.fn().mockResolvedValue({}),
		buscarLojasPorNome: vi.fn().mockResolvedValue({
			lojas: [
				{ id: '100', nome: 'Glory of Seoul', marketplace: 'shopee', monitorada: false }
			],
			total: 1
		})
	};
}

describe('BuscaEngine — OMNIBOX_* handlers', () => {
	let engine;
	let effects;

	beforeEach(async () => {
		effects = criarMockEffects();
		engine = new BuscaEngine(effects);
		await engine.send({ type: 'INICIALIZAR' });
	});

	describe('OMNIBOX_INPUT', () => {
		it('sets inputValue in ui.omnibox', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
			expect(engine.omnibox.inputValue).toBe('serum');
		});

		it('opens the dropdown', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
			expect(engine.omnibox.aberto).toBe(true);
		});

		it('resets highlightIdx', () => {
			engine.ui.omnibox.highlightIdx = 2;
			engine.send({ type: 'OMNIBOX_INPUT', value: 'se' });
			expect(engine.omnibox.highlightIdx).toBe(-1);
		});

		it('generates intencao options for text without prefix', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
			expect(engine.omnibox.modo).toBe('intencao');
			expect(engine.omnibox.opcoes.length).toBeGreaterThanOrEqual(2);
			const tipos = engine.omnibox.opcoes.map((o) => o.tipo);
			expect(tipos).toContain('produtos');
			expect(tipos).toContain('lojas');
		});

		it('generates sugestoes prefixo for prefix token', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: '@gl' });
			expect(engine.omnibox.modo).toBe('sugestoes');
		});

		it('generates resolver_link for URL', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'https://shopee.com.br/shop/123' });
			expect(engine.omnibox.modo).toBe('intencao');
			expect(engine.omnibox.opcoes.length).toBe(1);
			expect(engine.omnibox.opcoes[0].tipo).toBe('resolver_link');
		});

		it('empty opcoes for text < 2 chars', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'a' });
			expect(engine.omnibox.opcoes).toHaveLength(0);
		});
	});

	describe('OMNIBOX_KEYDOWN', () => {
		it('ArrowDown increments highlightIdx', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
			engine.send({ type: 'OMNIBOX_KEYDOWN', key: 'ArrowDown' });
			expect(engine.omnibox.highlightIdx).toBe(0);
		});

		it('ArrowDown cycles (last -> first)', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
			const n = engine.omnibox.opcoes.length;
			// Navigate to last item (n presses from -1 -> 0, 1, ..., n-1)
			for (let i = 0; i < n; i++) {
				engine.send({ type: 'OMNIBOX_KEYDOWN', key: 'ArrowDown' });
			}
			expect(engine.omnibox.highlightIdx).toBe(n - 1);
			// One more press cycles back to 0
			engine.send({ type: 'OMNIBOX_KEYDOWN', key: 'ArrowDown' });
			expect(engine.omnibox.highlightIdx).toBe(0);
		});

		it('ArrowUp cycles (first -> last)', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
			engine.send({ type: 'OMNIBOX_KEYDOWN', key: 'ArrowUp' });
			const n = engine.omnibox.opcoes.length;
			expect(engine.omnibox.highlightIdx).toBe(n - 1);
		});

		it('Escape closes dropdown', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
			engine.send({ type: 'OMNIBOX_KEYDOWN', key: 'Escape' });
			expect(engine.omnibox.aberto).toBe(false);
			expect(engine.omnibox.highlightIdx).toBe(-1);
		});

		it('Enter executes first option when no highlight', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
			engine.send({ type: 'OMNIBOX_KEYDOWN', key: 'Enter' });
			// First option is "Pesquisar em Produtos" -> should set keyword and trigger busca
			expect(engine.ctx.keyword).toBe('serum');
			expect(engine.omnibox.aberto).toBe(false);
		});

		it('Enter executes highlighted option', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
			engine.send({ type: 'OMNIBOX_KEYDOWN', key: 'ArrowDown' }); // highlight 0 = produtos
			engine.send({ type: 'OMNIBOX_KEYDOWN', key: 'ArrowDown' }); // highlight 1 = lojas
			engine.send({ type: 'OMNIBOX_KEYDOWN', key: 'Enter' });
			// Second option is "Pesquisar em Lojas" -> should trigger buscarLojas
			expect(engine.modoResultados).toBe('lojas');
		});

		it('highlight via mouse sets idx', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
			engine.send({ type: 'OMNIBOX_KEYDOWN', key: 'highlight', idx: 1 });
			expect(engine.omnibox.highlightIdx).toBe(1);
		});
	});

	describe('OMNIBOX_SELECIONAR', () => {
		it('executes the option at given index', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
			engine.send({ type: 'OMNIBOX_SELECIONAR', indice: 0 });
			// First option is "Pesquisar em Produtos"
			expect(engine.ctx.keyword).toBe('serum');
			expect(engine.omnibox.aberto).toBe(false);
		});
	});

	describe('OMNIBOX_BLUR', () => {
		it('closes the dropdown', () => {
			engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
			engine.send({ type: 'OMNIBOX_BLUR' });
			expect(engine.omnibox.aberto).toBe(false);
		});
	});

	describe('BUSCAR_LOJAS', () => {
		it('calls buscarLojasPorNome and populates resultadosLojas', async () => {
			await engine.send({ type: 'BUSCAR_LOJAS', termo: 'glory' });
			expect(effects.buscarLojasPorNome).toHaveBeenCalledWith('glory');
			expect(engine.modoResultados).toBe('lojas');
			expect(engine.resultadosLojas).toHaveLength(1);
			expect(engine.resultadosLojas[0].nome).toBe('Glory of Seoul');
		});

		it('does nothing for term < 2 chars', async () => {
			await engine.send({ type: 'BUSCAR_LOJAS', termo: 'g' });
			expect(effects.buscarLojasPorNome).not.toHaveBeenCalled();
		});

		it('handles error gracefully', async () => {
			effects.buscarLojasPorNome.mockRejectedValue(new Error('Network error'));
			await engine.send({ type: 'BUSCAR_LOJAS', termo: 'test' });
			expect(engine.status).toBe(STATES.ERROR);
			expect(engine.resultadosLojas).toHaveLength(0);
		});
	});

	describe('MONITORAR_LOJA', () => {
		it('adds loja to scope and updates resultadosLojas', async () => {
			// First set some results
			engine.ui.resultados.lojas = [
				{ id: '100', nome: 'Glory of Seoul', marketplace: 'shopee', monitorada: false }
			];
			await engine.send({ type: 'MONITORAR_LOJA', loja: { id: '100', nome: 'Glory of Seoul', marketplace: 'shopee' } });
			expect(engine.ctx.shopIds).toContain(100);
			expect(engine.resultadosLojas[0].monitorada).toBe(true);
		});
	});

	describe('DIGITAR restores modo', () => {
		it('restores resultados mode to produtos', async () => {
			await engine.send({ type: 'BUSCAR_LOJAS', termo: 'glory' });
			expect(engine.modoResultados).toBe('lojas');
			engine.send({ type: 'DIGITAR', value: 'serum' });
			expect(engine.modoResultados).toBe('produtos');
		});
	});

	describe('criarUIInicial', () => {
		it('returns correct shape', () => {
			const ui = criarUIInicial();
			expect(ui.omnibox.inputValue).toBe('');
			expect(ui.omnibox.aberto).toBe(false);
			expect(ui.omnibox.highlightIdx).toBe(-1);
			expect(ui.omnibox.modo).toBe('intencao');
			expect(ui.omnibox.opcoes).toEqual([]);
			expect(ui.resultados.modo).toBe('produtos');
			expect(ui.resultados.lojas).toEqual([]);
			expect(ui.paineis.buscasSalvasAberto).toBe(false);
			expect(ui.paineis.filtrosAberto).toBe(false);
			expect(ui.paineis.salvarAberto).toBe(false);
		});
	});
});
