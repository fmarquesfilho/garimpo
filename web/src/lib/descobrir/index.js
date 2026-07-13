/**
 * Módulo Descobrir — API Pública (barrel).
 *
 * Este é o ponto de entrada recomendado para consumidores EXTERNOS ao módulo
 * (rotas, outros feature modules). Internamente, os arquivos referenciam-se
 * entre si por path relativo — isso é permitido.
 *
 * Consumidores externos devem importar daqui:
 *   import { BuscaEngine, criarEffects, STATES } from '$lib/descobrir';
 *
 * O ESLint (no-restricted-imports) bloqueia imports diretos dos internals
 * por código fora do módulo. Testes do próprio módulo estão excluídos.
 *
 * Internals (NÃO importar diretamente de fora):
 *   busca-engine.svelte.js, busca-engine-state.js, busca-engine-effects.js,
 *   busca-engine-omnibox.js, busca-engine-lojas.js, busca-engine-persistencia.js,
 *   busca-config.js, descobrir-logic.js, descobrir.js, busca-unificada-logic.js,
 *   omnibox-intencao.js, omnibox-parser.js, omnibox-sugestoes.js
 */

// ── Engine (classe + constantes exportadas) ───────────────────────────────
export { BuscaEngine, STATES, MODOS, guards, gerarLabelBusca, cronLabel, gerarResumo } from '../busca-engine.svelte.js';

// ── Effects (factory injetável) ───────────────────────────────────────────
export { criarEffects } from '../busca-engine-effects.js';

// ── State (factories de contexto/UI inicial) ──────────────────────────────
export { criarContextoInicial, criarUIInicial } from '../busca-engine-state.js';

// ── Config (constantes + funções puras do JSON de regras) ─────────────────
export {
	DEFAULTS,
	NORMALIZE,
	GUARDS,
	TRANSICOES,
	MARKETPLACES,
	CONTEXTO_CATEGORIAS,
	BUSCA_DUPLICADA,
	OMNIBOX,
	INTENCAO_CONFIG,
	LOJA_REGISTRO,
	FEED_DEFAULT,
	INTENT_TABLE,
	normalizarComissao,
	normalizarVendas,
	checarGuard,
	intentBusca,
	sourcesBusca,
	comissaoPercentLabel,
	proximoModo,
	fingerprint,
	buscaSalvaToCtx,
	buscarDuplicada
} from '../busca-config.js';

// ── Logic (filtragem client-side) ─────────────────────────────────────────
export { montarResultados, agruparCategoriasPorMarketplace } from '../descobrir-logic.js';

// ── Busca unificada (conversão payload ↔ config) ─────────────────────────
export { payloadToConfig, configToPayload, contarFiltrosAtivos } from '../busca-unificada-logic.js';
