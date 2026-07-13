/**
 * Estado e guards da BuscaEngine — extraídos da classe para manter o arquivo
 * principal (`busca-engine.svelte.js`) enxuto. São dados/funções puras, sem runes.
 */

import { DEFAULTS, checarGuard, buscarDuplicada } from './busca-config.js';

// ── Estados possíveis (ciclo de rede) ─────────────────────────────────────
export const STATES = { IDLE: 'idle', SEARCHING: 'searching', RESULTS: 'results', SAVING: 'saving', ERROR: 'error' };

// ── Modos de interação (relação com buscas salvas) ────────────────────────
export const MODOS = { EXPLORANDO: 'explorando', VINCULADA: 'vinculada', EDITANDO: 'editando' };

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
		buscaSelecionadaId: null, // id da busca salva vinculada (modo vinculada/editando)
		modo: DEFAULTS.modo, // 'explorando' | 'vinculada' | 'editando'
		erroDuplicata: null, // mensagem de erro de busca duplicada ao salvar
		fontes: { ...DEFAULTS.fontes },
		cron: '',
		resultados: [],
		contagens: { curadoria: 0, quedas: 0, novos: 0, lojas: 0 },
		dadosBrutos: { curadoria: [], quedas: [], novos: [], lojas: [], favoritos: [] },
		buscasSalvas: [],
		resolucaoLoja: { status: 'idle' },
		error: null,
		_feedDefault: false,
		_feedDefaultCategoria: null
	};
}

// ── UI state inicial (separado de domínio) ────────────────────────────────
export function criarUIInicial() {
	return {
		omnibox: {
			inputValue: '',
			aberto: false,
			highlightIdx: -1,
			modo: 'intencao', // 'intencao' | 'sugestoes'
			opcoes: [],
			placeholder: 'Buscar produtos, lojas ou categorias\u2026',
			chipRemovalMessage: ''
		},
		resultados: {
			modo: 'produtos', // 'produtos' | 'lojas'
			lojas: []
		},
		paineis: {
			buscasSalvasAberto: false,
			filtrosAberto: false,
			salvarAberto: false
		}
	};
}

// ── Guards ────────────────────────────────────────────────────────────────
// Delegam para a config declarativa (busca-config.js). `lojaInputValida` opera
// sobre o event, então permanece imperativo.
export const guards = {
	temContextoBusca: (ctx) => checarGuard('temContextoBusca', ctx),
	lojaInputValida: (_ctx, event) => {
		if (event.loja?.id) return Boolean(event.loja.marketplace);
		return (event.value ?? '').trim().length > 0;
	},
	resolucaoPermitida: (ctx) => ctx.resolucaoLoja.status !== 'resolvendo',
	podeSalvar: (ctx) => checarGuard('podeSalvar', ctx),

	/**
	 * Verifica se já existe uma busca salva com os mesmos parâmetros de identidade.
	 * Retorna a busca duplicada (ou null).
	 * @param {object} ctx — contexto atual
	 * @returns {object|null}
	 */
	buscaDuplicada: (ctx) => buscarDuplicada(ctx, ctx.buscasSalvas, ctx.editandoId)
};
