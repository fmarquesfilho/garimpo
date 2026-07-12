import { describe, it, expect, vi } from 'vitest';
import { BuscaEngine, guards, STATES, MODOS } from '$lib/busca-engine.svelte.js';

/**
 * Testes da BuscaEngine — cenários reais reportados pelo usuário.
 * Usa mock effects (zero chamadas de API).
 */

function mockEffects(overrides = {}) {
	return {
		carregarBuscasSalvas: vi.fn().mockResolvedValue([]),
		carregarCategorias: vi.fn().mockResolvedValue([{ nome: 'Cosméticos' }, { nome: 'Perfumaria' }]),
		executarBusca: vi.fn().mockResolvedValue({ curadoria: [], quedas: [], novos: [], lojas: [], favoritos: [] }),
		resolverLoja: vi.fn().mockResolvedValue({ id: 920292999, nome: 'Le Botanic', marketplace: 'shopee' }),
		salvarBusca: vi.fn().mockResolvedValue({}),
		removerBusca: vi.fn().mockResolvedValue({}),
		sincronizarStoreExterno: vi.fn().mockResolvedValue(undefined),
		...overrides
	};
}

describe('BuscaEngine — Guards', () => {
	it('temContextoBusca: true com keyword', () => {
		expect(guards.temContextoBusca({ keyword: 'serum', shopIds: [] })).toBe(true);
	});

	it('temContextoBusca: true com shopIds', () => {
		expect(guards.temContextoBusca({ keyword: '', shopIds: [123] })).toBe(true);
	});

	it('temContextoBusca: false sem contexto', () => {
		expect(guards.temContextoBusca({ keyword: '', shopIds: [] })).toBe(false);
	});

	it('lojaInputValida: false para vazio', () => {
		expect(guards.lojaInputValida({}, { value: '' })).toBe(false);
		expect(guards.lojaInputValida({}, { value: '  ' })).toBe(false);
	});

	it('lojaInputValida: true para URL', () => {
		expect(guards.lojaInputValida({}, { value: 'https://shopee.com.br/test' })).toBe(true);
	});
});

describe('BuscaEngine — Inicialização', () => {
	it('começa em IDLE', () => {
		const engine = new BuscaEngine(mockEffects());
		expect(engine.status).toBe(STATES.IDLE);
	});

	it('começa em modo explorando', () => {
		const engine = new BuscaEngine(mockEffects());
		expect(engine.modo).toBe(MODOS.EXPLORANDO);
	});

	it('INICIALIZAR carrega buscas salvas e categorias', async () => {
		const effects = mockEffects({
			carregarBuscasSalvas: vi
				.fn()
				.mockResolvedValue([{ id: 'b1', keywords: ['serum'], shop_ids: [], cron: '0 */8 * * *' }])
		});
		const engine = new BuscaEngine(effects);
		await engine.send({ type: 'INICIALIZAR' });

		expect(effects.carregarBuscasSalvas).toHaveBeenCalled();
		expect(effects.carregarCategorias).toHaveBeenCalled();
		expect(engine.ctx.buscasSalvas).toHaveLength(1);
		expect(engine.ctx.categoriasDisponiveis).toHaveLength(2);
	});
});

describe('BuscaEngine — Busca por keyword', () => {
	it('DIGITAR atualiza keyword no contexto', async () => {
		const engine = new BuscaEngine(mockEffects());
		await engine.send({ type: 'DIGITAR', value: 'serum' });
		expect(engine.ctx.keyword).toBe('serum');
	});
});

describe('BuscaEngine — Adicionar loja + keyword', () => {
	it('ADICIONAR_LOJA resolve e adiciona shopId ao contexto', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);

		engine.ctx.keyword = 'serum';
		await engine.send({ type: 'ADICIONAR_LOJA', value: 'https://s.shopee.com.br/8fQYnxWQqu' });

		expect(effects.resolverLoja).toHaveBeenCalledWith('https://s.shopee.com.br/8fQYnxWQqu');
		expect(engine.ctx.shopIds).toContain(920292999);
		expect(engine.ctx.shopNomes[920292999]).toBe('Le Botanic');
	});

	it('ADICIONAR_LOJA dispara busca com keyword + shopIds combinados', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);

		engine.ctx.keyword = 'serum';
		await engine.send({ type: 'ADICIONAR_LOJA', value: 'lebotanic' });

		// executarBusca é chamado com ctx que tem keyword E shopIds
		expect(effects.executarBusca).toHaveBeenCalled();
		const ctxPassado = effects.executarBusca.mock.calls[0][0];
		expect(ctxPassado.keyword).toBe('serum');
		expect(ctxPassado.shopIds).toContain(920292999);
	});

	it('ADICIONAR_LOJA com erro altera resolucaoLoja para status erro', async () => {
		const effects = mockEffects({
			resolverLoja: vi.fn().mockRejectedValue(new Error('Loja não encontrada'))
		});
		const engine = new BuscaEngine(effects);
		await engine.send({ type: 'ADICIONAR_LOJA', value: 'invalida' });
		expect(engine.ctx.resolucaoLoja.erro).toBe('Loja não encontrada');
		expect(engine.ctx.resolucaoLoja.status).toBe('erro');
	});

	it('REMOVER_LOJA remove shopId e redispara busca', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);

		engine.ctx.shopIds = [123, 456];
		engine.ctx.shopNomes = { 123: 'Loja A', 456: 'Loja B' };
		await engine.send({ type: 'REMOVER_LOJA', shopId: 123 });

		expect(engine.ctx.shopIds).toEqual([456]);
		expect(engine.ctx.shopNomes[123]).toBeUndefined();
	});
});

describe('BuscaEngine — Filtros', () => {
	it('MUDAR_FILTRO normaliza comissão > 1 para decimal', async () => {
		const engine = new BuscaEngine(mockEffects());
		await engine.send({ type: 'MUDAR_FILTRO', comissaoMin: 7 });
		expect(engine.ctx.comissaoMin).toBe(0.07);
	});

	it('MUDAR_FILTRO mantém comissão entre 0 e 1', async () => {
		const engine = new BuscaEngine(mockEffects());
		await engine.send({ type: 'MUDAR_FILTRO', comissaoMin: 0.15 });
		expect(engine.ctx.comissaoMin).toBe(0.15);
	});

	it('MUDAR_FILTRO nunca mostra float com precisão excessiva', async () => {
		const engine = new BuscaEngine(mockEffects());
		await engine.send({ type: 'MUDAR_FILTRO', comissaoMin: 0.07000001 });
		expect(String(engine.ctx.comissaoMin).length).toBeLessThanOrEqual(6); // "0.07" max
	});

	it('MUDAR_FILTRO atualiza vendasMin como inteiro', async () => {
		const engine = new BuscaEngine(mockEffects());
		await engine.send({ type: 'MUDAR_FILTRO', vendasMin: 50.7 });
		expect(engine.ctx.vendasMin).toBe(50);
	});

	it('MUDAR_FILTRO atualiza categorias', async () => {
		const engine = new BuscaEngine(mockEffects());
		await engine.send({ type: 'MUDAR_FILTRO', categorias: ['Cosméticos', 'Perfumaria'] });
		expect(engine.ctx.categorias).toEqual(['Cosméticos', 'Perfumaria']);
	});
});

describe('BuscaEngine — Salvar e Carregar', () => {
	it('SALVAR captura contexto completo', async () => {
		const effects = mockEffects({
			carregarBuscasSalvas: vi.fn().mockResolvedValue([{ id: 'new', keywords: ['serum'], shop_ids: [920292999] }])
		});
		const engine = new BuscaEngine(effects);

		engine.ctx.keyword = 'serum';
		engine.ctx.shopIds = [920292999];
		engine.ctx.comissaoMin = 0.1;
		engine.ctx.cron = '0 */8 * * *';

		await engine.send({ type: 'SALVAR' });

		expect(effects.salvarBusca).toHaveBeenCalledWith(
			expect.objectContaining({
				keywords: ['serum'],
				shop_ids: [920292999],
				comissao_min: 0.1,
				cron: '0 */8 * * *'
			})
		);
	});

	it('CARREGAR_SALVA restaura TUDO e dispara busca', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);

		const config = {
			keywords: ['retinol'],
			shopIds: [111],
			shopNomes: { 111: 'Loja X' },
			comissaoMin: 0.12,
			vendasMin: 100,
			categorias: ['skincare'],
			fontes: ['curadoria', 'quedas'],
			cron: '0 9 * * *'
		};
		await engine.send({ type: 'CARREGAR_SALVA', config });

		expect(engine.ctx.keyword).toBe('retinol');
		expect(engine.ctx.shopIds).toEqual([111]);
		expect(engine.ctx.comissaoMin).toBe(0.12);
		expect(engine.ctx.vendasMin).toBe(100);
		expect(engine.ctx.categorias).toEqual(['skincare']);
		expect(engine.ctx.fontes.curadoria).toBe(true);
		expect(engine.ctx.fontes.novos).toBe(false);
		expect(effects.executarBusca).toHaveBeenCalled();
	});

	it('CARREGAR_SALVA nunca produz label "(sem keywords)"', async () => {
		const engine = new BuscaEngine(mockEffects());
		const config = { keywords: ['serum'], shopIds: [123], shopNomes: {} };

		await engine.send({ type: 'CARREGAR_SALVA', config });

		expect(engine.ctx.keyword).toBe('serum');
		// O label usa a keyword real
		const { gerarLabelBusca } = await import('$lib/busca-unificada-logic.js');
		expect(gerarLabelBusca(config)).not.toContain('sem keywords');
	});
});

describe('BuscaEngine — Fontes', () => {
	it('MUDAR_FONTES atualiza fontes e redispara busca', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);
		engine.ctx.keyword = 'test';

		await engine.send({
			type: 'MUDAR_FONTES',
			fontes: { curadoria: false, quedas: true, novos: true, lojas: false, favoritos: false }
		});

		// Debounce: executar não é chamado imediatamente
		// Mas fontes são atualizadas
		expect(engine.ctx.fontes.curadoria).toBe(false);
		expect(engine.ctx.fontes.quedas).toBe(true);
	});
});

describe('BuscaEngine — Estado', () => {
	it('loading é true durante SEARCHING', async () => {
		const effects = mockEffects({
			executarBusca: () => new Promise(() => {}) // nunca resolve
		});
		const engine = new BuscaEngine(effects);
		engine.ctx.keyword = 'test';

		// Não await — queremos ver o estado intermediário
		engine.send({ type: 'INICIALIZAR' });
		// Após o primeiro tick, deve estar em searching
		await new Promise((r) => setTimeout(r, 10));
		expect(engine.loading).toBe(true);
	});

	it('LIMPAR reseta para contexto inicial', async () => {
		const engine = new BuscaEngine(mockEffects());
		engine.ctx.keyword = 'serum';
		engine.ctx.shopIds = [123];
		engine.ctx.comissaoMin = 0.15;

		await engine.send({ type: 'LIMPAR' });

		expect(engine.ctx.keyword).toBe('');
		expect(engine.ctx.shopIds).toEqual([]);
		expect(engine.ctx.comissaoMin).toBe(0.07);
		expect(engine.status).toBe(STATES.IDLE);
	});
});

// ── v3: Modos de interação ──────────────────────────────────────────────────

describe('BuscaEngine — Modos de interação (v3)', () => {
	it('começa em modo explorando', () => {
		const engine = new BuscaEngine(mockEffects());
		expect(engine.modo).toBe(MODOS.EXPLORANDO);
	});

	it('CARREGAR_SALVA transiciona para modo vinculada', async () => {
		const engine = new BuscaEngine(mockEffects());
		const config = { id: 'b1', keywords: ['serum'], shopIds: [] };

		await engine.send({ type: 'CARREGAR_SALVA', config });

		expect(engine.modo).toBe(MODOS.VINCULADA);
		expect(engine.ctx.buscaSelecionadaId).toBe('b1');
		expect(engine.ctx.editandoId).toBeNull();
	});

	it('EDITAR_SALVA transiciona para modo editando', async () => {
		const engine = new BuscaEngine(mockEffects());
		const config = { id: 'b1', keywords: ['serum'], shopIds: [] };

		await engine.send({ type: 'EDITAR_SALVA', config });

		expect(engine.modo).toBe(MODOS.EDITANDO);
		expect(engine.ctx.buscaSelecionadaId).toBe('b1');
		expect(engine.ctx.editandoId).toBe('b1');
	});

	it('alterar keyword em modo vinculada volta para explorando', async () => {
		const engine = new BuscaEngine(mockEffects());
		const config = { id: 'b1', keywords: ['serum'], shopIds: [] };

		await engine.send({ type: 'CARREGAR_SALVA', config });
		expect(engine.modo).toBe(MODOS.VINCULADA);

		await engine.send({ type: 'DIGITAR', value: 'retinol' });
		expect(engine.modo).toBe(MODOS.EXPLORANDO);
		expect(engine.ctx.buscaSelecionadaId).toBeNull();
	});

	it('ADICIONAR_LOJA em modo vinculada volta para explorando', async () => {
		const engine = new BuscaEngine(mockEffects());
		const config = { id: 'b1', keywords: ['serum'], shopIds: [] };

		await engine.send({ type: 'CARREGAR_SALVA', config });
		expect(engine.modo).toBe(MODOS.VINCULADA);

		await engine.send({ type: 'ADICIONAR_LOJA', value: 'lebotanic' });
		expect(engine.modo).toBe(MODOS.EXPLORANDO);
	});

	it('MUDAR_FILTRO em modo vinculada volta para explorando', async () => {
		const engine = new BuscaEngine(mockEffects());
		const config = { id: 'b1', keywords: ['serum'], shopIds: [] };

		await engine.send({ type: 'CARREGAR_SALVA', config });
		await engine.send({ type: 'MUDAR_FILTRO', comissaoMin: 0.15 });

		expect(engine.modo).toBe(MODOS.EXPLORANDO);
		expect(engine.ctx.buscaSelecionadaId).toBeNull();
	});

	it('MUDAR_MARKETPLACES em modo vinculada volta para explorando', async () => {
		const engine = new BuscaEngine(mockEffects());
		const config = { id: 'b1', keywords: ['serum'], shopIds: [] };

		await engine.send({ type: 'CARREGAR_SALVA', config });
		await engine.send({ type: 'MUDAR_MARKETPLACES', marketplaces: ['shopee'] });

		expect(engine.modo).toBe(MODOS.EXPLORANDO);
	});

	it('CANCELAR_EDICAO volta para explorando e reseta ids', async () => {
		const engine = new BuscaEngine(mockEffects());
		const config = { id: 'b1', keywords: ['serum'], shopIds: [] };

		await engine.send({ type: 'EDITAR_SALVA', config });
		expect(engine.modo).toBe(MODOS.EDITANDO);

		await engine.send({ type: 'CANCELAR_EDICAO' });
		expect(engine.modo).toBe(MODOS.EXPLORANDO);
		expect(engine.ctx.editandoId).toBeNull();
		expect(engine.ctx.buscaSelecionadaId).toBeNull();
		expect(engine.salvarAberto).toBe(false);
	});

	it('CARREGAR_SALVA em modo editando transiciona para vinculada', async () => {
		const engine = new BuscaEngine(mockEffects());
		await engine.send({ type: 'EDITAR_SALVA', config: { id: 'b1', keywords: ['serum'], shopIds: [] } });
		expect(engine.modo).toBe(MODOS.EDITANDO);

		await engine.send({ type: 'CARREGAR_SALVA', config: { id: 'b2', keywords: ['retinol'], shopIds: [] } });
		expect(engine.modo).toBe(MODOS.VINCULADA);
		expect(engine.ctx.buscaSelecionadaId).toBe('b2');
		expect(engine.ctx.editandoId).toBeNull();
	});

	it('SALVAR em modo editando volta para explorando', async () => {
		const effects = mockEffects({
			carregarBuscasSalvas: vi.fn().mockResolvedValue([])
		});
		const engine = new BuscaEngine(effects);

		await engine.send({ type: 'EDITAR_SALVA', config: { id: 'b1', keywords: ['serum'], shopIds: [] } });
		expect(engine.modo).toBe(MODOS.EDITANDO);

		engine.ctx.keyword = 'serum';
		await engine.send({ type: 'SALVAR' });

		expect(engine.modo).toBe(MODOS.EXPLORANDO);
		expect(engine.ctx.editandoId).toBeNull();
		expect(engine.ctx.buscaSelecionadaId).toBeNull();
	});

	it('LIMPAR reseta modo para explorando', async () => {
		const engine = new BuscaEngine(mockEffects());
		await engine.send({ type: 'CARREGAR_SALVA', config: { id: 'b1', keywords: ['serum'], shopIds: [] } });
		expect(engine.modo).toBe(MODOS.VINCULADA);

		await engine.send({ type: 'LIMPAR' });
		expect(engine.modo).toBe(MODOS.EXPLORANDO);
	});
});

// ── v3: Detecção de busca duplicada ──────────────────────────────────────────

describe('BuscaEngine — Detecção de busca duplicada (v3)', () => {
	it('buscaDuplicada retorna null quando não há duplicata', () => {
		const engine = new BuscaEngine(mockEffects());
		engine.ctx.keyword = 'serum';
		engine.ctx.buscasSalvas = [{ id: 'b1', keywords: ['retinol'], shopIds: [], categorias: [] }];

		expect(engine.buscaDuplicada).toBeNull();
	});

	it('buscaDuplicada retorna a busca existente quando parâmetros coincidem', () => {
		const engine = new BuscaEngine(mockEffects());
		engine.ctx.keyword = 'serum';
		engine.ctx.shopIds = [];
		engine.ctx.categorias = [];
		engine.ctx.marketplacesFiltro = [];
		engine.ctx.buscasSalvas = [{ id: 'b1', keywords: ['serum'], shopIds: [], categorias: [], marketplaces: 'shopee' }];

		expect(engine.buscaDuplicada).not.toBeNull();
		expect(engine.buscaDuplicada.id).toBe('b1');
	});

	it('SALVAR com busca duplicada bloqueia e seta erroDuplicata', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);

		engine.ctx.keyword = 'serum';
		engine.ctx.buscasSalvas = [{ id: 'b1', keywords: ['serum'], shopIds: [], categorias: [], marketplaces: 'shopee' }];

		await engine.send({ type: 'SALVAR' });

		expect(effects.salvarBusca).not.toHaveBeenCalled();
		expect(engine.ctx.erroDuplicata).toContain('serum');
	});

	it('SALVAR em edit mode exclui a própria busca da verificação de duplicata', async () => {
		const effects = mockEffects({
			carregarBuscasSalvas: vi.fn().mockResolvedValue([{ id: 'b1', keywords: ['serum'], shop_ids: [] }])
		});
		const engine = new BuscaEngine(effects);

		engine.ctx.keyword = 'serum';
		engine.ctx.editandoId = 'b1';
		engine.ctx.buscasSalvas = [{ id: 'b1', keywords: ['serum'], shopIds: [], categorias: [], marketplaces: 'shopee' }];

		await engine.send({ type: 'SALVAR' });

		// Deve permitir salvar (está editando a própria busca)
		expect(effects.salvarBusca).toHaveBeenCalled();
		expect(engine.ctx.erroDuplicata).toBeNull();
	});

	it('erroDuplicata é limpo a cada nova interação', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);

		engine.ctx.keyword = 'serum';
		engine.ctx.buscasSalvas = [{ id: 'b1', keywords: ['serum'], shopIds: [], categorias: [], marketplaces: 'shopee' }];
		await engine.send({ type: 'SALVAR' });
		expect(engine.ctx.erroDuplicata).not.toBeNull();

		// Qualquer interação limpa o erro
		await engine.send({ type: 'DIGITAR', value: 'retinol' });
		expect(engine.ctx.erroDuplicata).toBeNull();
	});
});
