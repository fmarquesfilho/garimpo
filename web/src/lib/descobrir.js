/**
 * Lógica de carregamento e filtragem da página Descobrir.
 * Extraído de +page.svelte para manter a página dentro do limite de linhas.
 * Funções puras (testáveis) estão em descobrir-logic.js.
 */
import { buscarCandidatos, buscarNovidades } from './api.js';
export { montarResultados, encontrarLojaPorNome } from './descobrir-logic.js';
import { encontrarLojaPorNome } from './descobrir-logic.js';

/**
 * Carrega produtos da curadoria (Shopee API).
 * Se o termo bate com uma loja monitorada, busca via shop_ids.
 */
export async function carregarCuradoria({ busca, comissaoMin, vendasMin, categorias, buscasComLojas }) {
	try {
		const termo = (busca ?? '').trim();
		const lojaMatch = encontrarLojaPorNome(termo, buscasComLojas);

		let params;
		if (lojaMatch) {
			params = {
				estrategia: 'nicho', top: 50,
				fonte: 'shopee-shop',
				shopIds: lojaMatch.shop_ids.join(','),
				semFiltro: true
			};
		} else {
			// Se não há keyword mas há categoria, usa a categoria como keyword
			const keywordEfetiva = termo || (categorias?.length > 0 ? categorias[0] : '');
			params = {
				estrategia: 'nicho', top: 20,
				keyword: keywordEfetiva || undefined,
				comissaoMin: comissaoMin > 0 ? comissaoMin : undefined
			};
		}
		if (categorias?.length > 0) params.categoria = categorias[0];
		const r = await buscarCandidatos(params);
		return (r.candidatos ?? []).map(c => ({ ...c, _fonte: 'curadoria' }));
	} catch {
		return [];
	}
}

/**
 * Carrega oportunidades (novidades/quedas) de todas as lojas monitoradas.
 * Usa cache de 2 minutos.
 */
let cacheOportunidades = { em: 0, quedas: [], novos: [] };

export async function carregarOportunidades(buscasComLojas, nomesLojas) {
	if (Date.now() - cacheOportunidades.em < 120000 &&
		(cacheOportunidades.quedas.length > 0 || cacheOportunidades.novos.length > 0)) {
		return { quedas: cacheOportunidades.quedas, novos: cacheOportunidades.novos };
	}

	try {
		const promises = buscasComLojas.map(b =>
			buscarNovidades({ buscaId: b.id, dias: 7 }).then(r => ({ ...r, loja: b.id })).catch(() => null)
		);
		const resultados = await Promise.all(promises);
		const quedas = [], novos = [];

		for (const r of resultados) {
			if (!r) continue;
			for (const v of (r.variacoes ?? [])) {
				if (v.variacao_pct < 0) {
					quedas.push({
						id: v.produto_id, produto_id: v.produto_id, nome: v.nome,
						preco: v.preco_atual, preco_anterior: v.preco_anterior,
						variacao_pct: v.variacao_pct, detectado_em: v.detectado_em,
						loja: v.loja || (nomesLojas[r.loja] ?? r.loja), _loja_id: r.loja,
						imagem: v.imagem, link: v.link, comissao: v.comissao ?? 0,
						vendas: v.vendas ?? 0, _fonte: 'queda'
					});
				}
			}
			for (const p of (r.produtos_novos ?? [])) {
				novos.push({
					id: p.produto_id, produto_id: p.produto_id, nome: p.nome,
					preco: p.preco, comissao: p.comissao ?? 0, vendas: p.vendas ?? 0,
					detectado_em: p.detectado_em, loja: p.loja || (nomesLojas[r.loja] ?? r.loja),
					_loja_id: r.loja, imagem: p.imagem, link: p.link, _fonte: 'novo'
				});
			}
		}
		quedas.sort((a, b) => a.variacao_pct - b.variacao_pct);
		novos.sort((a, b) => (b.detectado_em ?? '').localeCompare(a.detectado_em ?? ''));
		cacheOportunidades = { em: Date.now(), quedas, novos };
		return { quedas, novos };
	} catch {
		return { quedas: [], novos: [] };
	}
}

/**
 * Monta a lista final de resultados aplicando todos os filtros client-side.
 * Re-exported from descobrir-logic.js for backward compat.
 */
