/**
 * Lógica de carregamento e filtragem da página Descobrir.
 * Extraído de +page.svelte para manter a página dentro do limite de linhas.
 * Funções puras (testáveis) estão em descobrir-logic.js.
 */
import { buscarCandidatos, buscarNovidades } from './api.js';
export { montarResultados } from './descobrir-logic.js';


/**
 * Carrega produtos da curadoria (Shopee API).
 * Se o termo bate com uma loja monitorada, busca via shop_ids.
 */
export async function carregarCuradoria({ busca, comissaoMin, categorias, buscasComLojas, shopIds = null }) {
	try {
		const termo = (busca ?? '').trim();
		const cat0 = categorias?.length > 0 ? categorias[0] : undefined;
		const lojaIds = shopIds?.length > 0 ? shopIds : null;
		const params = buildCuradoriaParams(lojaIds, termo, comissaoMin, cat0);

		const r = await buscarCandidatos(params);
		return (r.candidatos ?? []).map((c) => ({ ...c, _fonte: 'curadoria' }));
	} catch (e) {
		if (isServerError(e)) throw e;
		return [];
	}
}

/** Monta parâmetros de busca: com loja (escopada) ou global (keyword). */
function buildCuradoriaParams(lojaIds, termo, comissaoMin, cat0) {
	if (lojaIds?.length > 0) {
		return {
			estrategia: 'nicho',
			top: 50,
			fonte: 'shopee-shop',
			shopIds: lojaIds.join(','),
			semFiltro: true,
			categoria: cat0
		};
	}
	return {
		estrategia: 'nicho',
		top: 20,
		keyword: termo || cat0 || undefined,
		comissaoMin: comissaoMin > 0 ? comissaoMin : undefined,
		categoria: cat0
	};
}

/** Determina se um erro deve ser propagado ao usuário (servidor respondeu com falha). */
function isServerError(e) {
	return e?.status >= 400 || (e?.message ?? '').includes('HTTP');
}

/**
 * Carrega oportunidades (novidades/quedas) de todas as lojas monitoradas.
 * Usa cache de 2 minutos.
 */
let cacheOportunidades = { em: 0, quedas: [], novos: [] };

function extrairQuedas(resultado, nomesLojas) {
	return (resultado.variacoes ?? [])
		.filter((v) => v.variacao_pct < 0)
		.map((v) => ({
			id: v.produto_id,
			produto_id: v.produto_id,
			nome: v.nome,
			preco: v.preco_atual,
			preco_anterior: v.preco_anterior,
			variacao_pct: v.variacao_pct,
			detectado_em: v.detectado_em,
			loja: v.loja || (nomesLojas[resultado.loja] ?? resultado.loja),
			_loja_id: resultado.loja,
			imagem: v.imagem,
			link: v.link,
			comissao: v.comissao ?? 0,
			vendas: v.vendas ?? 0,
			_fonte: 'queda'
		}));
}

function extrairNovos(resultado, nomesLojas) {
	return (resultado.produtos_novos ?? []).map((p) => ({
		id: p.produto_id,
		produto_id: p.produto_id,
		nome: p.nome,
		preco: p.preco,
		comissao: p.comissao ?? 0,
		vendas: p.vendas ?? 0,
		detectado_em: p.detectado_em,
		loja: p.loja || (nomesLojas[resultado.loja] ?? resultado.loja),
		_loja_id: resultado.loja,
		imagem: p.imagem,
		link: p.link,
		_fonte: 'novo'
	}));
}

export async function carregarOportunidades(buscasComLojas, nomesLojas) {
	const cacheValido =
		Date.now() - cacheOportunidades.em < 120000 &&
		(cacheOportunidades.quedas.length > 0 || cacheOportunidades.novos.length > 0);
	if (cacheValido) {
		return { quedas: cacheOportunidades.quedas, novos: cacheOportunidades.novos };
	}

	try {
		const promises = buscasComLojas.map((b) => {
			// BuscaContract: sempre enviar o UUID da busca como busca_id.
			// O Analyzer faz WHERE busca_id = @busca_id (exact match).
			const buscaId = b.id;
			return buscarNovidades({ buscaId, dias: 7 })
				.then((r) => ({ ...r, loja: b.id }))
				.catch(() => null);
		});
		const resultados = (await Promise.all(promises)).filter(Boolean);

		const quedas = resultados.flatMap((r) => extrairQuedas(r, nomesLojas));
		const novos = resultados.flatMap((r) => extrairNovos(r, nomesLojas));

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

/**
 * Orchestrates parallel loading of all data sources based on active fontes.
 * Returns { curadoria, quedas, novos, lojas } arrays.
 */
export async function carregarFontes({ fontes, busca, comissaoMin, categorias, buscasComLojas, nomesLojas }) {
	const resultados = { curadoria: [], quedas: [], novos: [], lojas: [] };
	const promises = [];

	if (fontes.curadoria && (busca.trim() || categorias.length > 0)) {
		promises.push(
			carregarCuradoria({ busca, comissaoMin, categorias, buscasComLojas }).then((r) => {
				resultados.curadoria = r;
			})
		);
	}

	if ((fontes.quedas || fontes.novos) && buscasComLojas.length > 0) {
		promises.push(
			carregarOportunidades(buscasComLojas, nomesLojas).then((r) => {
				resultados.quedas = r.quedas;
				resultados.novos = r.novos;
			})
		);
	}

	if (fontes.lojas && buscasComLojas.length > 0) {
		promises.push(
			carregarProdutosLojas(buscasComLojas).then((r) => {
				resultados.lojas = r;
			})
		);
	}

	const timeoutMs = 25000;
	await Promise.race([
		Promise.all(promises),
		new Promise((_, rej) => setTimeout(() => rej(new Error('A busca demorou demais. Tente novamente.')), timeoutMs))
	]);

	return resultados;
}

/**
 * Carrega produtos das lojas monitoradas (fonte 🏪).
 * Cache de 2 minutos para evitar re-fetch em toggle rápido.
 */
let cacheLojas = { em: 0, produtos: [] };

export async function carregarProdutosLojas(buscasComLojas) {
	const cacheValido = Date.now() - cacheLojas.em < 120000 && cacheLojas.produtos.length > 0;
	if (cacheValido) return cacheLojas.produtos;

	const promises = buscasComLojas.map((b) =>
		buscarCandidatos({
			estrategia: 'nicho',
			top: 50,
			fonte: 'shopee-shop',
			shopIds: b.shop_ids.join(','),
			semFiltro: true
		})
			.then((r) =>
				(r.candidatos ?? []).map((c) => ({
					...c,
					_fonte: 'loja',
					_loja_id: b.id,
					loja: c.loja || (b.shop_names ? Object.values(b.shop_names)[0] : null) || b.nome || b.id
				}))
			)
			.catch(() => [])
	);

	const resultados = await Promise.all(promises);
	const produtos = resultados.flat();
	cacheLojas = { em: Date.now(), produtos };
	return produtos;
}
