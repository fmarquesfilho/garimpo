/**
 * busca-config.js - Configuração declarativa da BuscaEngine (página Garimpar).
 *
 * Importa regras de `rules/busca-rules.json` (fonte única, external ao código).
 * O JSON é testável independentemente por qualquer linguagem/tool:
 *   - E2E importam o JSON para validar comportamento do frontend
 *   - CI valida o schema em tempo de build (drift check)
 *   - Qualquer serviço futuro pode ler o mesmo arquivo
 *
 * Este módulo re-exporta as regras no formato consumido pela BuscaEngine e
 * fornece funções puras de avaliação (normalização, guards, intent, modos,
 * fingerprint para detecção de buscas duplicadas).
 */

import rules from '../../../rules/busca-rules.json';

// ── Re-export das regras externas ────────────────────────────────────────────
export const DEFAULTS = rules.defaults;
export const NORMALIZE = rules.normalize;
export const GUARDS = rules.guards;
export const TRANSICOES = rules.transicoes;
export const MARKETPLACES = rules.marketplaces;
export const CONTEXTO_CATEGORIAS = rules.contextoCategorias;
export const MODOS = rules.modos;
export const BUSCA_DUPLICADA = rules.buscaDuplicada;
export const OMNIBOX = rules.omnibox;
export const INTENCAO_CONFIG = rules.omnibox?.intencao;
export const LOJA_REGISTRO = rules.lojaRegistro;
export const FEED_DEFAULT = rules.feedDefault;

// Tabela de intent no formato esperado pela engine
export const INTENT_TABLE = rules.intent.map((r) => ({
	when: { keyword: r.keyword, shop: r.shop },
	intent: r.result,
	sources: r.sources
}));

// ── Funções puras derivadas da config (testáveis sem DOM/API) ─────────────────

/** Normaliza comissão para decimal em [0,1] conforme NORMALIZE.comissao. */
export function normalizarComissao(v) {
	const c = NORMALIZE.comissao;
	if (typeof v !== 'number' || isNaN(v)) return DEFAULTS.comissaoMin;
	if (c.divideBy100IfGt1 && v > 1) v = v / 100;
	const clamped = Math.max(c.min, Math.min(c.max, v));
	const f = 10 ** c.decimals;
	return Math.round(clamped * f) / f;
}

/** Normaliza vendas mínimas para inteiro >= min. */
export function normalizarVendas(v) {
	const s = NORMALIZE.vendas;
	let n = Number(v);
	if (!Number.isFinite(n)) n = 0;
	if (s.floor) n = Math.floor(n);
	return Math.max(s.min, n);
}

/** Avalia um guard declarativo (`requiresAny`) contra o contexto. */
export function checarGuard(nome, ctx) {
	const g = GUARDS[nome];
	if (!g) return true;
	if (g.requiresAny) {
		return g.requiresAny.some((campo) => {
			const val = ctx?.[campo];
			if (Array.isArray(val)) return val.length > 0;
			if (typeof val === 'string') return val.trim().length > 0;
			return Boolean(val);
		});
	}
	return true;
}

/** Deriva o intent de busca a partir do contexto (keyword × loja). */
export function intentBusca(ctx) {
	const keyword = (ctx?.keyword ?? '').trim().length > 0;
	const shop = (ctx?.shopIds ?? []).length > 0;
	const row = INTENT_TABLE.find((r) => r.when.keyword === keyword && r.when.shop === shop);
	return row?.intent ?? 'nenhum';
}

/**
 * Sources efetivos a consultar dado o contexto.
 * Segue a intent table (keyword × loja); quando o intent é `nenhum` mas há
 * categorias selecionadas, cai no contexto de categorias (sources globais),
 * pois uma busca só-categorias é válida (lista produtos das categorias).
 */
export function sourcesBusca(ctx) {
	const keyword = (ctx?.keyword ?? '').trim().length > 0;
	const shop = (ctx?.shopIds ?? []).length > 0;
	const row = INTENT_TABLE.find((r) => r.when.keyword === keyword && r.when.shop === shop);
	const sources = row?.sources ?? [];
	if (sources.length === 0 && (ctx?.categorias ?? []).length > 0) {
		return CONTEXTO_CATEGORIAS.sources;
	}
	return sources;
}

/** Label de comissão para exibição (0.07 → "7%"), evitando "7.0000000". */
export function comissaoPercentLabel(comissaoMin) {
	return `${Math.round((comissaoMin ?? 0) * 100)}%`;
}

// ── Modos de interação (FSM declarativa) ─────────────────────────────────────

/**
 * Calcula o próximo modo de interação dado o modo atual e o tipo de evento.
 * Lê as regras declarativas de `rules.modos`.
 *
 * Lógica:
 * 1. Se o modo atual tem uma transição explícita para o evento → usa-a.
 * 2. Se o modo atual é `vinculada` e o evento está em `desvinculaEm` → volta a `explorando`.
 * 3. Caso contrário → modo permanece inalterado.
 *
 * @param {string} modoAtual - 'explorando' | 'vinculada' | 'editando'
 * @param {string} tipoEvento - ex: 'DIGITAR', 'CARREGAR_SALVA', 'SALVAR'
 * @returns {string} próximo modo
 */
export function proximoModo(modoAtual, tipoEvento) {
	const modoDef = MODOS[modoAtual];
	if (!modoDef) return DEFAULTS.modo;

	// 1. Transição explícita declarada no JSON
	if (modoDef.transicoes?.[tipoEvento]) {
		return modoDef.transicoes[tipoEvento];
	}

	// 2. Eventos que desvinculam (modo vinculada → explorando)
	if (modoDef.desvinculaEm?.includes(tipoEvento)) {
		return 'explorando';
	}

	// 3. Sem transição → modo inalterado
	return modoAtual;
}

// ── Detecção de busca duplicada ──────────────────────────────────────────────

/**
 * Gera um fingerprint determinístico de uma configuração de busca, usando
 * apenas os campos de identidade definidos em `rules.buscaDuplicada`.
 *
 * @param {object} ctx - contexto da engine ou config de busca salva (normalizada)
 * @returns {string} fingerprint para comparação
 */
export function fingerprint(ctx) {
	const campos = BUSCA_DUPLICADA.camposIdentidade;
	const partes = campos.map((campo) => {
		const val = ctx?.[campo];
		const norm = BUSCA_DUPLICADA.normalizacao[campo];

		if (val == null) {
			// For array fields, null/undefined is equivalent to []
			return norm ? '[]' : '';
		}
		if (Array.isArray(val)) {
			if (val.length === 0) return '[]';
			if (norm === 'sort_lowercase') {
				return JSON.stringify([...val].map((s) => String(s).toLowerCase()).sort());
			}
			// sort — coage a string para o fingerprint casar independente da origem
			// (shopIds de busca salva chegam numéricos; do registro, como string).
			return JSON.stringify([...val].map((v) => String(v)).sort());
		}
		if (typeof val === 'string') return val.trim().toLowerCase();
		return String(val);
	});
	return partes.join('|');
}

/**
 * Converte uma busca salva (formato payloadToConfig) para o formato de ctx
 * para comparação via fingerprint.
 * @param {object} busca - config de busca salva
 * @returns {object} ctx parcial com os campos de identidade
 */
export function buscaSalvaToCtx(busca) {
	return {
		keyword: (busca.keywords ?? [])[0] ?? '',
		shopIds: busca.shopIds ?? [],
		categorias: busca.categorias ?? [],
		marketplacesFiltro: Array.isArray(busca.marketplaces) ? busca.marketplaces : []
	};
}

/**
 * Procura uma busca salva com os mesmos parâmetros de identidade que o contexto
 * atual. Retorna a busca duplicada (ou null).
 *
 * @param {object} ctx - contexto atual da engine
 * @param {object[]} buscasSalvas - lista de buscas salvas (formato payloadToConfig)
 * @param {string|null} [excluirId] - id a excluir da comparação (edit mode)
 * @returns {object|null} busca duplicada ou null
 */
export function buscarDuplicada(ctx, buscasSalvas, excluirId = null) {
	const fp = fingerprint(ctx);
	return (
		(buscasSalvas ?? []).find((b) => {
			if (excluirId && b.id === excluirId) return false;
			return fingerprint(buscaSalvaToCtx(b)) === fp;
		}) ?? null
	);
}
