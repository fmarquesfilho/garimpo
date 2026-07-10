/**
 * BuscaEngine — Máquina de estados da página Garimpar (v3).
 *
 * Classe Svelte 5 com campos $state reativos. Toda transição de estado é
 * explícita via send(event). Guards impedem estados incoerentes.
 * Effects (API calls) são injetáveis para testabilidade.
 *
 * v3 adiciona:
 * - Modos de interação (explorando/vinculada/editando) — FSM declarativa
 * - Detecção reativa de busca duplicada (fingerprint)
 * - Guard de duplicata ao salvar
 * - Evento CANCELAR_EDICAO
 *
 * Uso:
 *   const engine = new BuscaEngine(effects);
 *   engine.send({ type: 'DIGITAR', value: 'serum' });
 *   // engine.status, engine.ctx, engine.modo, engine.buscaDuplicada são reativos
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
import {
	DEFAULTS,
	normalizarComissao,
	normalizarVendas,
	intentBusca,
	proximoModo,
	buscarDuplicada,
	BUSCA_DUPLICADA
} from './busca-config.js';
import { STATES, MODOS, criarContextoInicial, guards } from './busca-engine-state.js';

// Re-export para consumidores que importavam da engine.
export { STATES, MODOS, guards };

// ── Classe Engine ─────────────────────────────────────────────────────────
export class BuscaEngine {
	// Estado reativo (Svelte 5 compila para getter/setter)
	status = $state(STATES.IDLE);
	ctx = $state(criarContextoInicial());

	// UI state (não faz parte da FSM formal mas é reativo)
	filtrosAberto = $state(false);
	salvarAberto = $state(false);
	colapsado = $state(false);
	buscasPainelAberto = $state(false);

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

	// ── Derivados v3: modos e duplicata ───────────────────────────────────

	/** Modo de interação atual: 'explorando' | 'vinculada' | 'editando'. */
	get modo() {
		return this.ctx.modo;
	}

	/**
	 * Busca salva com os mesmos parâmetros de identidade que o contexto atual.
	 * Null se não há duplicata. Usado pela view para feedback reativo.
	 */
	get buscaDuplicada() {
		if (!BUSCA_DUPLICADA.feedbackReativo) return null;
		return buscarDuplicada(this.ctx, this.ctx.buscasSalvas, this.ctx.editandoId);
	}

	// ── Derivados para as raias (view) ────────────────────────────────────────

	/** Cards de categoria selecionada: `{ nome, marketplaces[] }[]`. */
	get categoriaCards() {
		return this.ctx.categorias.map((nome) => ({ nome, marketplaces: this.ctx.categoriaMeta[nome] ?? [] }));
	}

	/** Cards de loja no escopo: `{ id, nome, marketplace, origem, monitorada, cron }[]`. */
	get lojaCards() {
		return this.ctx.shopIds.map((id) => ({
			id,
			nome: this.ctx.shopNomes[id] || id,
			...(this.ctx.shopMeta[id] ?? { marketplace: 'shopee', origem: null, monitorada: false, cron: '' })
		}));
	}

	/** Contador da raia Filtros: fontes ativas não-default + quantitativos + categorias. */
	get contadorFiltros() {
		let n = this.ctx.categorias.length + this.ctx.marketplacesFiltro.length;
		if (this.ctx.comissaoMin !== DEFAULTS.comissaoMin) n++;
		if (this.ctx.vendasMin > 0) n++;
		// fontes que diferem do default
		for (const [k, v] of Object.entries(this.ctx.fontes)) {
			if (v !== DEFAULTS.fontes[k]) n++;
		}
		return n;
	}

	/** Contador da raia Lojas: lojas no escopo. */
	get contadorLojas() {
		return this.ctx.shopIds.length;
	}

	/** Contador de buscas salvas. */
	get contadorBuscas() {
		return this.ctx.buscasSalvas.length;
	}

	/** @type {import('./busca-engine-effects.js').Effects} */
	#effects;
	#debounceTimer = null;

	constructor(effects) {
		this.#effects = effects;
	}

	// ── API pública ─────────────────────────────────────────────────────────
	/** Mapa event.type → handler. Mantém `send` com baixa complexidade. */
	get #handlers() {
		return {
			INICIALIZAR: () => this.#inicializar(),
			DIGITAR: (e) => this.#digitar(e),
			ADICIONAR_LOJA: (e) => this.#adicionarLoja(e),
			REMOVER_LOJA: (e) => this.#removerLoja(e),
			ADICIONAR_CATEGORIA: (e) => this.#adicionarCategoria(e),
			REMOVER_CATEGORIA: (e) => this.#removerCategoria(e),
			MUDAR_FILTRO: (e) => this.#mudarFiltro(e),
			MUDAR_FONTES: (e) => this.#mudarFontes(e),
			MUDAR_MARKETPLACES: (e) => this.#mudarMarketplaces(e),
			SALVAR: () => this.#salvar(),
			CARREGAR_SALVA: (e) => this.#carregarSalva(e),
			EDITAR_SALVA: (e) => this.#editarSalva(e),
			REMOVER_SALVA: (e) => this.#removerSalva(e),
			CANCELAR_EDICAO: () => this.#cancelarEdicao(),
			RETRY: () => this.#executarBusca(),
			LIMPAR: () => this.#limpar()
		};
	}

	/**
	 * Despacha um evento para a engine. Antes de executar o handler,
	 * avalia a transição de modo via regras declarativas.
	 */
	async send(event) {
		// Transição de modo (declarativa, via rules)
		const novoModo = proximoModo(this.ctx.modo, event.type);
		if (novoModo !== this.ctx.modo) {
			this.ctx.modo = novoModo;
			// Ao voltar para explorando via desvinculação, limpar vínculo
			if (novoModo === MODOS.EXPLORANDO && event.type !== 'SALVAR') {
				this.ctx.buscaSelecionadaId = null;
				this.ctx.editandoId = null;
			}
		}

		// Limpar erro de duplicata a cada interação
		this.ctx.erroDuplicata = null;

		return this.#handlers[event.type]?.(event);
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

			// Sincroniza o store externo para que executarBusca veja as buscas com lojas
			await this.#effects.sincronizarStoreExterno();

			// Lojas monitoradas para o autocomplete da raia Lojas (deriva das buscas salvas)
			this.ctx.lojasDisponiveis = this.#effects.listarLojasMonitoradas?.() ?? [];

			// Só executa busca automática se há contexto explícito (keyword, loja no escopo, categoria).
			// Sem contexto: fica em IDLE. O usuário inicia a busca ao digitar, clicar pill, ou adicionar loja.
			if (this.ctx.keyword.trim() || this.ctx.shopIds.length > 0 || this.ctx.categorias.length > 0) {
				await this.#executarBusca();
			} else {
				this.status = STATES.IDLE;
			}
		} catch (e) {
			this.ctx.error = e?.message ?? 'Falha ao inicializar';
			this.status = STATES.ERROR;
		}
	}

	#digitar(event) {
		this.ctx.keyword = event.value ?? '';
		this.#debounce();
	}

	/** Loja já-monitorada escolhida no dropdown: adiciona direto, sem resolver. */
	#adicionarLojaMonitorada(loja) {
		const { id, nome, marketplace, origem, monitorada, cron } = loja;
		if (this.ctx.shopIds.includes(id)) return;
		this.ctx.shopIds = [...this.ctx.shopIds, id];
		this.ctx.shopNomes = { ...this.ctx.shopNomes, [id]: nome };
		this.ctx.shopMeta = {
			...this.ctx.shopMeta,
			[id]: {
				marketplace: marketplace ?? 'shopee',
				origem: origem ?? null,
				monitorada: Boolean(monitorada),
				cron: cron ?? ''
			}
		};
		return this.#executarBusca();
	}

	async #adicionarLoja(event) {
		if (event.loja?.id) return this.#adicionarLojaMonitorada(event.loja);
		if (!guards.lojaInputValida(this.ctx, event)) return;
		this.ctx.lojaResolvendo = true;
		this.ctx.lojaErro = '';
		try {
			const r = await this.#effects.resolverLoja(event.value);
			if (r.shop_ids?.length) {
				const id = r.shop_ids[0];
				this.ctx.shopIds = [...this.ctx.shopIds, ...r.shop_ids];
				this.ctx.shopNomes = { ...this.ctx.shopNomes, [id]: r.keyword };
				this.ctx.shopMeta = {
					...this.ctx.shopMeta,
					[id]: {
						marketplace: r.marketplace ?? event.marketplace ?? 'shopee',
						origem: r.origem ?? r.origem_padrao ?? event.origem ?? null,
						monitorada: Boolean(r.cron),
						cron: r.cron ?? ''
					}
				};
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
		const meta = { ...this.ctx.shopMeta };
		delete meta[event.shopId];
		this.ctx.shopMeta = meta;
		this.#debounce();
	}

	#adicionarCategoria(event) {
		const nome = event.nome ?? event.categoria?.nome;
		if (!nome || this.ctx.categorias.includes(nome)) return;
		this.ctx.categorias = [...this.ctx.categorias, nome];
		if (event.categoria?.marketplaces) {
			this.ctx.categoriaMeta = { ...this.ctx.categoriaMeta, [nome]: event.categoria.marketplaces };
		}
		// Categoria é contexto de busca → refetch imediato (ver rules.transicoes)
		this.#executarBusca();
	}

	#removerCategoria(event) {
		const nome = event.nome;
		this.ctx.categorias = this.ctx.categorias.filter((c) => c !== nome);
		const meta = { ...this.ctx.categoriaMeta };
		delete meta[nome];
		this.ctx.categoriaMeta = meta;
		this.#debounce();
	}

	#mudarMarketplaces(event) {
		this.ctx.marketplacesFiltro = event.marketplaces ?? [];
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

		// Guard v3: detecção de busca duplicada
		if (BUSCA_DUPLICADA.erroAoSalvar) {
			const dup = guards.buscaDuplicada(this.ctx);
			if (dup) {
				this.ctx.erroDuplicata = `Já existe uma busca salva com os mesmos parâmetros: "${gerarLabelBusca(dup)}"`;
				return;
			}
		}

		this.status = STATES.SAVING;
		try {
			const payload = configToPayload({
				// editandoId presente → sincronizarBusca faz update in-place (mesmo id)
				id: this.ctx.editandoId ?? undefined,
				keywords: this.ctx.keyword ? [this.ctx.keyword] : [],
				shopIds: this.ctx.shopIds,
				shopNomes: this.ctx.shopNomes,
				comissaoMin: this.ctx.comissaoMin,
				vendasMin: this.ctx.vendasMin,
				categorias: this.ctx.categorias,
				fontes: this.fontesAtivas,
				cron: this.ctx.cron,
				marketplaces: this.ctx.marketplacesFiltro.length ? this.ctx.marketplacesFiltro : 'shopee'
			});
			await this.#effects.salvarBusca(payload);
			// Sincroniza store externo (para que executarBusca veja a nova loja)
			await this.#effects.sincronizarStoreExterno();
			// Recarregar lista de buscas salvas
			const buscas = await this.#effects.carregarBuscasSalvas();
			this.ctx.buscasSalvas = (buscas ?? []).map(payloadToConfig);
			this.ctx.editandoId = null;
			this.ctx.buscaSelecionadaId = null;
			this.ctx.modo = MODOS.EXPLORANDO;
			this.salvarAberto = false;
			this.status = STATES.RESULTS;
		} catch (e) {
			this.ctx.error = e?.message ?? 'Falha ao salvar';
			this.status = STATES.RESULTS;
		}
	}

	#carregarSalva(event) {
		this.#restaurarConfig(event.config);
		// Carregar (rodar): modo vinculada, sem edit mode
		this.ctx.buscaSelecionadaId = event.config.id ?? null;
		this.ctx.editandoId = null;
		// Modo já foi transicionado no send() via proximoModo()
		this.#executarBusca();
	}

	/** Edit mode: restaura a config E marca o id para update in-place ao salvar. */
	#editarSalva(event) {
		this.#restaurarConfig(event.config);
		this.ctx.buscaSelecionadaId = event.config.id ?? null;
		this.ctx.editandoId = event.config.id ?? null;
		// Modo já foi transicionado no send() via proximoModo()
		this.salvarAberto = true;
		this.#executarBusca();
	}

	/** Cancela edit mode: reseta vínculo e fecha painel salvar. */
	#cancelarEdicao() {
		this.ctx.editandoId = null;
		this.ctx.buscaSelecionadaId = null;
		// Modo já foi transicionado no send() via proximoModo()
		this.salvarAberto = false;
	}

	/** Restaura o contexto a partir de uma config salva (compartilhado por carregar/editar). */
	#restaurarConfig(config) {
		this.ctx.keyword = (config.keywords ?? [])[0] ?? '';
		this.ctx.shopIds = config.shopIds ?? [];
		this.ctx.shopNomes = config.shopNomes ?? {};
		this.ctx.comissaoMin = config.comissaoMin || 0.07;
		this.ctx.vendasMin = config.vendasMin || 0;
		this.ctx.categorias = config.categorias ?? [];
		this.ctx.marketplacesFiltro = Array.isArray(config.marketplaces) ? config.marketplaces : [];
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
	}

	async #removerSalva(event) {
		await this.#effects.removerBusca(event.config);
		await this.#effects.sincronizarStoreExterno();
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
