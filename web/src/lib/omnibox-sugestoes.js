/**
 * omnibox-sugestoes — função pura que gera sugestões agrupadas a partir do
 * último token (incompleto) e do contexto disponível no cliente.
 *
 * Sem runes / sem DOM. Separa a lógica de match do componente visual.
 */
import { normalizarNome, matchLojas as registryMatchLojas } from './loja-registry.js';

/**
 * @typedef {import('./omnibox-parser.js').Token} Token
 * @typedef {Object} Sugestao
 * @property {'loja'|'categoria'|'marketplace'|'busca_salva'} tipo
 * @property {string} label - texto para exibir
 * @property {string} valor - texto a inserir no input (com prefixo)
 * @property {string} [icone]
 * @property {Object} [meta] - dados extras (id da loja, config da busca, etc.)
 */

const ICONE = { loja: '🏪', categoria: '📂', marketplace: '🌐', busca_salva: '💾' };

/** Ordem canônica dos grupos (buscas salvas sempre primeiro — Req 4.3). */
const ORDEM = ['busca_salva', 'loja', 'categoria', 'marketplace'];

/**
 * @typedef {Object} SugestoesContext
 * @property {Array<{id:number|string,nome:string,nome_normalizado?:string,marketplace?:string}>} [lojasMonitoradas]
 * @property {Array<{nome:string,marketplaces?:string[]}|string>} [categoriasDisponiveis]
 * @property {string[]} [marketplaces]
 * @property {Array} [buscasSalvas]
 */

/**
 * Gera sugestões agrupadas por tipo para o último token.
 * @param {Token} ultimoToken
 * @param {SugestoesContext} ctx
 * @param {{minChars?:number, maxSugestoes?:number, matchBuscaSalva?:boolean}} [config]
 * @returns {Map<string, Sugestao[]>}
 */
export function gerarSugestoes(ultimoToken, ctx = {}, config = {}) {
	const result = new Map();
	const minChars = config.minChars ?? 2;
	const max = config.maxSugestoes ?? 7;
	const valor = ultimoToken?.valor ?? '';
	if (valor.length < minChars) return result;

	const q = valor.toLowerCase();
	// Token com prefixo → só o tipo correspondente; keyword → todos os tipos.
	const tipos =
		ultimoToken.tipo && ultimoToken.tipo !== 'keyword'
			? [ultimoToken.tipo]
			: ['busca_salva', 'loja', 'categoria', 'marketplace'];

	const geradores = {
		busca_salva: () => (config.matchBuscaSalva ? matchBuscasSalvas(ctx.buscasSalvas, q, max) : []),
		loja: () => matchLojas(ctx.lojasMonitoradas, q, max),
		categoria: () => matchCategorias(ctx.categoriasDisponiveis, q, max),
		marketplace: () => matchMarketplaces(ctx.marketplaces, q, max)
	};

	// Percorre na ORDEM canônica para garantir buscas salvas no topo.
	for (const tipo of ORDEM) {
		if (!tipos.includes(tipo)) continue;
		const itens = geradores[tipo]();
		if (itens.length) result.set(tipo, itens);
	}
	return result;
}

function matchLojas(lojas, q, max) {
	const normQ = normalizarNome(q);
	return registryMatchLojas(normQ, lojas, max).map((l) => ({
		tipo: 'loja',
		label: l.nome,
		valor: '@' + l.nome,
		icone: ICONE.loja,
		meta: l
	}));
}

function matchCategorias(categorias, q, max) {
	return (categorias ?? [])
		.filter((c) =>
			String(c.nome ?? c)
				.toLowerCase()
				.includes(q)
		)
		.slice(0, max)
		.map((c) => {
			const nome = c.nome ?? c;
			return { tipo: 'categoria', label: nome, valor: '#' + nome, icone: ICONE.categoria, meta: c };
		});
}

function matchMarketplaces(marketplaces, q, max) {
	return (marketplaces ?? [])
		.filter((m) => m.toLowerCase().startsWith(q))
		.slice(0, max)
		.map((m) => ({
			tipo: 'marketplace',
			label: m,
			valor: '!' + m,
			icone: ICONE.marketplace,
			meta: { marketplace: m }
		}));
}

function matchBuscasSalvas(buscas, q, max) {
	return (buscas ?? [])
		.filter(
			(b) =>
				(b.keywords ?? []).some((k) => String(k).toLowerCase().includes(q)) ||
				Object.values(b.shopNomes ?? {}).some((n) => String(n).toLowerCase().includes(q))
		)
		.slice(0, max)
		.map((b) => {
			const label = (b.keywords ?? []).join(', ') || Object.values(b.shopNomes ?? {})[0] || 'busca salva';
			return {
				tipo: 'busca_salva',
				label,
				valor: (b.keywords ?? [])[0] ?? '',
				icone: ICONE.busca_salva,
				meta: { config: b }
			};
		});
}
