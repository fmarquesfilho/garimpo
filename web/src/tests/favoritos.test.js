import { describe, it, expect, vi, beforeEach } from 'vitest';

/**
 * Testes da store de favoritos — lógica de toggle, persistência local e dedup.
 * Cobre cenários 22, 23, 25 do TESTES_DESCOBRIR.
 */

// Mock das APIs
vi.mock('$lib/api.js', () => ({
	listarFavoritos: vi.fn().mockResolvedValue({ favoritos: [] }),
	salvarFavorito: vi.fn().mockResolvedValue({}),
	removerFavorito: vi.fn().mockResolvedValue({})
}));

// Mock localStorage
const storage = {};
vi.stubGlobal('localStorage', {
	getItem: (k) => storage[k] ?? null,
	setItem: (k, v) => {
		storage[k] = v;
	},
	removeItem: (k) => delete storage[k]
});

// Import depois dos mocks
const { favoritos } = await import('$lib/favoritos.js');
const { get } = await import('svelte/store');

const produtoBase = {
	id: 'P1',
	produto_id: 'P1',
	nome: 'Sérum Vitamina C',
	preco: 89.9,
	comissao: 0.12,
	link: 'https://shopee.com/P1',
	imagem: 'https://img.com/p1.jpg',
	loja: 'SKIN1004',
	categoria: 'Cuidados com a Pele',
	origem: 'Coreia'
};

describe('Favoritos — toggle (cenários 22-23)', () => {
	beforeEach(() => {
		// Limpar favoritos
		favoritos.subscribe(() => {})(); // subscribe para ativar
		// Reset via API mock
		const lista = get(favoritos);
		for (const f of lista) {
			favoritos.remover(f.produto_id || f.id);
		}
	});

	it('cenário 22: toggle adiciona produto aos favoritos', async () => {
		expect(favoritos.isFavorito('P1')).toBe(false);

		await favoritos.toggle(produtoBase);

		expect(favoritos.isFavorito('P1')).toBe(true);
		const lista = get(favoritos);
		expect(lista).toHaveLength(1);
		expect(lista[0].nome).toBe('Sérum Vitamina C');
	});

	it('cenário 23: toggle remove produto já favoritado', async () => {
		await favoritos.adicionar(produtoBase);
		expect(favoritos.isFavorito('P1')).toBe(true);

		await favoritos.toggle(produtoBase);

		expect(favoritos.isFavorito('P1')).toBe(false);
		const lista = get(favoritos);
		expect(lista.find((f) => f.produto_id === 'P1')).toBeUndefined();
	});

	it('adicionar o mesmo produto duas vezes não duplica', async () => {
		await favoritos.adicionar(produtoBase);
		await favoritos.adicionar(produtoBase);

		const lista = get(favoritos);
		expect(lista.filter((f) => f.produto_id === 'P1')).toHaveLength(1);
	});

	it('isFavorito funciona com id ou produto_id', async () => {
		await favoritos.adicionar({ ...produtoBase, produto_id: 'X1', id: 'X1' });

		expect(favoritos.isFavorito('X1')).toBe(true);
		expect(favoritos.isFavorito('inexistente')).toBe(false);
	});
});

describe('Favoritos — cross-fonte (cenário 25)', () => {
	it('produto favoritado de Quedas aparece na lista de favoritos', async () => {
		const produtoQueda = {
			id: 'V1',
			produto_id: 'V1',
			nome: 'Tônico COSRX Queda',
			preco: 49.9,
			comissao: 0.1,
			loja: 'COSRX Store',
			_fonte: 'queda',
			variacao_pct: -0.25
		};

		await favoritos.adicionar(produtoQueda);

		expect(favoritos.isFavorito('V1')).toBe(true);
		const lista = get(favoritos);
		expect(lista[0].nome).toBe('Tônico COSRX Queda');
		expect(lista[0].produto_id).toBe('V1');
	});
});
