import { describe, it, expect } from 'vitest';

/**
 * Testes da lógica da página Descobrir.
 * Testa as funções de filtragem e montagem de resultados — executa em <1s.
 */

// ── Replica da lógica de montarResultados da +page.svelte ─────────────────

function montarResultados({ fontes, dadosCuradoria, dadosQuedas, dadosNovos, favoritos, busca }) {
	let todos = [];
	if (fontes.curadoria) todos.push(...dadosCuradoria);
	if (fontes.quedas) todos.push(...dadosQuedas);
	if (fontes.novos) todos.push(...dadosNovos);
	if (fontes.favoritos) {
		const favs = (favoritos ?? []).map(f => ({ ...f, id: f.produto_id, _fonte: 'favorito' }));
		todos.push(...favs);
	}

	const termo = (busca ?? '').trim().toLowerCase();
	if (termo) {
		todos = todos.filter(r =>
			(r.nome ?? '').toLowerCase().includes(termo) ||
			(r.loja ?? '').toLowerCase().includes(termo)
		);
	}

	return todos;
}

// ── Dados de teste ────────────────────────────────────────────────────────

const curadoria = [
	{ id: 'P1', nome: 'Sérum Vitamina C SKIN1004', preco: 89.9, loja: 'SKIN1004 Official', _fonte: 'curadoria' },
	{ id: 'P2', nome: 'Perfume Kenzo 50ml', preco: 299.9, loja: 'Perfumaria JP', _fonte: 'curadoria' }
];

const quedas = [
	{ id: 'V1', nome: 'Tônico COSRX', preco: 59.9, loja: 'COSRX Store', variacao_pct: -0.25, _fonte: 'queda' },
	{ id: 'V2', nome: 'Skin1004 Centella', preco: 95, loja: 'SKIN1004 Official', variacao_pct: -0.21, _fonte: 'queda' }
];

const novos = [
	{ id: 'N1', nome: 'Retinol Serum Novo', preco: 45.5, loja: 'SKIN1004 Official', _fonte: 'novo' }
];

const favoritos = [
	{ produto_id: 'F1', nome: 'Meu Favorito Perfume', preco: 150, loja: 'Loja ABC' }
];

// ── Cenários de fonte ─────────────────────────────────────────────────────

describe('Descobrir — Fontes de dados', () => {
	it('cenário 1: Busca com keyword retorna curadoria', () => {
		const r = montarResultados({ fontes: { curadoria: true, quedas: false, novos: false, favoritos: false }, dadosCuradoria: curadoria, dadosQuedas: [], dadosNovos: [], busca: 'sérum' });
		expect(r).toHaveLength(1);
		expect(r[0].nome).toContain('Sérum');
	});

	it('cenário 2: Busca sem keyword retorna vazio para curadoria', () => {
		const r = montarResultados({ fontes: { curadoria: true, quedas: false, novos: false, favoritos: false }, dadosCuradoria: [], dadosQuedas: [], dadosNovos: [], busca: '' });
		expect(r).toHaveLength(0);
	});

	it('cenário 3: Quedas sem keyword mostra todas as quedas', () => {
		const r = montarResultados({ fontes: { curadoria: false, quedas: true, novos: false, favoritos: false }, dadosCuradoria: [], dadosQuedas: quedas, dadosNovos: [], busca: '' });
		expect(r).toHaveLength(2);
	});

	it('cenário 4: Quedas com keyword filtra por nome', () => {
		const r = montarResultados({ fontes: { curadoria: false, quedas: true, novos: false, favoritos: false }, dadosCuradoria: [], dadosQuedas: quedas, dadosNovos: [], busca: 'Skin1004' });
		expect(r).toHaveLength(1);
		expect(r[0].nome).toBe('Skin1004 Centella');
	});

	it('cenário 5: Novos sem keyword mostra todos', () => {
		const r = montarResultados({ fontes: { curadoria: false, quedas: false, novos: true, favoritos: false }, dadosCuradoria: [], dadosQuedas: [], dadosNovos: novos, busca: '' });
		expect(r).toHaveLength(1);
	});

	it('cenário 6: Novos com keyword filtra', () => {
		const r = montarResultados({ fontes: { curadoria: false, quedas: false, novos: true, favoritos: false }, dadosCuradoria: [], dadosQuedas: [], dadosNovos: novos, busca: 'retinol' });
		expect(r).toHaveLength(1);
		expect(r[0].nome).toContain('Retinol');
	});

	it('cenário 7: Favoritos sem keyword mostra todos', () => {
		const r = montarResultados({ fontes: { curadoria: false, quedas: false, novos: false, favoritos: true }, dadosCuradoria: [], dadosQuedas: [], dadosNovos: [], favoritos, busca: '' });
		expect(r).toHaveLength(1);
	});

	it('cenário 8: Favoritos com keyword filtra', () => {
		const r = montarResultados({ fontes: { curadoria: false, quedas: false, novos: false, favoritos: true }, dadosCuradoria: [], dadosQuedas: [], dadosNovos: [], favoritos, busca: 'perfume' });
		expect(r).toHaveLength(1);
	});

	it('cenário 9: Nenhuma fonte ativa retorna vazio', () => {
		const r = montarResultados({ fontes: { curadoria: false, quedas: false, novos: false, favoritos: false }, dadosCuradoria: curadoria, dadosQuedas: quedas, dadosNovos: novos, favoritos, busca: '' });
		expect(r).toHaveLength(0);
	});

	it('cenário 10: Busca + Quedas + Novos sem keyword = só quedas e novos', () => {
		const r = montarResultados({ fontes: { curadoria: true, quedas: true, novos: true, favoritos: false }, dadosCuradoria: [], dadosQuedas: quedas, dadosNovos: novos, busca: '' });
		// curadoria retorna [] pois sem keyword não busca
		expect(r).toHaveLength(3); // 2 quedas + 1 novo
	});

	it('cenário 11: Todas as fontes com keyword filtra tudo', () => {
		const r = montarResultados({ fontes: { curadoria: true, quedas: true, novos: true, favoritos: true }, dadosCuradoria: curadoria, dadosQuedas: quedas, dadosNovos: novos, favoritos, busca: 'sérum' });
		// Sérum Vitamina C (curadoria) contém "sérum" — match
		// Retinol Serum Novo NÃO contém "sérum" (sem acento) — no match
		expect(r).toHaveLength(1);
		expect(r[0].nome).toContain('Sérum');
	});
});

// ── Filtragem por loja ────────────────────────────────────────────────────

describe('Descobrir — Busca por nome de loja', () => {
	it('cenário 13: keyword = nome da loja filtra produtos dessa loja', () => {
		const r = montarResultados({ fontes: { curadoria: true, quedas: true, novos: true, favoritos: false }, dadosCuradoria: curadoria, dadosQuedas: quedas, dadosNovos: novos, busca: 'SKIN1004' });
		// Sérum SKIN1004 (curadoria) + Skin1004 Centella (queda) + Retinol (novo, loja SKIN1004) = 3
		expect(r).toHaveLength(3);
		expect(r.every(p => p.nome.toLowerCase().includes('skin1004') || p.loja.toLowerCase().includes('skin1004'))).toBe(true);
	});

	it('cenário 14: keyword parcial filtra lojas com match parcial', () => {
		const r = montarResultados({ fontes: { curadoria: false, quedas: true, novos: false, favoritos: false }, dadosCuradoria: [], dadosQuedas: quedas, dadosNovos: [], busca: 'COSRX' });
		expect(r).toHaveLength(1);
		expect(r[0].loja).toBe('COSRX Store');
	});
});

// ── Combinações ───────────────────────────────────────────────────────────

describe('Descobrir — Combinações de fontes', () => {
	it('múltiplas fontes misturam resultados', () => {
		const r = montarResultados({ fontes: { curadoria: true, quedas: true, novos: false, favoritos: false }, dadosCuradoria: curadoria, dadosQuedas: quedas, dadosNovos: [], busca: '' });
		expect(r).toHaveLength(4); // 2 curadoria + 2 quedas
	});

	it('keyword vazia não filtra (mostra tudo das fontes ativas)', () => {
		const r = montarResultados({ fontes: { curadoria: true, quedas: true, novos: true, favoritos: true }, dadosCuradoria: curadoria, dadosQuedas: quedas, dadosNovos: novos, favoritos, busca: '' });
		expect(r).toHaveLength(6); // 2 + 2 + 1 + 1
	});

	it('keyword case-insensitive', () => {
		const r = montarResultados({ fontes: { curadoria: true, quedas: false, novos: false, favoritos: false }, dadosCuradoria: curadoria, dadosQuedas: [], dadosNovos: [], busca: 'KENZO' });
		expect(r).toHaveLength(1);
		expect(r[0].nome).toContain('Kenzo');
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
