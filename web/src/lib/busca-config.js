/**
 * busca-config.js — Configuração declarativa da BuscaEngine (página Garimpar).
 *
 * Importa regras de `rules/busca-rules.json` (fonte única, external ao código).
 * O JSON é testável independentemente por qualquer linguagem/tool:
 *   - E2E importam o JSON para validar comportamento do frontend
 *   - CI valida o schema em tempo de build (drift check)
 *   - Qualquer serviço futuro pode ler o mesmo arquivo
 *
 * Este módulo re-exporta as regras no formato consumido pela BuscaEngine e
 * fornece funções puras de avaliação (normalização, guards, intent).
 */

import rules from '../../../rules/busca-rules.json';

// ── Re-export das regras externas ────────────────────────────────────────────
export const DEFAULTS = rules.defaults;
export const NORMALIZE = rules.normalize;
export const GUARDS = rules.guards;
export const TRANSICOES = rules.transicoes;

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

/** Label de comissão para exibição (0.07 → "7%"), evitando "7.0000000". */
export function comissaoPercentLabel(comissaoMin) {
	return `${Math.round((comissaoMin ?? 0) * 100)}%`;
}
