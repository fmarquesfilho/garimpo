/**
 * Efeitos colaterais da BuscaEngine — chamadas de API isoladas.
 * Injetáveis: para testes, substituir por mocks.
 *
 * @typedef {ReturnType<typeof criarEffects>} Effects
 */

import { adicionarLoja, sincronizarBusca, listarBuscasServidor } from './api.js';
import { buscarCategorias } from './categorias.js';
import { carregarCuradoria, carregarOportunidades, carregarProdutosLojas } from './descobrir.js';
import { agruparCategoriasPorMarketplace, listarLojasMonitoradas } from './descobrir-logic.js';

/**
 * Cria effects concretos para produção.
 * @param {object} deps — dependências reativas (stores)
 * @param {() => any[]} deps.getBuscasSalvas — getter reativo de buscas do store
 * @param {() => any[]} deps.getFavoritos — getter reativo de favoritos do store
 * @param {Function} [deps.sincronizarStore] - sync externo após salvar/remover
 */
export function criarEffects({ getBuscasSalvas, getFavoritos, sincronizarStore }) {
	return {
		/** Carrega buscas salvas do servidor. */
		async carregarBuscasSalvas() {
			const r = await listarBuscasServidor();
			return r?.buscas ?? [];
		},

		/** Sincroniza o store externo de buscas com o servidor (para que executarBusca veja dados frescos). */
		async sincronizarStoreExterno() {
			if (sincronizarStore) await sincronizarStore();
		},

		/** Carrega categorias agrupadas por marketplace para autocomplete (`{ nome, marketplaces[] }[]`). */
		async carregarCategorias() {
			const cruas = await buscarCategorias();
			return agruparCategoriasPorMarketplace(cruas);
		},

		/** Lista lojas monitoradas (deriva das buscas salvas) para o autocomplete da raia Lojas. */
		listarLojasMonitoradas() {
			return listarLojasMonitoradas(getBuscasSalvas());
		},

		/** Executa busca em todas as fontes ativas. */
		async executarBusca(ctx) {
			const buscasSalvas = getBuscasSalvas();
			const { buscasComLojas, nomesLojas } = buildBuscasComLojas(buscasSalvas, ctx);

			const resultado = { curadoria: [], quedas: [], novos: [], lojas: [], favoritos: getFavoritos() ?? [] };
			const promises = [];

			if (ctx.fontes.curadoria && (ctx.keyword.trim() || ctx.categorias.length > 0 || ctx.shopIds.length > 0)) {
				promises.push(
					carregarCuradoria({
						busca: ctx.keyword,
						comissaoMin: ctx.comissaoMin,
						categorias: ctx.categorias,
						buscasComLojas,
						// Escopa a curadoria na loja recém-adicionada (fix #2 le botanic):
						// com keyword + loja, busca DENTRO da loja, não global.
						shopIds: ctx.shopIds
					}).then((r) => {
						resultado.curadoria = r;
					})
				);
			}

			if ((ctx.fontes.quedas || ctx.fontes.novos) && buscasComLojas.length > 0) {
				promises.push(
					carregarOportunidades(buscasComLojas, nomesLojas).then((r) => {
						resultado.quedas = r.quedas;
						resultado.novos = r.novos;
					})
				);
			}

			if (ctx.fontes.lojas && buscasComLojas.length > 0) {
				promises.push(
					carregarProdutosLojas(buscasComLojas).then((r) => {
						resultado.lojas = r;
					})
				);
			}

			await Promise.race([
				Promise.all(promises),
				new Promise((_, rej) => setTimeout(() => rej(new Error('A busca demorou demais. Tente novamente.')), 25000))
			]);

			return resultado;
		},

		/** Resolve URL/ID de loja via Collector. */
		async resolverLoja(input) {
			return adicionarLoja({ input });
		},

		/** Salva busca no servidor. */
		async salvarBusca(payload) {
			return sincronizarBusca(payload);
		},

		/** Remove (desativa) uma busca salva. */
		async removerBusca(config) {
			return sincronizarBusca({ keywords: config.keywords }, { remover: true });
		}
	};
}

/**
 * Combina buscas do store com lojas do contexto atual (que podem não ter sido salvas ainda).
 * Garante que lojas recém-adicionadas via ADICIONAR_LOJA participem de quedas/novos/lojas.
 */
function buildBuscasComLojas(buscasSalvas, ctx) {
	const buscasComLojas = (buscasSalvas ?? []).filter((b) => b.shop_ids?.length > 0);
	const nomesLojas = Object.fromEntries(
		buscasComLojas.map((b) => {
			// Usa shop_names (novo) ou fallback para nome (legado)
			const label = b.shop_names ? Object.values(b.shop_names)[0] : b.nome || b.id;
			return [b.id, label];
		})
	);

	// Lojas do ctx que ainda não estão no store (adicionou mas não salvou)
	const lojasNoStore = new Set(buscasComLojas.flatMap((b) => b.shop_ids ?? []));
	for (const id of ctx.shopIds ?? []) {
		if (!lojasNoStore.has(id)) {
			const syntheticId = `ctx-loja-${id}`;
			buscasComLojas.push({
				id: syntheticId,
				shop_ids: [id],
				nome: ctx.shopNomes?.[id] || String(id),
				keywords: ctx.keyword ? [ctx.keyword] : []
			});
			nomesLojas[syntheticId] = ctx.shopNomes?.[id] || String(id);
		}
	}

	return { buscasComLojas, nomesLojas };
}
