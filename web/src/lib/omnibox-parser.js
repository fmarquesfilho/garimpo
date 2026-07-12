/**
 * omnibox-parser — funções puras que tokenizam o texto do Omnibox.
 *
 * O input funciona como uma mini-CLI: `serum @lebotanic #beleza !shopee`.
 * Prefixos (`@`, `#`, `!`) forçam o tipo; texto sem prefixo é keyword (default).
 * Ver gramática em T-0055 e `design.md` do spec `omnibox-input`.
 *
 * Sem runes / sem DOM — testável isoladamente.
 */

import { normalizarNome, matchLojas } from './loja-registry.js';

/**
 * @typedef {Object} Token
 * @property {'keyword'|'loja'|'categoria'|'marketplace'} tipo
 * @property {string} valor - texto sem o caractere de prefixo
 * @property {boolean} completo - true se seguido de espaço ou não for o último token
 */

/** prefixo → tipo */
const PREFIXO_TIPO = { '@': 'loja', '#': 'categoria', '!': 'marketplace' };
/** tipo → prefixo (inverso, para serializar) */
const TIPO_PREFIXO = { loja: '@', categoria: '#', marketplace: '!', keyword: '' };

/**
 * Tokeniza o raw text do input em Token[].
 * @param {string} raw
 * @returns {Token[]}
 */
export function parsearInput(raw) {
	if (typeof raw !== 'string') return [];
	const partes = raw.split(/\s+/).filter((p) => p.length > 0);
	if (partes.length === 0) return [];
	const terminaComEspaco = /\s$/.test(raw);

	return partes.map((part, i) => {
		const completo = i < partes.length - 1 || terminaComEspaco;
		const tipo = PREFIXO_TIPO[part[0]];
		if (tipo) return { tipo, valor: part.slice(1), completo };
		return { tipo: 'keyword', valor: part, completo };
	});
}

/**
 * Serializa tokens de volta para string (inverso de parsearInput).
 * Round-trip: parsearInput(serializarTokens(parsearInput(x))) === parsearInput(x).
 * @param {Token[]} tokens
 * @returns {string}
 */
export function serializarTokens(tokens) {
	if (!Array.isArray(tokens) || tokens.length === 0) return '';
	const texto = tokens.map((t) => (TIPO_PREFIXO[t.tipo] ?? '') + t.valor).join(' ');
	// Se o último token está completo, houve um espaço final que precisa ser preservado
	return tokens[tokens.length - 1].completo ? texto + ' ' : texto;
}

/**
 * @typedef {Object} ResolveCtx
 * @property {Array<{id:string,nome:string,marketplace?:string}>} [lojasMonitoradas]
 * @property {Array<{nome:string}|string>} [categoriasDisponiveis]
 * @property {string[]} [marketplaces]
 */

/**
 * Resolve tokens contra o contexto disponível → contexto para a BuscaEngine.
 * Só inclui lojas/categorias/marketplaces que casam com o contexto (evita poluir
 * a engine com valores inexistentes). Keywords são sempre concatenadas.
 *
 * @param {Token[]} tokens
 * @param {ResolveCtx} [ctx]
 * @returns {{ keyword:string, shopIds:string[], categorias:string[], marketplacesFiltro:string[], lojasResolvidas:Array }}
 */
export function tokensParaContexto(tokens, ctx = {}) {
	const out = { keyword: '', shopIds: [], categorias: [], marketplacesFiltro: [], lojasResolvidas: [] };
	const keywords = [];

	for (const t of tokens ?? []) {
		const q = t.valor.trim().toLowerCase();
		if (t.tipo === 'keyword') {
			if (t.valor.trim()) keywords.push(t.valor.trim());
		} else if (q) {
			RESOLVERS[t.tipo]?.(q, ctx, out);
		}
	}

	out.keyword = keywords.join(' ');
	return out;
}

/** Resolvers por tipo de token — mantêm `tokensParaContexto` com baixa complexidade. */
const RESOLVERS = {
	loja(q, ctx, out) {
		// Usa o match normalizado do registry (mesmo do dropdown) para consistência:
		// `@gloryofseoul` deve resolver "Glory of Seoul" tanto no dropdown quanto no Enter.
		// A fonte de escopo real é `lojasResolvidas` (a engine coage o id para número).
		const [loja] = matchLojas(normalizarNome(q), ctx.lojasMonitoradas ?? [], 1);
		if (loja && !out.shopIds.includes(loja.id)) {
			out.shopIds.push(loja.id);
			out.lojasResolvidas.push(loja);
		}
	},
	categoria(q, ctx, out) {
		const cat = casar(ctx.categoriasDisponiveis ?? [], (c) => c.nome ?? c, q);
		const nome = cat ? (cat.nome ?? cat) : null;
		if (nome && !out.categorias.includes(nome)) out.categorias.push(nome);
	},
	marketplace(q, ctx, out) {
		const mkts = ctx.marketplaces ?? [];
		const mkt = mkts.find((m) => m.toLowerCase() === q) || mkts.find((m) => m.toLowerCase().startsWith(q));
		if (mkt && !out.marketplacesFiltro.includes(mkt)) out.marketplacesFiltro.push(mkt);
	}
};

/** Casa por igualdade exata (prioridade) e cai para match parcial case-insensitive. */
function casar(lista, getNome, q) {
	return (
		lista.find((it) => String(getNome(it) ?? '').toLowerCase() === q) ||
		lista.find((it) =>
			String(getNome(it) ?? '')
				.toLowerCase()
				.includes(q)
		) ||
		null
	);
}
