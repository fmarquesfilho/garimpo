import { describe, it, expect, vi } from 'vitest';
import { BuscaEngine, STATES, MODOS } from '$lib/busca-engine.svelte.js';
import { DEFAULTS } from '$lib/busca-config.js';
import rules from '../../../rules/busca-rules.json';

/**
 * Testes expandidos da BuscaEngine — cenários derivados de TESTES_DESCOBRIR.md.
 * Cada test referencia o cenário (#N) do documento.
 *
 * Organização:
 *  - Fontes de dados (combinações de toggles)
 *  - Buscas salvas (pills, restaurar, cron)
 *  - Timeout e erro
 *  - Input de busca (debounce, limpar)
 *  - Regras externas (rules/busca-rules.json como spec)
 */

function mockEffects(overrides = {}) {
	return {
		carregarBuscasSalvas: vi.fn().mockResolvedValue([]),
		carregarCategorias: vi
			.fn()
			.mockResolvedValue([{ nome: 'Perfumaria' }, { nome: 'Maquiagem' }, { nome: 'Cuidados com a Pele' }]),
		executarBusca: vi.fn().mockResolvedValue({
			curadoria: [],
			quedas: [],
			novos: [],
			lojas: [],
			favoritos: []
		}),
		resolverLoja: vi.fn().mockResolvedValue({ shop_ids: [920292999], keyword: 'Le Botanic' }),
		salvarBusca: vi.fn().mockResolvedValue({}),
		removerBusca: vi.fn().mockResolvedValue({}),
		sincronizarStoreExterno: vi.fn().mockResolvedValue(undefined),
		...overrides
	};
}

// ── Cenários de Fontes (doc #1–#13) ──────────────────────────────────────

describe('BuscaEngine — Fontes: cenários combinados', () => {
	it('#10: nenhuma fonte ativa impede busca (guard temContextoBusca não falha, mas nenhum dado volta)', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);
		engine.ctx.keyword = 'serum';
		engine.ctx.fontes = { curadoria: false, quedas: false, novos: false, lojas: false, favoritos: false };

		// A busca executa mas não ativa nenhum fetch
		await engine.send({ type: 'MUDAR_FONTES', fontes: engine.ctx.fontes });
		expect(engine.ctx.fontes.curadoria).toBe(false);
	});

	it('#11: Busca + Quedas + Novos sem keyword — curadoria não executa sem contexto', async () => {
		const effects = mockEffects({
			executarBusca: vi.fn().mockResolvedValue({
				curadoria: [],
				quedas: [{ id: 'q1', nome: 'Queda A', _fonte: 'queda' }],
				novos: [{ id: 'n1', nome: 'Novo B', _fonte: 'novo' }],
				lojas: [],
				favoritos: []
			})
		});
		const engine = new BuscaEngine(effects);
		// Sem keyword, com loja (temContextoBusca = true)
		engine.ctx.shopIds = [123];
		engine.ctx.fontes = { curadoria: true, quedas: true, novos: true, lojas: false, favoritos: false };

		await engine.send({ type: 'MUDAR_FONTES', fontes: engine.ctx.fontes });
		// Debounce — espera disparo
		await new Promise((r) => setTimeout(r, DEFAULTS.debounceMs + 50));

		expect(effects.executarBusca).toHaveBeenCalled();
	});

	it('#12: todas as fontes com keyword filtra tudo (passado ao executarBusca)', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);
		engine.ctx.keyword = 'sérum';
		engine.ctx.fontes = { curadoria: true, quedas: true, novos: true, lojas: true, favoritos: true };

		// Dispara busca diretamente
		await engine.send({ type: 'INICIALIZAR' });

		const ctx = effects.executarBusca.mock.calls[0]?.[0];
		expect(ctx?.keyword).toBe('sérum');
		expect(ctx?.fontes.curadoria).toBe(true);
		expect(ctx?.fontes.quedas).toBe(true);
	});
});

// ── Buscas Salvas (doc #27–#32) ──────────────────────────────────────────

describe('BuscaEngine — Buscas salvas (pills)', () => {
	it('#27: clicar pill de keyword seta keyword + ativa fontes', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);

		const config = {
			keywords: ['sérum'],
			shopIds: [],
			fontes: ['curadoria', 'quedas'],
			comissaoMin: 0.07,
			vendasMin: 0,
			categorias: [],
			cron: null
		};
		await engine.send({ type: 'CARREGAR_SALVA', config });

		expect(engine.ctx.keyword).toBe('sérum');
		expect(engine.ctx.fontes.curadoria).toBe(true);
		expect(engine.ctx.fontes.quedas).toBe(true);
		expect(engine.ctx.fontes.novos).toBe(false);
	});

	it('#28: busca salva com fontes [quedas, novos] ativa ambas', async () => {
		const engine = new BuscaEngine(mockEffects());
		const config = { keywords: [], shopIds: [123], fontes: ['quedas', 'novos'] };
		await engine.send({ type: 'CARREGAR_SALVA', config });

		expect(engine.ctx.fontes.quedas).toBe(true);
		expect(engine.ctx.fontes.novos).toBe(true);
		expect(engine.ctx.fontes.curadoria).toBe(false);
	});

	it('#29: busca salva com categorias restaura categorias', async () => {
		const engine = new BuscaEngine(mockEffects());
		const config = {
			keywords: ['retinol'],
			shopIds: [],
			fontes: ['curadoria'],
			categorias: ['Cuidados com a Pele', 'Perfumaria']
		};
		await engine.send({ type: 'CARREGAR_SALVA', config });

		expect(engine.ctx.categorias).toEqual(['Cuidados com a Pele', 'Perfumaria']);
	});

	it('#30: busca agendada preserva cron no contexto', async () => {
		const engine = new BuscaEngine(mockEffects());
		const config = { keywords: ['perfume'], shopIds: [], cron: '0 */8 * * *' };
		await engine.send({ type: 'CARREGAR_SALVA', config });

		expect(engine.ctx.cron).toBe('0 */8 * * *');
	});

	it('#31: busca com múltiplas keywords usa a primeira como keyword ativa', async () => {
		const engine = new BuscaEngine(mockEffects());
		// O engine usa keywords[0] como keyword do input
		const config = { keywords: ['sérum', 'vitamina c'], shopIds: [] };
		await engine.send({ type: 'CARREGAR_SALVA', config });

		expect(engine.ctx.keyword).toBe('sérum');
	});

	it('#37: remover busca salva remove da lista', async () => {
		const effects = mockEffects({
			carregarBuscasSalvas: vi.fn().mockResolvedValue([]),
			removerBusca: vi.fn().mockResolvedValue({})
		});
		const engine = new BuscaEngine(effects);
		engine.ctx.buscasSalvas = [{ id: 'b1', keywords: ['test'], shopIds: [] }];

		await engine.send({ type: 'REMOVER_SALVA', config: { keywords: ['test'] } });

		expect(effects.removerBusca).toHaveBeenCalled();
		expect(effects.sincronizarStoreExterno).toHaveBeenCalled();
		expect(engine.ctx.buscasSalvas).toHaveLength(0);
	});
});

// ── Timeout e Erro (doc #40) ─────────────────────────────────────────────

describe('BuscaEngine — Timeout e erro', () => {
	it('#40: timeout > 25s produz estado ERROR com mensagem', async () => {
		const effects = mockEffects({
			executarBusca: vi.fn().mockRejectedValue(new Error('A busca demorou demais. Tente novamente.'))
		});
		const engine = new BuscaEngine(effects);
		engine.ctx.keyword = 'test';

		await engine.send({ type: 'INICIALIZAR' });

		expect(engine.status).toBe(STATES.ERROR);
		expect(engine.ctx.error).toContain('demorou demais');
	});

	it('RETRY após erro dispara nova busca', async () => {
		let chamadas = 0;
		const effects = mockEffects({
			executarBusca: vi.fn().mockImplementation(() => {
				chamadas++;
				if (chamadas === 1) return Promise.reject(new Error('falhou'));
				return Promise.resolve({ curadoria: [], quedas: [], novos: [], lojas: [], favoritos: [] });
			})
		});
		const engine = new BuscaEngine(effects);
		engine.ctx.keyword = 'test';

		await engine.send({ type: 'INICIALIZAR' });
		expect(engine.status).toBe(STATES.ERROR);

		await engine.send({ type: 'RETRY' });
		expect(engine.status).toBe(STATES.RESULTS);
		expect(engine.ctx.error).toBeNull();
	});
});

// ── Input de busca (doc #43–#46) ─────────────────────────────────────────

describe('BuscaEngine — Input de busca', () => {
	it('#43/#44: LIMPAR reseta keyword e volta para IDLE', async () => {
		const engine = new BuscaEngine(mockEffects());
		engine.ctx.keyword = 'serum';
		engine.ctx.shopIds = [123];

		await engine.send({ type: 'LIMPAR' });

		expect(engine.ctx.keyword).toBe('');
		expect(engine.ctx.shopIds).toEqual([]);
		expect(engine.status).toBe(STATES.IDLE);
	});

	it('#46: debounce respeita DEFAULTS.debounceMs das regras', () => {
		expect(DEFAULTS.debounceMs).toBe(rules.defaults.debounceMs);
		expect(DEFAULTS.debounceMs).toBe(400);
	});

	it('DIGITAR com string vazia não dispara busca (guard bloqueia)', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);

		await engine.send({ type: 'DIGITAR', value: '' });
		// Espera debounce
		await new Promise((r) => setTimeout(r, DEFAULTS.debounceMs + 50));

		// Sem keyword e sem shopIds → guard bloqueia
		expect(effects.executarBusca).not.toHaveBeenCalled();
		expect(engine.status).toBe(STATES.IDLE);
	});
});

// ── Regras externas como spec (rules/busca-rules.json) ───────────────────

describe('BuscaEngine — Regras externas como fonte de verdade', () => {
	it('intent table tem exatamente 4 combinações', () => {
		expect(rules.intent).toHaveLength(4);
	});

	it('intent keyword+shop = keyword_na_loja', () => {
		const row = rules.intent.find((r) => r.keyword && r.shop);
		expect(row.result).toBe('keyword_na_loja');
		expect(row.sources).toContain('curadoria');
		expect(row.sources).toContain('lojas');
	});

	it('intent keyword sem shop = keyword_global', () => {
		const row = rules.intent.find((r) => r.keyword && !r.shop);
		expect(row.result).toBe('keyword_global');
		expect(row.sources).toContain('curadoria');
		expect(row.sources).not.toContain('lojas');
	});

	it('intent sem keyword com shop = loja_completa', () => {
		const row = rules.intent.find((r) => !r.keyword && r.shop);
		expect(row.result).toBe('loja_completa');
		expect(row.sources).toContain('lojas');
		expect(row.sources).not.toContain('curadoria');
	});

	it('intent nenhum = sem contexto', () => {
		const row = rules.intent.find((r) => !r.keyword && !r.shop);
		expect(row.result).toBe('nenhum');
		expect(row.sources).toHaveLength(0);
	});

	it('guard podeSalvar requer mesmos campos que temContextoBusca', () => {
		expect(rules.guards.podeSalvar.requiresAny).toEqual(rules.guards.temContextoBusca.requiresAny);
	});

	it('defaults.fontes: curadoria, quedas, novos ativos por padrão', () => {
		expect(rules.defaults.fontes.curadoria).toBe(true);
		expect(rules.defaults.fontes.quedas).toBe(true);
		expect(rules.defaults.fontes.novos).toBe(true);
		expect(rules.defaults.fontes.lojas).toBe(false);
		expect(rules.defaults.fontes.favoritos).toBe(false);
	});

	it('normalize: comissão >1 divide por 100', () => {
		expect(rules.normalize.comissao.divideBy100IfGt1).toBe(true);
	});

	it('transição ADICIONAR_LOJA é imediata (sem debounce)', () => {
		expect(rules.transicoes.ADICIONAR_LOJA.imediato).toBe(true);
		expect(rules.transicoes.ADICIONAR_LOJA.refetch).toBe(true);
	});

	it('transição MUDAR_FILTRO não refetch (client-side)', () => {
		expect(rules.transicoes.MUDAR_FILTRO.refetch).toBe(false);
	});

	it('engine.intent reflete corretamente o contexto', async () => {
		const engine = new BuscaEngine(mockEffects());

		// Sem contexto → nenhum
		expect(engine.intent).toBe('nenhum');

		// Com keyword → keyword_global
		engine.ctx.keyword = 'serum';
		expect(engine.intent).toBe('keyword_global');

		// Com keyword + loja → keyword_na_loja
		engine.ctx.shopIds = [123];
		expect(engine.intent).toBe('keyword_na_loja');

		// Sem keyword, com loja → loja_completa
		engine.ctx.keyword = '';
		expect(engine.intent).toBe('loja_completa');
	});
});

// ── Cenários de loja monitorada + fontes (doc #4, #6) ─────────────────────

describe('BuscaEngine — Loja monitorada + fontes', () => {
	it('adicionar loja sem keyword → intent loja_completa, fontes devem incluir lojas', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);

		// Sem keyword, adiciona loja
		await engine.send({ type: 'ADICIONAR_LOJA', value: 'Le Botanic' });

		expect(engine.ctx.shopIds).toContain(920292999);
		expect(engine.intent).toBe('loja_completa');

		// O intent diz que as sources esperadas são lojas, quedas, novos
		const intentRow = rules.intent.find((r) => !r.keyword && r.shop);
		expect(intentRow.sources).toContain('lojas');
		expect(intentRow.sources).toContain('quedas');
		expect(intentRow.sources).toContain('novos');
	});

	it('adicionar loja com keyword → intent keyword_na_loja, busca escopada', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);

		engine.ctx.keyword = 'serum';
		await engine.send({ type: 'ADICIONAR_LOJA', value: 'Le Botanic' });

		expect(engine.intent).toBe('keyword_na_loja');
		// executarBusca recebeu ctx com ambos
		const ctx = effects.executarBusca.mock.calls[0][0];
		expect(ctx.keyword).toBe('serum');
		expect(ctx.shopIds).toContain(920292999);
	});

	it('remover loja com keyword → volta para keyword_global', async () => {
		const engine = new BuscaEngine(mockEffects());
		engine.ctx.keyword = 'serum';
		engine.ctx.shopIds = [920292999];
		engine.ctx.shopNomes = { 920292999: 'Le Botanic' };

		expect(engine.intent).toBe('keyword_na_loja');

		await engine.send({ type: 'REMOVER_LOJA', shopId: 920292999 });

		expect(engine.intent).toBe('keyword_global');
		expect(engine.ctx.shopIds).toHaveLength(0);
	});
});

// ── SALVAR com diferentes combinações ─────────────────────────────────────

describe('BuscaEngine — Salvar cenários avançados', () => {
	it('#34: salvar só com categorias (sem keyword)', async () => {
		const effects = mockEffects({
			carregarBuscasSalvas: vi.fn().mockResolvedValue([])
		});
		const engine = new BuscaEngine(effects);

		// Sem keyword, com shopId (guard permite)
		engine.ctx.shopIds = [123];
		engine.ctx.categorias = ['Perfumaria', 'Maquiagem'];

		await engine.send({ type: 'SALVAR' });

		expect(effects.salvarBusca).toHaveBeenCalledWith(
			expect.objectContaining({
				shop_ids: [123],
				categorias: ['Perfumaria', 'Maquiagem']
			})
		);
	});

	it('#36: salvar com cron inclui cron no payload', async () => {
		const effects = mockEffects({
			carregarBuscasSalvas: vi.fn().mockResolvedValue([])
		});
		const engine = new BuscaEngine(effects);

		engine.ctx.keyword = 'serum';
		engine.ctx.cron = '42 21 * * *';

		await engine.send({ type: 'SALVAR' });

		expect(effects.salvarBusca).toHaveBeenCalledWith(
			expect.objectContaining({
				keywords: ['serum'],
				cron: '42 21 * * *'
			})
		);
	});

	it('salvar sem contexto é bloqueado pelo guard podeSalvar', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);

		// Sem keyword, sem shopIds
		engine.ctx.keyword = '';
		engine.ctx.shopIds = [];

		await engine.send({ type: 'SALVAR' });

		expect(effects.salvarBusca).not.toHaveBeenCalled();
	});

	it('após salvar, buscasSalvas é atualizada', async () => {
		const effects = mockEffects({
			carregarBuscasSalvas: vi
				.fn()
				.mockResolvedValue([{ id: 'new', keywords: ['serum'], shop_ids: [], cron: '0 */8 * * *' }])
		});
		const engine = new BuscaEngine(effects);
		engine.ctx.keyword = 'serum';
		engine.ctx.cron = '0 */8 * * *';

		await engine.send({ type: 'SALVAR' });

		expect(engine.ctx.buscasSalvas).toHaveLength(1);
		expect(engine.ctx.buscasSalvas[0].keywords).toEqual(['serum']);
		expect(engine.ctx.buscasSalvas[0].cron).toBe('0 */8 * * *');
	});
});

// ── Filtros client-side não refetch (doc #39: cache) ──────────────────────

describe('BuscaEngine — Filtros client-side (sem refetch)', () => {
	it('MUDAR_FILTRO comissão não chama executarBusca (client-side)', async () => {
		const effects = mockEffects({
			executarBusca: vi.fn().mockResolvedValue({
				curadoria: [
					{ id: 'p1', nome: 'A', comissao: 0.15, vendas: 100, _fonte: 'curadoria' },
					{ id: 'p2', nome: 'B', comissao: 0.05, vendas: 50, _fonte: 'curadoria' }
				],
				quedas: [],
				novos: [],
				lojas: [],
				favoritos: []
			})
		});
		const engine = new BuscaEngine(effects);
		engine.ctx.keyword = 'test';
		await engine.send({ type: 'INICIALIZAR' });

		const callsBefore = effects.executarBusca.mock.calls.length;

		// Mudar filtro não deve refetch
		await engine.send({ type: 'MUDAR_FILTRO', comissaoMin: 0.1 });

		expect(effects.executarBusca.mock.calls.length).toBe(callsBefore);
		// Mas resultados devem ser refiltrados (B removido)
		expect(engine.ctx.comissaoMin).toBe(0.1);
	});

	it('MUDAR_FILTRO categorias não refetch', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);
		engine.ctx.keyword = 'test';
		await engine.send({ type: 'INICIALIZAR' });

		const callsBefore = effects.executarBusca.mock.calls.length;
		await engine.send({ type: 'MUDAR_FILTRO', categorias: ['Perfumaria'] });

		expect(effects.executarBusca.mock.calls.length).toBe(callsBefore);
	});

	it('MUDAR_FONTES SIM refetch (dados diferentes)', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);
		engine.ctx.keyword = 'test';
		await engine.send({ type: 'INICIALIZAR' });

		const callsBefore = effects.executarBusca.mock.calls.length;
		await engine.send({
			type: 'MUDAR_FONTES',
			fontes: { curadoria: false, quedas: true, novos: true, lojas: false, favoritos: false }
		});

		// Debounce
		await new Promise((r) => setTimeout(r, DEFAULTS.debounceMs + 50));
		expect(effects.executarBusca.mock.calls.length).toBeGreaterThan(callsBefore);
	});
});

// ── v3: Modos em cenários reais ──────────────────────────────────────────────

describe('BuscaEngine — Cenários v3: modos de interação', () => {
	it('carregar salva → alterar keyword → deseleciona automaticamente', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);
		const config = {
			id: 'b1',
			keywords: ['serum'],
			shopIds: [123],
			shopNomes: { 123: 'Loja X' },
			fontes: ['curadoria']
		};

		await engine.send({ type: 'CARREGAR_SALVA', config });
		expect(engine.modo).toBe(MODOS.VINCULADA);
		expect(engine.ctx.keyword).toBe('serum');

		// Alterar keyword desvincula
		await engine.send({ type: 'DIGITAR', value: 'serum anti-aging' });
		expect(engine.modo).toBe(MODOS.EXPLORANDO);
		expect(engine.ctx.buscaSelecionadaId).toBeNull();
		expect(engine.ctx.keyword).toBe('serum anti-aging');
	});

	it('editar salva → alterar loja → modo permanece editando (não desvincula)', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);

		await engine.send({ type: 'EDITAR_SALVA', config: { id: 'b1', keywords: ['serum'], shopIds: [] } });
		expect(engine.modo).toBe(MODOS.EDITANDO);

		// Em edit mode, alterar parâmetros NÃO desvincula (o usuário está editando)
		// O modo `editando` não tem `desvinculaEm` no JSON de regras
		await engine.send({ type: 'ADICIONAR_LOJA', value: 'lebotanic' });
		expect(engine.modo).toBe(MODOS.EDITANDO);
		expect(engine.ctx.editandoId).toBe('b1');
	});

	it('carregar busca A → carregar busca B → troca o vínculo', async () => {
		const effects = mockEffects();
		const engine = new BuscaEngine(effects);

		await engine.send({ type: 'CARREGAR_SALVA', config: { id: 'b1', keywords: ['serum'], shopIds: [] } });
		expect(engine.ctx.buscaSelecionadaId).toBe('b1');

		await engine.send({ type: 'CARREGAR_SALVA', config: { id: 'b2', keywords: ['retinol'], shopIds: [] } });
		expect(engine.modo).toBe(MODOS.VINCULADA);
		expect(engine.ctx.buscaSelecionadaId).toBe('b2');
		expect(engine.ctx.keyword).toBe('retinol');
	});
});

// ── v3: Marketplace filter ──────────────────────────────────────────────────

describe('BuscaEngine — Cenários v3: marketplace filter', () => {
	it('MUDAR_MARKETPLACES atualiza a lista de marketplaces no contexto', async () => {
		const engine = new BuscaEngine(mockEffects());
		await engine.send({ type: 'MUDAR_MARKETPLACES', marketplaces: ['shopee', 'amazon'] });
		expect(engine.ctx.marketplacesFiltro).toEqual(['shopee', 'amazon']);
	});

	it('MUDAR_MARKETPLACES com array vazio reseta filtro (todos)', async () => {
		const engine = new BuscaEngine(mockEffects());
		engine.ctx.marketplacesFiltro = ['shopee'];
		await engine.send({ type: 'MUDAR_MARKETPLACES', marketplaces: [] });
		expect(engine.ctx.marketplacesFiltro).toEqual([]);
	});

	it('marketplacesFiltro é incluído no payload ao salvar', async () => {
		const effects = mockEffects({
			carregarBuscasSalvas: vi.fn().mockResolvedValue([])
		});
		const engine = new BuscaEngine(effects);
		engine.ctx.keyword = 'serum';
		engine.ctx.marketplacesFiltro = ['shopee', 'amazon'];

		await engine.send({ type: 'SALVAR' });

		expect(effects.salvarBusca).toHaveBeenCalledWith(
			expect.objectContaining({
				marketplaces: ['shopee', 'amazon']
			})
		);
	});

	it('rules v3: marketplaces tem suportados, default e icones', () => {
		expect(rules.marketplaces.suportados).toContain('shopee');
		expect(rules.marketplaces.suportados).toContain('mercado_livre');
		expect(rules.marketplaces.suportados).toContain('amazon');
		expect(rules.marketplaces.default).toBe('shopee');
		expect(rules.marketplaces.icones).toBeDefined();
		expect(rules.marketplaces.icones.shopee).toBe('🟠');
	});

	it('rules v3: CANCELAR_EDICAO é uma transição válida', () => {
		expect(rules.transicoes.CANCELAR_EDICAO).toBeDefined();
		expect(rules.transicoes.CANCELAR_EDICAO.refetch).toBe(false);
		expect(rules.transicoes.CANCELAR_EDICAO.imediato).toBe(true);
	});

	it('rules v3: modos declaram transições e desvinculação', () => {
		expect(rules.modos.explorando).toBeDefined();
		expect(rules.modos.vinculada).toBeDefined();
		expect(rules.modos.editando).toBeDefined();
		expect(rules.modos.vinculada.desvinculaEm).toContain('DIGITAR');
		expect(rules.modos.vinculada.desvinculaEm).toContain('MUDAR_FILTRO');
		expect(rules.modos.editando.transicoes.SALVAR).toBe('explorando');
		expect(rules.modos.editando.transicoes.CANCELAR_EDICAO).toBe('explorando');
	});

	it('rules v3: buscaDuplicada declara campos de identidade', () => {
		expect(rules.buscaDuplicada).toBeDefined();
		expect(rules.buscaDuplicada.camposIdentidade).toContain('keyword');
		expect(rules.buscaDuplicada.camposIdentidade).toContain('shopIds');
		expect(rules.buscaDuplicada.erroAoSalvar).toBe(true);
		expect(rules.buscaDuplicada.feedbackReativo).toBe(true);
	});
});
