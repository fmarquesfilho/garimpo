/**
 * Parity test: verifica que normalizarNome (JS) produz output identico a Loja.Normalizar (C#).
 * Ambos leem o mesmo fixture: fixtures/normalizacao-pares.json.
 * Se este teste falha, há divergencia cross-language na busca local.
 */
import { describe, it, expect } from 'vitest';
import { normalizarNome } from '$lib/loja-registry.js';
import pares from '../../../fixtures/normalizacao-pares.json';

describe('normalizarNome — paridade com C# Loja.Normalizar', () => {
	it('fixture tem pares definidos', () => {
		expect(pares.length).toBeGreaterThan(0);
	});

	for (const { input, expected } of pares) {
		it(`"${input}" → "${expected}"`, () => {
			expect(normalizarNome(input)).toBe(expected);
		});
	}
});
