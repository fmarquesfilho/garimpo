/**
 * BuscaEngine — Máquina de estados da página Garimpar.
 *
 * Classe Svelte 5 com campos $state reativos. Toda transição de estado é
 * explícita via send(event). Guards impedem estados incoerentes.
 * Effects (API calls) são injetáveis para testabilidade.
 *
 * Uso:
 *   const engine = new BuscaEngine(effects);
 *   engine.send({ type: 'DIGITAR', value: 'serum' });
 *   // engine.status, engine.ctx são reativos
 */

import { montarResultados } from './descobrir-logic.js';
import {
	configToPayload,
	payloadToConfig,
	gerarResumo,
	contarFiltrosAtivos,
	gerarLabelBusca,
	cronLabel
} from './busca-unificada-logic.js';
import { DEFAULTS, normalizarComissao, normalizarVendas, checarGuard, intentBusca } from './busca-config.js';

// ── Estados possíveis ─────────────────────────────────────────────────────
export const STATES = { IDLE: 'idle', SEARCHING: 'searching', RESULTS: 'results', SAVING: 'saving', ERROR: 'error' };

// ── Context inicial ───────────────────────────────────────────────────────
function criarContextoInicial() {
	return {
		keyword: '',
		shopIds: [],
		shopNomes: {},
		comissaoMin: DEFAULTS.comissaoMin,
		vendasMin: DEFAULTS.vendasMin,
		categorias: [],
		categoriasDisponiveis: [],
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

// ── Classe Engine ─────────────────────────────────────────────────────────
export class BuscaEngine {
	// Estado reativo (Svelte 5 compila para getter/setter)
	status = $state(STATES.IDLE);
	ctx = $state(criarContextoInicial());

	// UI state (não faz parte da FSM formal mas é reativo)
	filtrosAberto = $state(false);
	salvarAberto = $state(false);
	colapsado = $state(false);

	// Derivados reativos
	get loading() {
		return this.status === STATES.SEARCHING;
	}
	get resumo() {
		return gerarResumo(this.ctx);
	}
	get filtrosAtivos() {
		return contarFiltrosAtivos(this.ctx);
	}
	get fontesAtivas() {
		return Object.entries(this.ctx.fontes)
			.filter(([, v]) => v)
			.map(([k]) => k);
	}
	/** Intent de busca derivado do contexto (keyword × loja) — ver busca-config.js. */
	get intent() {
		return intentBusca(this.ctx);
	}

	/** @type {import('./busca-engine-effects.js').Effects} */
	#effects;
	#debounceTimer = null;

	constructor(effects) {
		this.#effects = effects;
	}

	// ── API pública ─────────────────────────────────────────────────────────
	async send(event) {
		switch (event.type) {
			case 'INICIALIZAR':
				return this.#inicializar();
			case 'DIGITAR':
				return this.#digitar(event);
			case 'ADICIONAR_LOJA':
				return this.#adicionarLoja(event);
			case 'REMOVER_LOJA':
				return this.#removerLoja(event);
			case 'MUDAR_FILTRO':
				return this.#mudarFiltro(event);
			case 'MUDAR_FONTES':
				return this.#mudarFontes(event);
			case 'SALVAR':
				return this.#salvar();
			case 'CARREGAR_SALVA':
				return this.#carregarSalva(event);
			case 'REMOVER_SALVA':
				return this.#removerSalva(event);
			case 'RETRY':
				return this.#executarBusca();
			case 'LIMPAR':
				return this.#limpar();
		}
	}

	// ── Transições privadas ─────────────────────────────────────────────────

	async #inicializar() {
		this.status = STATES.SEARCHING;
		try {
			const [buscas, categorias] = await Promise.all([
				this.#effects.carregarBuscasSalvas(),
				this.#effects.carregarCategorias()
			]);
			this.ctx.buscasSalvas = (buscas ?? []).map(payloadToConfig);
			this.ctx.categoriasDisponiveis = categorias ?? [];

			await this.#executarBusca();
		} catch (e) {
			this.ctx.error = e?.message ?? 'Falha ao inicializar';
			this.status = STATES.ERROR;
		}
	}

	#digitar(event) {
		this.ctx.keyword = event.value ?? '';
		this.#debounce();
	}

	async #adicionarLoja(event) {
		if (!guards.lojaInputValida(this.ctx, event)) return;
		this.ctx.lojaResolvendo = true;
		this.ctx.lojaErro = '';
		try {
			const r = await this.#effects.resolverLoja(event.value);
			if (r.shop_ids?.length) {
				this.ctx.shopIds = [...this.ctx.shopIds, ...r.shop_ids];
				this.ctx.shopNomes = { ...this.ctx.shopNomes, [r.shop_ids[0]]: r.keyword };
			}
			this.ctx.lojaResolvendo = false;
			// Busca imediata (sem debounce) com keyword + nova loja
			await this.#executarBusca();
		} catch (e) {
			this.ctx.lojaErro = e?.message ?? 'Falha ao resolver loja';
			this.ctx.lojaResolvendo = false;
		}
	}

	#removerLoja(event) {
		this.ctx.shopIds = this.ctx.shopIds.filter((id) => id !== event.shopId);
		const nomes = { ...this.ctx.shopNomes };
		delete nomes[event.shopId];
		this.ctx.shopNomes = nomes;
		this.#debounce();
	}

	#mudarFiltro(event) {
		if ('comissaoMin' in event) this.ctx.comissaoMin = normalizarComissao(event.comissaoMin);
		if ('vendasMin' in event) this.ctx.vendasMin = normalizarVendas(event.vendasMin);
		if ('categorias' in event) this.ctx.categorias = event.categorias ?? [];
		// Filtros aplicam client-side sobre dados brutos (sem re-fetch)
		this.#refiltrar();
	}

	#mudarFontes(event) {
		this.ctx.fontes = event.fontes;
		// Fontes requerem re-fetch (dados diferentes)
		this.#debounce();
	}

	async #salvar() {
		if (!guards.podeSalvar(this.ctx)) return;
		this.status = STATES.SAVING;
		try {
			const payload = configToPayload({
				keywords: this.ctx.keyword ? [this.ctx.keyword] : [],
				shopIds: this.ctx.shopIds,
				comissaoMin: this.ctx.comissaoMin,
				vendasMin: this.ctx.vendasMin,
				categorias: this.ctx.categorias,
				fontes: this.fontesAtivas,
				cron: this.ctx.cron,
				marketplaces: 'shopee'
			});
			await this.#effects.salvarBusca(payload);
			// Recarregar lista de buscas salvas
			const buscas = await this.#effects.carregarBuscasSalvas();
			this.ctx.buscasSalvas = (buscas ?? []).map(payloadToConfig);
			this.salvarAberto = false;
			this.status = STATES.RESULTS;
		} catch (e) {
			this.ctx.error = e?.message ?? 'Falha ao salvar';
			this.status = STATES.RESULTS; // volta para results, não error
		}
	}

	#carregarSalva(event) {
		const config = event.config;
		// Restaura TUDO
		this.ctx.keyword = (config.keywords ?? [])[0] ?? '';
		this.ctx.shopIds = config.shopIds ?? [];
		this.ctx.shopNomes = config.shopNomes ?? {};
		this.ctx.comissaoMin = config.comissaoMin || 0.07;
		this.ctx.vendasMin = config.vendasMin || 0;
		this.ctx.categorias = config.categorias ?? [];
		this.ctx.cron = config.cron ?? '';
		if (config.fontes?.length) {
			this.ctx.fontes = {
				curadoria: config.fontes.includes('curadoria'),
				quedas: config.fontes.includes('quedas'),
				novos: config.fontes.includes('novos'),
				lojas: config.fontes.includes('lojas'),
				favoritos: config.fontes.includes('favoritos')
			};
		}
		// Busca imediata com contexto restaurado
		this.#executarBusca();
	}

	async #removerSalva(event) {
		await this.#effects.removerBusca(event.config);
		const buscas = await this.#effects.carregarBuscasSalvas();
		this.ctx.buscasSalvas = (buscas ?? []).map(payloadToConfig);
	}

	#limpar() {
		const categoriasDispo = this.ctx.categoriasDisponiveis;
		const salvas = this.ctx.buscasSalvas;
		Object.assign(this.ctx, criarContextoInicial());
		this.ctx.categoriasDisponiveis = categoriasDispo;
		this.ctx.buscasSalvas = salvas;
		this.status = STATES.IDLE;
	}

	// ── Internos ────────────────────────────────────────────────────────────

	#debounce() {
		clearTimeout(this.#debounceTimer);
		this.#debounceTimer = setTimeout(() => {
			if (guards.temContextoBusca(this.ctx)) {
				this.#executarBusca();
			} else {
				this.ctx.resultados = [];
				this.ctx.contagens = { curadoria: 0, quedas: 0, novos: 0, lojas: 0 };
				this.status = STATES.IDLE;
			}
		}, DEFAULTS.debounceMs);
	}

	async #executarBusca() {
		this.status = STATES.SEARCHING;
		this.ctx.error = null;
		try {
			const brutos = await this.#effects.executarBusca(this.ctx);
			this.ctx.dadosBrutos = brutos;
			this.#refiltrar();
			this.status = STATES.RESULTS;
		} catch (e) {
			this.ctx.error = e?.message ?? 'A busca falhou';
			this.status = STATES.ERROR;
		}
	}

	#refiltrar() {
		const r = montarResultados({
			fontes: this.ctx.fontes,
			dadosCuradoria: this.ctx.dadosBrutos.curadoria ?? [],
			dadosQuedas: this.ctx.dadosBrutos.quedas ?? [],
			dadosNovos: this.ctx.dadosBrutos.novos ?? [],
			dadosLojas: this.ctx.dadosBrutos.lojas ?? [],
			favoritos: this.ctx.dadosBrutos.favoritos ?? [],
			busca: this.ctx.keyword,
			categorias: this.ctx.categorias,
			comissaoMin: this.ctx.comissaoMin,
			vendasMin: this.ctx.vendasMin
		});
		this.ctx.resultados = r;
		this.ctx.contagens = {
			curadoria: r.filter((p) => p._fonte === 'curadoria').length,
			quedas: r.filter((p) => p._fonte === 'queda').length,
			novos: r.filter((p) => p._fonte === 'novo').length,
			lojas: r.filter((p) => p._fonte === 'loja').length
		};
	}
}

// Re-exports para a view
export { gerarLabelBusca, cronLabel, gerarResumo };
