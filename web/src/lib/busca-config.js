/**
 * busca-config.js — Configuração declarativa da BuscaEngine (página Garimpar).
 *
 * Fonte única e versionada (git) das REGRAS da busca: defaults, normalização,
 * guards simples, flags por transição e a tabela de decisão do "intent" de busca.
 * Cada regra é revisável por PR e serve de spec para os testes.
 *
 * Princípio de projeto:
 *   CONFIG para valores e tabelas de decisão (dados).
 *   CÓDIGO para fluxo de controle (a FSM em si).
 * Guards/normalizações simples viram dados aqui; orquestração assíncrona fica
 * nos effects; as transições de estado ficam no switch da engine.
 */

// ── Defaults do contexto ─────────────────────────────────────────────────────
export const DEFAULTS = {
	comissaoMin: 0.07,
	vendasMin: 0,
	fontes: { curadoria: true, quedas: true, novos: true, lojas: false, favoritos: false },
	debounceMs: 400,
	timeoutMs: 25000
};

// ── Especificação de normalização ────────────────────────────────────────────
export const NORMALIZE = {
	comissao: { divideBy100IfGt1: true, min: 0, max: 1, decimals: 4 },
	vendas: { floor: true, min: 0 }
};

// ── Guards declarativos (condições simples sobre o contexto) ──────────────────
// `requiresAny`: ao menos um dos campos precisa estar "preenchido"
// (string não-vazia ou array não-vazio).
export const GUARDS = {
	temContextoBusca: { requiresAny: ['keyword', 'shopIds'] },
	podeSalvar: { requiresAny: ['keyword', 'shopIds'] }
};

// ── Flags por transição (evento → comportamento) ──────────────────────────────
// refetch: o evento muda os DADOS (precisa ir à API) vs. só refiltra client-side.
// imediato: dispara a busca sem esperar o debounce.
export const TRANSICOES = {
	DIGITAR: { refetch: true, imediato: false },
	ADICIONAR_LOJA: { refetch: true, imediato: true },
	REMOVER_LOJA: { refetch: true, imediato: false },
	MUDAR_FILTRO: { refetch: false, imediato: false }, // filtros são client-side
	MUDAR_FONTES: { refetch: true, imediato: false },
	CARREGAR_SALVA: { refetch: true, imediato: true }
};

// ── Tabela de decisão do "intent" de busca ────────────────────────────────────
// Resolve o escopo da coleta a partir do contexto. É a regra que conserta o bug
// "adicionei a loja mas aparecem produtos de fora dela": com keyword E loja, a
// busca deve ser ESCOPADA na loja, não global.
export const INTENT_TABLE = [
	{ when: { keyword: true, shop: true }, intent: 'keyword_na_loja' },
	{ when: { keyword: true, shop: false }, intent: 'keyword_global' },
	{ when: { keyword: false, shop: true }, intent: 'loja_completa' },
	{ when: { keyword: false, shop: false }, intent: 'nenhum' }
];

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
