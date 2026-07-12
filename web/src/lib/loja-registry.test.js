import { describe, it, expect } from 'vitest';
import { normalizarNome, matchLojas } from './loja-registry.js';
import pares from '../../../fixtures/normalizacao-pares.json';

describe('loja-registry.js', () => {
	describe('normalizarNome', () => {
		it('normaliza usando fixtures parametrizadas', () => {
			for (const par of pares) {
				expect(normalizarNome(par.input)).toBe(par.expected);
			}
		});
	});

	describe('matchLojas', () => {
		const lojas = [
			{ id: '1', nome: 'Glory of Seoul', nome_normalizado: 'gloryofseoul', marketplace: 'shopee' },
			{ id: '2', nome: 'Le Botanic', nome_normalizado: 'lebotanic', marketplace: 'shopee' }
		];

		it('retorna vazio para input curto', () => {
			expect(matchLojas('g', lojas)).toEqual([]);
			expect(matchLojas('', lojas)).toEqual([]);
		});

		it('faz match pelo nome normalizado', () => {
			expect(matchLojas('glory', lojas)).toEqual([lojas[0]]);
			expect(matchLojas('gloryofseoul', lojas)).toEqual([lojas[0]]);
			expect(matchLojas('lebot', lojas)).toEqual([lojas[1]]);
		});

		it('faz match pelo nome original (substring lowercase)', () => {
			// "of seoul" não vai casar com o normalizado que é "gloryofseoul",
			// mas deve casar com o nome original em lowercase "glory of seoul".
			expect(matchLojas('of seoul', lojas)).toEqual([lojas[0]]);
		});
	});
});
