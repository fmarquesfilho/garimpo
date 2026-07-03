import { describe, it, expect } from 'vitest';
import { montarResultados, encontrarLojaPorNome } from '$lib/descobrir-logic.js';

/**
 * Testes da lógica da página Descobrir.
 * Testa as funções de filtragem e montagem de resultados — executa em <1s.
 */

// ── Dados de teste ────────────────────────────────────────────────────────

const curadoria = [
	{ id: 'P1', nome: 'Sérum Vitamina C SKIN1004', preco: 89.9, loja: 'SKIN1004 Official', _fonte: 'curadoria' },
	{ id: 'P2', nome: 'Perfume Kenzo 50ml', preco: 299.9, loja: 'Perfumaria JP', _fonte: 'curadoria' }
];

const quedas = [
	{ id: 'V1', nome: 'Tônico COSRX', preco: 59.9, loja: 'COSRX Store', variacao_pct: -0.25, _fonte: 'queda' },
	{ id: 'V2', nome: 'Skin1004 Centella', preco: 95, loja: 'SKIN1004 Official', variacao_pct: -0.21, _fonte: 'queda' }
];

const novos = [{ id: 'N1', nome: 'Retinol Serum Novo', preco: 45.5, loja: 'SKIN1004 Official', _fonte: 'novo' }];

const favoritos = [{ produto_id: 'F1', nome: 'Meu Favorito Perfume', preco: 150, loja: 'Loja ABC' }];

// ── Cenários de fonte ─────────────────────────────────────────────────────

// eslint-disable-next-line max-lines-per-function
describe('Descobrir — Fontes de dados', () => {
	it('cenário 1: Busca com keyword retorna curadoria', () => {
		const r = montarResultados({
			fontes: { curadoria: true, quedas: false, novos: false, favoritos: false },
			dadosCuradoria: curadoria,
			dadosQuedas: [],
			dadosNovos: [],
			busca: 'sérum'
		});
		expect(r).toHaveLength(1);
		expect(r[0].nome).toContain('Sérum');
	});

	it('cenário 2: Busca sem keyword retorna vazio para curadoria', () => {
		const r = montarResultados({
			fontes: { curadoria: true, quedas: false, novos: false, favoritos: false },
			dadosCuradoria: [],
			dadosQuedas: [],
			dadosNovos: [],
			busca: ''
		});
		expect(r).toHaveLength(0);
	});

	it('cenário 3: Quedas sem keyword mostra todas as quedas', () => {
		const r = montarResultados({
			fontes: { curadoria: false, quedas: true, novos: false, favoritos: false },
			dadosCuradoria: [],
			dadosQuedas: quedas,
			dadosNovos: [],
			busca: ''
		});
		expect(r).toHaveLength(2);
	});

	it('cenário 4: Quedas com keyword filtra por nome', () => {
		const r = montarResultados({
			fontes: { curadoria: false, quedas: true, novos: false, favoritos: false },
			dadosCuradoria: [],
			dadosQuedas: quedas,
			dadosNovos: [],
			busca: 'Skin1004'
		});
		expect(r).toHaveLength(1);
		expect(r[0].nome).toBe('Skin1004 Centella');
	});

	it('cenário 5: Novos sem keyword mostra todos', () => {
		const r = montarResultados({
			fontes: { curadoria: false, quedas: false, novos: true, favoritos: false },
			dadosCuradoria: [],
			dadosQuedas: [],
			dadosNovos: novos,
			busca: ''
		});
		expect(r).toHaveLength(1);
	});

	it('cenário 6: Novos com keyword filtra', () => {
		const r = montarResultados({
			fontes: { curadoria: false, quedas: false, novos: true, favoritos: false },
			dadosCuradoria: [],
			dadosQuedas: [],
			dadosNovos: novos,
			busca: 'retinol'
		});
		expect(r).toHaveLength(1);
		expect(r[0].nome).toContain('Retinol');
	});

	it('cenário 7: Favoritos sem keyword mostra todos', () => {
		const r = montarResultados({
			fontes: { curadoria: false, quedas: false, novos: false, favoritos: true },
			dadosCuradoria: [],
			dadosQuedas: [],
			dadosNovos: [],
			favoritos,
			busca: ''
		});
		expect(r).toHaveLength(1);
	});

	it('cenário 8: Favoritos com keyword filtra', () => {
		const r = montarResultados({
			fontes: { curadoria: false, quedas: false, novos: false, favoritos: true },
			dadosCuradoria: [],
			dadosQuedas: [],
			dadosNovos: [],
			favoritos,
			busca: 'perfume'
		});
		expect(r).toHaveLength(1);
	});

	it('cenário 9: Nenhuma fonte ativa retorna vazio', () => {
		const r = montarResultados({
			fontes: { curadoria: false, quedas: false, novos: false, favoritos: false },
			dadosCuradoria: curadoria,
			dadosQuedas: quedas,
			dadosNovos: novos,
			favoritos,
			busca: ''
		});
		expect(r).toHaveLength(0);
	});

	it('cenário 10: Busca + Quedas + Novos sem keyword = só quedas e novos', () => {
		const r = montarResultados({
			fontes: { curadoria: true, quedas: true, novos: true, favoritos: false },
			dadosCuradoria: [],
			dadosQuedas: quedas,
			dadosNovos: novos,
			busca: ''
		});
		// curadoria retorna [] pois sem keyword não busca
		expect(r).toHaveLength(3); // 2 quedas + 1 novo
	});

	it('cenário 11: Todas as fontes com keyword filtra tudo', () => {
		const r = montarResultados({
			fontes: { curadoria: true, quedas: true, novos: true, favoritos: true },
			dadosCuradoria: curadoria,
			dadosQuedas: quedas,
			dadosNovos: novos,
			favoritos,
			busca: 'sérum'
		});
		// Sérum Vitamina C (curadoria) contém "sérum" — match
		// Retinol Serum Novo NÃO contém "sérum" (sem acento) — no match
		expect(r).toHaveLength(1);
		expect(r[0].nome).toContain('Sérum');
	});
});

// ── Filtragem por loja ────────────────────────────────────────────────────

describe('Descobrir — Busca por nome de loja', () => {
	it('cenário 13: keyword = nome da loja filtra produtos dessa loja', () => {
		const r = montarResultados({
			fontes: { curadoria: true, quedas: true, novos: true, favoritos: false },
			dadosCuradoria: curadoria,
			dadosQuedas: quedas,
			dadosNovos: novos,
			busca: 'SKIN1004'
		});
		// Sérum SKIN1004 (curadoria) + Skin1004 Centella (queda) + Retinol (novo, loja SKIN1004) = 3
		expect(r).toHaveLength(3);
		expect(r.every((p) => p.nome.toLowerCase().includes('skin1004') || p.loja.toLowerCase().includes('skin1004'))).toBe(
			true
		);
	});

	it('cenário 14: keyword parcial filtra lojas com match parcial', () => {
		const r = montarResultados({
			fontes: { curadoria: false, quedas: true, novos: false, favoritos: false },
			dadosCuradoria: [],
			dadosQuedas: quedas,
			dadosNovos: [],
			busca: 'COSRX'
		});
		expect(r).toHaveLength(1);
		expect(r[0].loja).toBe('COSRX Store');
	});
});

// ── Combinações ───────────────────────────────────────────────────────────

describe('Descobrir — Combinações de fontes', () => {
	it('múltiplas fontes misturam resultados', () => {
		const r = montarResultados({
			fontes: { curadoria: true, quedas: true, novos: false, favoritos: false },
			dadosCuradoria: curadoria,
			dadosQuedas: quedas,
			dadosNovos: [],
			busca: ''
		});
		expect(r).toHaveLength(4); // 2 curadoria + 2 quedas
	});

	it('keyword vazia não filtra (mostra tudo das fontes ativas)', () => {
		const r = montarResultados({
			fontes: { curadoria: true, quedas: true, novos: true, favoritos: true },
			dadosCuradoria: curadoria,
			dadosQuedas: quedas,
			dadosNovos: novos,
			favoritos,
			busca: ''
		});
		expect(r).toHaveLength(6); // 2 + 2 + 1 + 1
	});

	it('keyword case-insensitive', () => {
		const r = montarResultados({
			fontes: { curadoria: true, quedas: false, novos: false, favoritos: false },
			dadosCuradoria: curadoria,
			dadosQuedas: [],
			dadosNovos: [],
			busca: 'KENZO'
		});
		expect(r).toHaveLength(1);
		expect(r[0].nome).toContain('Kenzo');
	});
});

// ── Filtragem por categoria ────────────────────────────────────────────────

describe('Descobrir — Filtragem por categoria', () => {
	const comCategoria = [
		{ id: 'P1', nome: 'Sérum Vitamina C', categoria: 'Cuidados com a Pele', loja: 'SKIN1004', _fonte: 'curadoria' },
		{ id: 'P2', nome: 'Perfume Kenzo', categoria: 'Perfumaria', loja: 'Loja X', _fonte: 'curadoria' },
		{ id: 'P3', nome: 'Batom Matte', categoria: 'Maquiagem', loja: 'Loja Y', _fonte: 'curadoria' },
		{ id: 'P4', nome: 'Tônico sem categoria', categoria: '', loja: 'Loja Z', _fonte: 'queda' }
	];

	function montarComCategoria(cats) {
		return montarResultados({
			fontes: { curadoria: true, quedas: true, novos: false, favoritos: false },
			dadosCuradoria: comCategoria.filter((c) => c._fonte === 'curadoria'),
			dadosQuedas: comCategoria.filter((c) => c._fonte === 'queda'),
			dadosNovos: [],
			busca: '',
			categorias: cats
		});
	}

	it('cenário 14: categoria única filtra só produtos daquela categoria', () => {
		const r = montarComCategoria(['Perfumaria']);
		expect(r).toHaveLength(2); // Perfume Kenzo + Tônico sem categoria (sem categoria passa)
		expect(r.some((p) => p.nome === 'Perfume Kenzo')).toBe(true);
	});

	it('cenário 15: múltiplas categorias filtra OR', () => {
		const r = montarComCategoria(['Perfumaria', 'Maquiagem']);
		expect(r).toHaveLength(3); // Perfume + Batom + Tônico sem categoria
	});

	it('sem categorias selecionadas mostra tudo', () => {
		const r = montarComCategoria([]);
		expect(r).toHaveLength(4);
	});

	it('categoria case-insensitive', () => {
		const r = montarComCategoria(['perfumaria']);
		expect(r.some((p) => p.nome === 'Perfume Kenzo')).toBe(true);
	});

	it('cenário 16: keyword + categoria combinam (AND)', () => {
		const r = montarResultados({
			fontes: { curadoria: true, quedas: false, novos: false, favoritos: false },
			dadosCuradoria: comCategoria.filter((c) => c._fonte === 'curadoria'),
			dadosQuedas: [],
			dadosNovos: [],
			busca: 'Kenzo',
			categorias: ['Perfumaria']
		});
		expect(r).toHaveLength(1);
		expect(r[0].nome).toBe('Perfume Kenzo');
	});

	it('cenário 17: keyword + categoria + loja filtra interseção', () => {
		const r = montarResultados({
			fontes: { curadoria: true, quedas: false, novos: false, favoritos: false },
			dadosCuradoria: comCategoria.filter((c) => c._fonte === 'curadoria'),
			dadosQuedas: [],
			dadosNovos: [],
			busca: 'SKIN1004',
			categorias: ['Cuidados com a Pele']
		});
		expect(r).toHaveLength(1);
		expect(r[0].nome).toBe('Sérum Vitamina C');
	});
});

// ── Buscas salvas — aplicação de fontes ───────────────────────────────────

describe('Descobrir — Aplicar busca salva', () => {
	function aplicarBuscaSalva(b) {
		const kw = (b.keywords ?? [])[0] ?? '';
		const fontes = { curadoria: false, quedas: false, novos: false, favoritos: false };
		if (b.fontes?.length) {
			fontes.curadoria = b.fontes.includes('curadoria');
			fontes.quedas = b.fontes.includes('quedas');
			fontes.novos = b.fontes.includes('novos');
			fontes.favoritos = b.fontes.includes('favoritos');
		} else {
			fontes.curadoria = kw.length > 0;
		}
		return { busca: kw, fontes };
	}

	it('cenário 19: busca com keyword ativa curadoria', () => {
		const { busca, fontes } = aplicarBuscaSalva({ keywords: ['sérum'], fontes: ['curadoria'] });
		expect(busca).toBe('sérum');
		expect(fontes.curadoria).toBe(true);
		expect(fontes.quedas).toBe(false);
	});

	it('cenário 20: busca com fontes [quedas, novos] ativa ambas', () => {
		const { fontes } = aplicarBuscaSalva({ keywords: [], fontes: ['quedas', 'novos'] });
		expect(fontes.quedas).toBe(true);
		expect(fontes.novos).toBe(true);
		expect(fontes.curadoria).toBe(false);
	});

	it('busca sem fontes explícitas usa keyword para inferir', () => {
		const { busca, fontes } = aplicarBuscaSalva({ keywords: ['perfume'] });
		expect(busca).toBe('perfume');
		expect(fontes.curadoria).toBe(true);
	});

	it('busca sem keyword e sem fontes desativa tudo', () => {
		const { fontes } = aplicarBuscaSalva({ keywords: [] });
		expect(fontes.curadoria).toBe(false);
		expect(fontes.quedas).toBe(false);
	});
});

// ── Filtros numéricos (comissão, vendas, nota) ────────────────────────────

describe('Descobrir — Filtros numéricos', () => {
	const produtos = [
		{
			id: 'A',
			nome: 'Sérum Premium',
			comissao: 0.15,
			vendas: 200,
			avaliacao: 4.9,
			loja: 'Loja A',
			_fonte: 'curadoria'
		},
		{ id: 'B', nome: 'Creme Básico', comissao: 0.05, vendas: 50, avaliacao: 3.8, loja: 'Loja B', _fonte: 'curadoria' },
		{ id: 'C', nome: 'Perfume Médio', comissao: 0.1, vendas: 100, avaliacao: 4.5, loja: 'Loja C', _fonte: 'curadoria' },
		{ id: 'D', nome: 'Queda sem comissao', comissao: 0, vendas: 0, avaliacao: 0, loja: 'Loja D', _fonte: 'queda' }
	];

	const base = {
		fontes: { curadoria: true, quedas: true, novos: false, favoritos: false },
		dadosCuradoria: produtos.filter((p) => p._fonte === 'curadoria'),
		dadosQuedas: produtos.filter((p) => p._fonte === 'queda'),
		dadosNovos: [],
		busca: ''
	};

	it('comissaoMin=0.10 filtra produtos com comissão < 10%', () => {
		const r = montarResultados({ ...base, comissaoMin: 0.1 });
		// A (15%), C (10%) passam; B (5%) não; D (0, sem comissão) passa
		expect(r).toHaveLength(3);
		expect(r.find((p) => p.id === 'B')).toBeUndefined();
	});

	it('comissaoMin=0.15 filtra tudo exceto 15%+', () => {
		const r = montarResultados({ ...base, comissaoMin: 0.15 });
		// A (15%) + D (sem comissão, passa)
		expect(r).toHaveLength(2);
		expect(r.find((p) => p.id === 'A')).toBeDefined();
	});

	it('vendasMin=100 filtra produtos com menos vendas', () => {
		const r = montarResultados({ ...base, vendasMin: 100 });
		// A (200), C (100) passam; B (50) não; D (0, sem vendas) passa
		expect(r).toHaveLength(3);
		expect(r.find((p) => p.id === 'B')).toBeUndefined();
	});

	it('filtros combinam com keyword (AND)', () => {
		const r = montarResultados({ ...base, busca: 'Sérum', comissaoMin: 0.1 });
		expect(r).toHaveLength(1);
		expect(r[0].nome).toBe('Sérum Premium');
	});

	it('sem filtros numéricos mostra tudo', () => {
		const r = montarResultados({ ...base, comissaoMin: 0, vendasMin: 0 });
		expect(r).toHaveLength(4);
	});

	it('produto sem dados numéricos não é filtrado (graceful)', () => {
		const r = montarResultados({ ...base, comissaoMin: 0.1, vendasMin: 50 });
		// D tem comissao=0, vendas=0 — não tem dados, não é filtrado
		expect(r.find((p) => p.id === 'D')).toBeDefined();
	});
});

// ── Todos os filtros combinados (cenários completos) ──────────────────────

const todosProdutos = [
	{
		id: 'A',
		nome: 'Sérum SKIN1004',
		preco: 89.9,
		comissao: 0.15,
		vendas: 200,
		avaliacao: 4.9,
		loja: 'SKIN1004 Official',
		categoria: 'Cuidados com a Pele',
		_fonte: 'curadoria'
	},
	{
		id: 'B',
		nome: 'Perfume Kenzo',
		preco: 299.9,
		comissao: 0.08,
		vendas: 80,
		avaliacao: 4.6,
		loja: 'Perfumaria JP',
		categoria: 'Perfumaria',
		_fonte: 'curadoria'
	},
	{
		id: 'C',
		nome: 'Batom Barato',
		preco: 15,
		comissao: 0.03,
		vendas: 5,
		avaliacao: 3.2,
		loja: 'Loja X',
		categoria: 'Maquiagem',
		_fonte: 'curadoria'
	},
	{
		id: 'D',
		nome: 'Tônico COSRX Queda',
		preco: 59.9,
		comissao: 0.12,
		vendas: 150,
		avaliacao: 4.7,
		loja: 'COSRX Store',
		categoria: 'Cuidados com a Pele',
		variacao_pct: -0.25,
		_fonte: 'queda'
	},
	{
		id: 'E',
		nome: 'Retinol Novo SKIN1004',
		preco: 45.5,
		comissao: 0.1,
		vendas: 0,
		avaliacao: 0,
		loja: 'SKIN1004 Official',
		categoria: 'Cuidados com a Pele',
		_fonte: 'novo'
	},
	{
		id: 'F',
		nome: 'Meu Favorito Perfume',
		preco: 150,
		comissao: 0.09,
		vendas: 60,
		avaliacao: 4.4,
		loja: 'Loja Y',
		categoria: 'Perfumaria',
		_fonte: 'favorito'
	}
];

const allFontes = { curadoria: true, quedas: true, novos: true, favoritos: true };

function filtrar(opts) {
	return montarResultados({
		fontes: opts.fontes ?? allFontes,
		dadosCuradoria: todosProdutos.filter((p) => p._fonte === 'curadoria'),
		dadosQuedas: todosProdutos.filter((p) => p._fonte === 'queda'),
		dadosNovos: todosProdutos.filter((p) => p._fonte === 'novo'),
		favoritos: todosProdutos.filter((p) => p._fonte === 'favorito').map((p) => ({ ...p, produto_id: p.id })),
		busca: opts.busca ?? '',
		categorias: opts.categorias ?? [],
		comissaoMin: opts.comissaoMin ?? 0,
		vendasMin: opts.vendasMin ?? 0
	});
}

describe('Descobrir — Filtros combinados (básico)', () => {
	it('sem filtros mostra todos (6 produtos)', () => {
		expect(filtrar({})).toHaveLength(6);
	});

	it('keyword "SKIN1004" filtra por nome e loja', () => {
		const r = filtrar({ busca: 'SKIN1004' });
		expect(r).toHaveLength(2);
	});

	it('keyword + comissaoMin combina (AND)', () => {
		const r = filtrar({ busca: 'SKIN1004', comissaoMin: 0.12 });
		expect(r).toHaveLength(1);
		expect(r[0].id).toBe('A');
	});

	it('categoria "Perfumaria" + todas as fontes', () => {
		const r = filtrar({ categorias: ['Perfumaria'] });
		expect(r).toHaveLength(2);
		expect(r.every((p) => p.categoria === 'Perfumaria')).toBe(true);
	});

	it('vendasMin=100 remove produtos com poucas vendas', () => {
		const r = filtrar({ vendasMin: 100 });
		expect(r).toHaveLength(3);
		expect(r.map((p) => p.id).sort()).toEqual(['A', 'D', 'E']);
	});
});

describe('Descobrir — Filtros combinados (avançado)', () => {
	it('categoria + keyword + comissaoMin', () => {
		const r = filtrar({ categorias: ['Cuidados com a Pele'], busca: 'SKIN1004', comissaoMin: 0.12 });
		expect(r).toHaveLength(1);
		expect(r[0].id).toBe('A');
	});

	it('vendasMin=100 + categoria "Cuidados com a Pele"', () => {
		const r = filtrar({ vendasMin: 100, categorias: ['Cuidados com a Pele'] });
		expect(r).toHaveLength(3);
	});

	it('todos os filtros ao mesmo tempo (cenário máximo)', () => {
		const r = filtrar({
			busca: 'SKIN1004',
			categorias: ['Cuidados com a Pele'],
			comissaoMin: 0.1,
			vendasMin: 50
		});
		expect(r).toHaveLength(2);
	});

	it('só fonte Quedas + comissaoMin', () => {
		const r = filtrar({ fontes: { curadoria: false, quedas: true, novos: false, favoritos: false }, comissaoMin: 0.1 });
		expect(r).toHaveLength(1);
		expect(r[0].id).toBe('D');
	});

	it('só fonte Favoritos + keyword', () => {
		const r = filtrar({ fontes: { curadoria: false, quedas: false, novos: false, favoritos: true }, busca: 'perfume' });
		expect(r).toHaveLength(1);
		expect(r[0].id).toBe('F');
	});

	it('fonte Novos + categoria', () => {
		const r = filtrar({
			fontes: { curadoria: false, quedas: false, novos: true, favoritos: false },
			categorias: ['Cuidados com a Pele']
		});
		expect(r).toHaveLength(1);
		expect(r[0].id).toBe('E');
	});

	it('nenhuma fonte ativa retorna vazio independente dos filtros', () => {
		const r = filtrar({
			fontes: { curadoria: false, quedas: false, novos: false, favoritos: false },
			busca: 'SKIN1004'
		});
		expect(r).toHaveLength(0);
	});
});

// ── Detecção de loja por nome ─────────────────────────────────────────────

describe('Descobrir — Detecção de loja por nome', () => {
	const lojas = [
		{ id: 'loja-123', nome: 'SKIN1004 Official', shop_ids: [123] },
		{ id: 'loja-456', nome: 'Belezura Distribuidora', shop_ids: [456] },
		{ id: 'loja-789', nome: 'COSRX Store', shop_ids: [789] }
	];

	it('encontra loja por nome exato', () => {
		const r = encontrarLojaPorNome('Belezura Distribuidora', lojas);
		expect(r).not.toBeNull();
		expect(r.shop_ids).toEqual([456]);
	});

	it('encontra loja por parte do nome (case-insensitive)', () => {
		const r = encontrarLojaPorNome('belezura', lojas);
		expect(r).not.toBeNull();
		expect(r.id).toBe('loja-456');
	});

	it('encontra loja por match parcial bidirecional', () => {
		const r = encontrarLojaPorNome('SKIN1004', lojas);
		expect(r).not.toBeNull();
		expect(r.id).toBe('loja-123');
	});

	it('retorna null se não encontrar', () => {
		expect(encontrarLojaPorNome('inexistente', lojas)).toBeNull();
	});

	it('retorna null se termo vazio', () => {
		expect(encontrarLojaPorNome('', lojas)).toBeNull();
	});
});
