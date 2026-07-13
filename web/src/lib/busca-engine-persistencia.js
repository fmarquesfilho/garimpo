/**
 * busca-engine-persistencia — dominio de buscas salvas.
 *
 * Responsabilidade: restaurar config, montar payload para salvar,
 * validar duplicatas. Funções puras — sem estado reativo.
 */

import { configToPayload, gerarLabelBusca } from './busca-unificada-logic.js';
import { BUSCA_DUPLICADA } from './busca-config.js';
import { guards } from './busca-engine-state.js';

/**
 * Restaura campos do contexto a partir de uma config de busca salva.
 * Retorna objeto parcial para aplicar via Object.assign(ctx, ...).
 *
 * @param {object} config
 * @returns {object}
 */
export function restaurarConfig(config) {
	const resultado = {
		keyword: (config.keywords ?? [])[0] ?? '',
		shopIds: config.shopIds ?? [],
		shopNomes: config.shopNomes ?? {},
		comissaoMin: config.comissaoMin || 0.07,
		vendasMin: config.vendasMin || 0,
		categorias: config.categorias ?? [],
		marketplacesFiltro: Array.isArray(config.marketplaces) ? config.marketplaces : [],
		cron: config.cron ?? ''
	};
	if (config.fontes?.length) {
		resultado.fontes = {
			curadoria: config.fontes.includes('curadoria'),
			quedas: config.fontes.includes('quedas'),
			novos: config.fontes.includes('novos'),
			lojas: config.fontes.includes('lojas'),
			favoritos: config.fontes.includes('favoritos')
		};
	}
	return resultado;
}

/**
 * Monta o payload para enviar ao servidor ao salvar.
 *
 * @param {object} ctx
 * @param {string[]} fontesAtivas
 * @returns {object}
 */
export function montarPayloadSalvar(ctx, fontesAtivas) {
	return configToPayload({
		id: ctx.editandoId ?? undefined,
		keywords: ctx.keyword ? [ctx.keyword] : [],
		shopIds: ctx.shopIds,
		shopNomes: ctx.shopNomes,
		comissaoMin: ctx.comissaoMin,
		vendasMin: ctx.vendasMin,
		categorias: ctx.categorias,
		fontes: fontesAtivas,
		cron: ctx.cron,
		marketplaces: ctx.marketplacesFiltro.length ? ctx.marketplacesFiltro : 'shopee'
	});
}

/**
 * Valida se a busca pode ser salva (guards + duplicata).
 *
 * @param {object} ctx
 * @returns {{ok: boolean, erro?: string}}
 */
export function validarAntesDeSalvar(ctx) {
	if (!guards.podeSalvar(ctx)) return { ok: false };
	if (BUSCA_DUPLICADA.erroAoSalvar) {
		const dup = guards.buscaDuplicada(ctx);
		if (dup) {
			return { ok: false, erro: `Ja existe uma busca salva com os mesmos parametros: "${gerarLabelBusca(dup)}"` };
		}
	}
	return { ok: true };
}
