/**
 * busca-engine-lojas — dominio de gerenciamento de lojas no escopo.
 *
 * Responsabilidade: adicionar/remover lojas do escopo da busca,
 * resolver lojas remotas (via Collector), match local no registro.
 *
 * Principio: funções recebem estado + effects e retornam mutacoes.
 * A engine aplica as mutacoes no $state reativo.
 */

import { normalizarNome, matchLojas } from './loja-registry.js';
import { guards } from './busca-engine-state.js';

/**
 * Tenta adicionar uma loja ao escopo. Decide o caminho:
 * - loja.id presente → adiciona direto (conhecida)
 * - input como texto → tenta match local, senao resolve remota
 *
 * @param {object} event - {loja?, value?}
 * @param {object} ctx - estado atual
 * @param {object[]} lojasDisponiveis - registro local
 * @returns {{tipo: 'conhecida', loja: object} | {tipo: 'resolver', input: string} | {tipo: 'noop'}}
 */
export function classificarAdicionarLoja(event, ctx, lojasDisponiveis) {
	if (event.loja?.id != null) {
		return { tipo: 'conhecida', loja: event.loja };
	}

	const input = (event.value ?? '').trim();
	if (!guards.lojaInputValida(ctx, event)) return { tipo: 'noop' };

	// Match local exato no registro
	const norm = normalizarNome(input);
	const [match] = matchLojas(norm, lojasDisponiveis, 1);
	if (match && norm === (match.nome_normalizado || normalizarNome(match.nome))) {
		return { tipo: 'conhecida', loja: match };
	}

	return { tipo: 'resolver', input };
}

/**
 * Prepara as mutacoes de ctx ao adicionar uma loja conhecida ao escopo.
 *
 * @param {object} loja - {id, nome, marketplace, origem, cron}
 * @param {object} ctx - estado atual
 * @returns {object|null} mutacoes para aplicar no ctx, ou null se ja existe
 */
export function prepararAdicionarLojaConhecida(loja, ctx) {
	const shopId = Number(loja.id);
	if (loja.id == null || Number.isNaN(shopId) || ctx.shopIds.includes(shopId)) return null;

	const { nome, marketplace, origem, cron } = loja;
	return {
		shopIds: [...ctx.shopIds, shopId],
		shopNomes: { ...ctx.shopNomes, [shopId]: nome },
		shopMeta: {
			...ctx.shopMeta,
			[shopId]: {
				marketplace: marketplace ?? 'shopee',
				origem: origem ?? null,
				monitorada: Boolean(cron),
				cron: cron ?? '',
				tipo: cron ? 'monitorada' : 'escopada'
			}
		}
	};
}

/**
 * Prepara as mutacoes de ctx ao remover uma loja do escopo.
 *
 * @param {number} shopId
 * @param {object} ctx
 * @returns {object} mutacoes
 */
export function prepararRemoverLoja(shopId, ctx) {
	const nomes = { ...ctx.shopNomes };
	delete nomes[shopId];
	const meta = { ...ctx.shopMeta };
	delete meta[shopId];
	return {
		shopIds: ctx.shopIds.filter((id) => id !== shopId),
		shopNomes: nomes,
		shopMeta: meta
	};
}
