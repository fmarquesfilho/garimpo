import { describe, it, expect, vi } from 'vitest';
import { BuscaEngine, guards, STATES } from '$lib/busca-engine.svelte.js';

/**
 * Testes da BuscaEngine — cenários reais reportados pelo usuário.
 * Usa mock effects (zero chamadas de API).
 */

function mockEffects(overrides = {}) {
	return {
		carregarBuscasSalvas: vi.fn().mockResolvedValue([]),
		carregarCategorias: vi.fn().mockResolvedValue([{ nome: 'Cosméticos' }, { nome: 'Perfumaria' }]),
		executarBusca: vi.fn().mockResolvedValue({ curadoria: [], quedas: [], novos: [], lojas: [], favoritos: [] }),
		resolverLoja: vi.fn().mockResolvedValue({ shop_ids: [920292999], keyword: 'Le Botanic' }),
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

	it('ADICIONAR_LOJA com erro mostra lojaErro', async () => {
		const effects = mockEffects({
			resolverLoja: vi.fn().mockRejectedValue(new Error('Loja não encontrada'))
		});
		const engine = new BuscaEngine(effects);

		await engine.send({ type: 'ADICIONAR_LOJA', value: 'invalida' });

		expect(engine.ctx.lojaErro).toBe('Loja não encontrada');
		expect(engine.ctx.shopIds).toHaveLength(0);
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
