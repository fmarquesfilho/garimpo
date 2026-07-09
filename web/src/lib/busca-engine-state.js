/**
 * Estado e guards da BuscaEngine — extraídos da classe para manter o arquivo
 * principal (`busca-engine.svelte.js`) enxuto. São dados/funções puras, sem runes.
 */

import { DEFAULTS, checarGuard } from './busca-config.js';

// ── Estados possíveis ─────────────────────────────────────────────────────
export const STATES = { IDLE: 'idle', SEARCHING: 'searching', RESULTS: 'results', SAVING: 'saving', ERROR: 'error' };

// ── Context inicial ───────────────────────────────────────────────────────
export function criarContextoInicial() {
	return {
		keyword: '',
		shopIds: [],
		shopNomes: {},
		comissaoMin: DEFAULTS.comissaoMin,
		vendasMin: DEFAULTS.vendasMin,
		categorias: [],
		categoriasDisponiveis: [],
		lojasDisponiveis: [], // lojas monitoradas para o autocomplete da raia Lojas
		categoriaMeta: {}, // nome → marketplaces[] (para os cards de categoria)
		shopMeta: {}, // id → { marketplace, origem, monitorada, cron } (para os cards de loja)
		marketplacesFiltro: [], // marketplaces em que a busca é escopada (vazio = todos)
		editandoId: null, // id da busca salva em edição (edit mode)
		fontes: { ...DEFAULTS.fontes },
		cron: '',
		resultados: [],
		contagens: { curadoria: 0, quedas: 0, novos: 0, lojas: 0 },
		dadosBrutos: { curadoria: [], quedas: [], novos: [], lojas: [], favoritos: [] },
		buscasSalvas: [],
		lojaResolvendo: false,
		lojaErro: '',
		error: null
	};
}

// ── Guards ────────────────────────────────────────────────────────────────
// Delegam para a config declarativa (busca-config.js). `lojaInputValida` opera
// sobre o event, então permanece imperativo.
export const guards = {
	temContextoBusca: (ctx) => checarGuard('temContextoBusca', ctx),
	lojaInputValida: (_ctx, event) => (event.value ?? '').trim().length > 0,
	podeSalvar: (ctx) => checarGuard('podeSalvar', ctx)
};
