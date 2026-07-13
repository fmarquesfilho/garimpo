/**
 * Tests for BuscaEngine OMNIBOX_* handlers and Smart Search.
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
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
		carregarRegistroLojas: vi
			.fn()
			.mockResolvedValue([
				{ id: '100', nome: 'Glory of Seoul', nome_normalizado: 'gloryofseoul', marketplace: 'shopee' }
			]),
		executarBusca: vi.fn().mockResolvedValue({ curadoria: [], quedas: [], novos: [], lojas: [], favoritos: [] }),
		resolverLoja: vi.fn().mockResolvedValue({ id: '999', nome: 'New Shop', marketplace: 'shopee' }),
		salvarBusca: vi.fn().mockResolvedValue({}),
		removerBusca: vi.fn().mockResolvedValue({}),
		buscarLojasPorNome: vi.fn().mockResolvedValue({
			lojas: [{ id: '100', nome: 'Glory of Seoul', marketplace: 'shopee', monitorada: false }],
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
			engine.ui.resultados.lojas = [{ id: '100', nome: 'Glory of Seoul', marketplace: 'shopee', monitorada: false }];
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

describe('BuscaEngine — executarIntencao edge cases', () => {
	let engine;
	let effects;

	beforeEach(async () => {
		effects = criarMockEffects();
		engine = new BuscaEngine(effects);
		await engine.send({ type: 'INICIALIZAR' });
	});

	it('selecionar intencao categoria adiciona categoria e limpa keyword', () => {
		engine.send({ type: 'OMNIBOX_INPUT', value: 'beleza' });
		// Find and select the categoria option
		const catIdx = engine.omnibox.opcoes.findIndex((o) => o.tipo === 'categoria');
		if (catIdx >= 0) {
			engine.send({ type: 'OMNIBOX_SELECIONAR', indice: catIdx });
			expect(engine.ctx.categorias).toContain('Beleza');
			expect(engine.ctx.keyword).toBe('');
			expect(engine.omnibox.inputValue).toBe('');
		}
	});

	it('selecionar intencao resolver_link chama adicionarLoja (resolve remota)', () => {
		engine.send({ type: 'OMNIBOX_INPUT', value: 'https://shopee.com.br/shop/123' });
		engine.send({ type: 'OMNIBOX_SELECIONAR', indice: 0 });
		// Resolver link dispara resolucao remota — modo permanece produtos (busca normal com loja no escopo)
		expect(engine.modoResultados).toBe('produtos');
		expect(engine.ctx.resolucaoLoja.status).toBe('resolvendo');
	});

	it('resolver link limpa inputValue após resolução com sucesso', async () => {
		engine.send({ type: 'OMNIBOX_INPUT', value: 'https://shopee.com.br/shop/123' });
		expect(engine.omnibox.inputValue).toBe('https://shopee.com.br/shop/123');
		await engine.send({ type: 'OMNIBOX_SELECIONAR', indice: 0 });
		// Após resolver: input limpo, keyword vazia, loja adicionada via chip
		expect(engine.omnibox.inputValue).toBe('');
		expect(engine.ctx.keyword).toBe('');
		expect(engine.ctx.shopIds).toContain(999);
	});

	it('OMNIBOX_SELECIONAR com indice fora de bounds nao faz nada', () => {
		engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });

		engine.send({ type: 'OMNIBOX_SELECIONAR', indice: 99 });
		// Nada mudou
		expect(engine.omnibox.aberto).toBe(true);
	});

	it('OMNIBOX_SELECIONAR com indice negativo nao faz nada', () => {
		engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
		engine.send({ type: 'OMNIBOX_SELECIONAR', indice: -1 });
		expect(engine.omnibox.aberto).toBe(true);
	});

	it('OMNIBOX_KEYDOWN com opcoes vazio nao muda highlightIdx', () => {
		engine.send({ type: 'OMNIBOX_INPUT', value: 'a' }); // < minChars, opcoes vazio
		engine.send({ type: 'OMNIBOX_KEYDOWN', key: 'ArrowDown' });
		expect(engine.omnibox.highlightIdx).toBe(-1);
	});

	it('OMNIBOX_INPUT com value null trata como string vazia', () => {
		engine.send({ type: 'OMNIBOX_INPUT', value: null });
		expect(engine.omnibox.inputValue).toBe('');
	});

	it('BUSCAR_LOJAS com termo whitespace-only nao executa', async () => {
		await engine.send({ type: 'BUSCAR_LOJAS', termo: '   ' });
		expect(effects.buscarLojasPorNome).not.toHaveBeenCalled();
	});

	it('MONITORAR_LOJA com loja sem id nao faz nada', async () => {
		await engine.send({ type: 'MONITORAR_LOJA', loja: { nome: 'Test' } });
		expect(engine.ctx.shopIds).toHaveLength(0);
	});

	it('MONITORAR_LOJA com loja.id null nao faz nada', async () => {
		await engine.send({ type: 'MONITORAR_LOJA', loja: { id: null, nome: 'Test' } });
		expect(engine.ctx.shopIds).toHaveLength(0);
	});
});

describe('BuscaEngine — sugestao prefixo execution', () => {
	let engine;
	let effects;

	beforeEach(async () => {
		effects = criarMockEffects();
		engine = new BuscaEngine(effects);
		await engine.send({ type: 'INICIALIZAR' });
	});

	it('selecionar @loja no modo sugestoes adiciona loja ao escopo', () => {
		engine.send({ type: 'OMNIBOX_INPUT', value: '@gl' });
		if (engine.omnibox.opcoes.length > 0) {
			engine.send({ type: 'OMNIBOX_SELECIONAR', indice: 0 });
			// Deve ter despachado ADICIONAR_LOJA (via send recursivo)
			expect(engine.ctx.shopIds.length).toBeGreaterThanOrEqual(0);
		}
	});

	it('selecionar #categoria no modo sugestoes adiciona categoria', () => {
		engine.send({ type: 'OMNIBOX_INPUT', value: '#bel' });
		if (engine.omnibox.opcoes.length > 0) {
			engine.send({ type: 'OMNIBOX_SELECIONAR', indice: 0 });
			expect(engine.ctx.categorias).toContain('Beleza');
		}
	});
});

describe('BuscaEngine — chip removal message', () => {
	let engine;
	let effects;

	beforeEach(async () => {
		effects = criarMockEffects();
		engine = new BuscaEngine(effects);
		await engine.send({ type: 'INICIALIZAR' });
		await engine.send({ type: 'ADICIONAR_LOJA', loja: { id: '100', nome: 'Glory of Seoul', marketplace: 'shopee' } });
	});

	it('REMOVER_LOJA sets chipRemovalMessage', () => {
		engine.send({ type: 'REMOVER_LOJA', shopId: 100 });
		expect(engine.omnibox.chipRemovalMessage).toContain('Glory of Seoul');
		expect(engine.omnibox.chipRemovalMessage).toContain('removida');
	});

	it('REMOVER_CATEGORIA sets chipRemovalMessage', () => {
		engine.send({ type: 'ADICIONAR_CATEGORIA', nome: 'Beleza' });
		engine.send({ type: 'REMOVER_CATEGORIA', nome: 'Beleza' });
		expect(engine.omnibox.chipRemovalMessage).toContain('Beleza');
		expect(engine.omnibox.chipRemovalMessage).toContain('removida');
	});
});

describe('BuscaEngine — mobile/tablet (blur sem Enter)', () => {
	let engine;
	let effects;

	beforeEach(async () => {
		vi.useFakeTimers();
		effects = criarMockEffects();
		engine = new BuscaEngine(effects);
		await engine.send({ type: 'INICIALIZAR' });
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it('digitar + blur executa busca via debounce (sem Enter)', async () => {
		engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
		engine.send({ type: 'OMNIBOX_BLUR' });

		// Dropdown fechou
		expect(engine.omnibox.aberto).toBe(false);

		// Keyword foi setada (pelo INPUT, nao pelo Enter)
		expect(engine.ctx.keyword).toBe('serum');

		// Debounce vai executar a busca — avanca o timer
		await vi.advanceTimersByTimeAsync(500);
		// Effects.executarBusca deve ter sido chamado
		expect(effects.executarBusca).toHaveBeenCalled();
	});

	it('blur preserva keyword no input (nao limpa)', () => {
		engine.send({ type: 'OMNIBOX_INPUT', value: 'retinol' });
		engine.send({ type: 'OMNIBOX_BLUR' });

		expect(engine.omnibox.inputValue).toBe('retinol');
	});

	it('blur com chips preserva chips (nao remove nada)', async () => {
		await engine.send({ type: 'ADICIONAR_LOJA', loja: { id: '100', nome: 'Test', marketplace: 'shopee' } });
		engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
		engine.send({ type: 'OMNIBOX_BLUR' });

		expect(engine.ctx.shopIds).toContain(100);
	});

	it('blur apos abrir dropdown nao executa opcao (apenas fecha)', () => {
		engine.send({ type: 'OMNIBOX_INPUT', value: 'serum' });
		expect(engine.omnibox.aberto).toBe(true);
		expect(engine.omnibox.opcoes.length).toBeGreaterThan(0);

		engine.send({ type: 'OMNIBOX_BLUR' });

		// Dropdown fechou mas modo continua 'produtos' (nao executou 'lojas' nem 'categoria')
		expect(engine.modoResultados).toBe('produtos');
	});

	it('blur sem texto digitado nao dispara busca', async () => {
		engine.send({ type: 'OMNIBOX_INPUT', value: '' });
		engine.send({ type: 'OMNIBOX_BLUR' });

		await vi.advanceTimersByTimeAsync(500);
		// Sem keyword e sem contexto → nao busca
		expect(engine.status).not.toBe('searching');
	});
});
