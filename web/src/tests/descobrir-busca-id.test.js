import { describe, it, expect, vi } from 'vitest';

/**
 * Verifica que carregarOportunidades sempre envia busca.id (UUID) como buscaId,
 * nunca shop_ids[0] ou keywords[0].
 */

// Mock api.js
vi.mock('$lib/api.js', () => ({
	buscarCandidatos: vi.fn().mockResolvedValue({ candidatos: [] }),
	buscarNovidades: vi.fn().mockResolvedValue({
		produtos_novos: [],
		variacoes: [],
		total_novos: 0,
		total_variacoes: 0
	})
}));

import { carregarOportunidades } from '$lib/descobrir.js';
import { buscarNovidades } from '$lib/api.js';

describe('carregarOportunidades - buscaId resolution', () => {
	it('always passes busca.id as buscaId, not shop_ids[0]', async () => {
		const buscas = [
			{
				id: 'busca-loja-uuid-123',
				shop_ids: [920292999],
				keywords: ['serum'],
				shop_names: { '920292999': 'Glory of Seoul' }
			}
		];

		await carregarOportunidades(buscas, {});

		expect(buscarNovidades).toHaveBeenCalledWith({
			buscaId: 'busca-loja-uuid-123',
			dias: 7
		});
	});

	it('never uses shop_ids[0] even when present', async () => {
		const buscas = [
			{
				id: 'busca-specific-uuid',
				shop_ids: [111222333],
				keywords: [],
				shop_names: { '111222333': 'Test Shop' }
			}
		];

		buscarNovidades.mockClear();
		await carregarOportunidades(buscas, {});

		const call = buscarNovidades.mock.calls[0][0];
		expect(call.buscaId).toBe('busca-specific-uuid');
		expect(call.buscaId).not.toBe('111222333');
	});

	it('never uses keywords[0] even when present', async () => {
		const buscas = [
			{
				id: 'busca-kw-uuid',
				shop_ids: [],
				keywords: ['retinol'],
				shop_names: null
			}
		];

		buscarNovidades.mockClear();
		await carregarOportunidades(buscas, {});

		const call = buscarNovidades.mock.calls[0][0];
		expect(call.buscaId).toBe('busca-kw-uuid');
		expect(call.buscaId).not.toBe('retinol');
	});
});
