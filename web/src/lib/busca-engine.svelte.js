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
	BUSCA_DUPLICADA,
	OMNIBOX,
	MARKETPLACES
} from './busca-config.js';
import { normalizarNome, matchLojas } from './loja-registry.js';
import { criarContextoInicial, criarUIInicial, guards, STATES, MODOS } from './busca-engine-state.js';
import { detectarIntencao } from './omnibox-intencao.js';
import { parsearInput } from './omnibox-parser.js';
import { gerarSugestoes } from './omnibox-sugestoes.js';
import { trace } from '@opentelemetry/api';

const tracer = trace.getTracer('busca-engine');

// Re-export para consumidores que importavam da engine.
export { STATES, MODOS, guards };

// ── Classe Engine ─────────────────────────────────────────────────────────
export class BuscaEngine {
	// Estado reativo (Svelte 5 compila para getter/setter)
	status = $state(STATES.IDLE);
	ctx = $state(criarContextoInicial());
	ui = $state(criarUIInicial());

	// UI state — delegam para ui.paineis
	get filtrosAberto() { return this.ui.paineis.filtrosAberto; }
	set filtrosAberto(v) { this.ui.paineis.filtrosAberto = v; }
	get salvarAberto() { return this.ui.paineis.salvarAberto; }
	set salvarAberto(v) { this.ui.paineis.salvarAberto = v; }
	get buscasPainelAberto() { return this.ui.paineis.buscasSalvasAberto; }
	set buscasPainelAberto(v) { this.ui.paineis.buscasSalvasAberto = v; }
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
	/** Intent de busca derivado do contexto (keyword x loja) — ver busca-config.js. */
	get intent() {
		return intentBusca(this.ctx);
	}

	// ── Derivados: Omnibox (sub-estado ui) ────────────────────────────────
	/** Estado completo do Omnibox para o componente renderizar. */
	get omnibox() { return this.ui.omnibox; }
	/** Modo de exibicao dos resultados ('produtos' | 'lojas'). */
	get modoResultados() { return this.ui.resultados.modo; }
	/** Resultados de busca por loja (quando modoResultados === 'lojas'). */
	get resultadosLojas() { return this.ui.resultados.lojas; }

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
		return this.ctx.shopIds.map((id) => {
			const meta = this.ctx.shopMeta[id] ?? { marketplace: 'shopee', origem: null, monitorada: false, cron: '' };
			return {
				id,
				nome: this.ctx.shopNomes[id] || id,
				...meta,
				tipo: meta.tipo ?? (meta.cron ? 'monitorada' : 'escopada')
			};
		});
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
			LIMPAR: () => this.#limpar(),
			// ── Smart Search handlers ──
			OMNIBOX_INPUT: (e) => this.#omniboxInput(e),
			OMNIBOX_KEYDOWN: (e) => this.#omniboxKeydown(e),
			OMNIBOX_SELECIONAR: (e) => this.#omniboxSelecionar(e),
			OMNIBOX_BLUR: () => this.#omniboxBlur(),
			BUSCAR_LOJAS: (e) => this.#buscarLojas(e),
			MONITORAR_LOJA: (e) => this.#monitorarLoja(e)
		};
	}

	/**
	 * Despacha um evento para a engine. Antes de executar o handler,
	 * avalia a transição de modo via regras declarativas.
	 * Cada evento gera um span OTel para observabilidade.
	 */
	async send(event) {
		const span = tracer.startSpan(`engine.${event.type}`, {
			attributes: {
				'engine.event_type': event.type,
				'engine.modo': this.ctx.modo,
				'engine.status': this.status
			}
		});

		try {
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

			const result = await this.#handlers[event.type]?.(event);
			span.setStatus({ code: 0 }); // OK
			return result;
		} catch (e) {
			span.setStatus({ code: 2, message: e?.message }); // ERROR
			throw e;
		} finally {
			span.end();
		}
	}

	// ── Transições privadas ─────────────────────────────────────────────────

	async #inicializar() {
		this.status = STATES.SEARCHING;
		try {
			const [buscas, categorias, lojas] = await Promise.all([
				this.#effects.carregarBuscasSalvas(),
				this.#effects.carregarCategorias(),
				this.#effects.carregarRegistroLojas?.() ?? Promise.resolve([])
			]);
			this.ctx.buscasSalvas = (buscas ?? []).map(payloadToConfig);
			this.ctx.categoriasDisponiveis = categorias ?? [];
			this.ctx.lojasDisponiveis = lojas ?? [];

			// Sincroniza o store externo para que executarBusca veja as buscas com lojas
			await this.#effects.sincronizarStoreExterno();

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
		if (this.ui.resultados.modo === 'lojas') {
			this.ui.resultados.modo = 'produtos';
		}
		this.#debounce();
	}

	/** Loja já-conhecida (dropdown ou match exato do registro): adiciona direto, sem resolver. */
	#adicionarLojaConhecida(loja) {
		// id é o ShopId (string na API); o escopo usa número, como o backend (Busca.ShopIds = long[]).
		const shopId = Number(loja.id);
		if (loja.id == null || Number.isNaN(shopId) || this.ctx.shopIds.includes(shopId)) return;
		const { nome, marketplace, origem, cron } = loja;
		this.ctx.shopIds = [...this.ctx.shopIds, shopId];
		this.ctx.shopNomes = { ...this.ctx.shopNomes, [shopId]: nome };
		this.ctx.shopMeta = {
			...this.ctx.shopMeta,
			[shopId]: {
				marketplace: marketplace ?? 'shopee',
				origem: origem ?? null,
				monitorada: Boolean(cron),
				cron: cron ?? '',
				tipo: cron ? 'monitorada' : 'escopada'
			}
		};
		return this.#executarBusca();
	}

	async #resolverLojaRemota(input) {
		if (!guards.resolucaoPermitida(this.ctx)) return;
		this.ctx.resolucaoLoja = { status: 'resolvendo' };
		// AbortController cancela o fetch no timeout (evita resolução órfã em background).
		const controller = new AbortController();
		const timer = setTimeout(() => controller.abort(), 10000);
		try {
			const r = await this.#effects.resolverLoja(input, controller.signal);
			// Anexa a loja resolvida ao registro local, se ainda não presente.
			if (!this.ctx.lojasDisponiveis.some((l) => String(l.id) === String(r.id))) {
				this.ctx.lojasDisponiveis = [...this.ctx.lojasDisponiveis, r];
			}
			this.ctx.resolucaoLoja = { status: 'idle' };
			return this.#adicionarLojaConhecida(r);
		} catch (e) {
			const abortou = e?.name === 'AbortError' || controller.signal.aborted;
			const erro = abortou ? 'Timeout na resolução da loja (10s).' : (e?.message ?? 'Falha ao resolver loja');
			this.ctx.resolucaoLoja = { status: 'erro', erro };
		} finally {
			clearTimeout(timer);
		}
	}

	async #adicionarLoja(event) {
		if (event.loja?.id != null) return this.#adicionarLojaConhecida(event.loja);

		const input = (event.value ?? '').trim();
		if (!guards.lojaInputValida(this.ctx, event)) return;

		// Nova tentativa limpa o erro de resolução anterior.
		if (this.ctx.resolucaoLoja.status === 'erro') this.ctx.resolucaoLoja = { status: 'idle' };

		// Match local exato no registro → adiciona sem rede. matchLojas retorna lojas cruas (sem `.meta`).
		const norm = normalizarNome(input);
		const [match] = matchLojas(norm, this.ctx.lojasDisponiveis, 1);
		if (match && norm === (match.nome_normalizado || normalizarNome(match.nome))) {
			return this.#adicionarLojaConhecida(match);
		}

		return this.#resolverLojaRemota(input);
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

	// ── Smart Search handlers (OMNIBOX_*, BUSCAR_LOJAS, MONITORAR_LOJA) ───

	/** Contexto para detectarIntencao — derivado do ctx atual. */
	get #intencaoCtx() {
		return {
			categoriasDisponiveis: this.ctx.categoriasDisponiveis,
			marketplacesFiltro: this.ctx.marketplacesFiltro,
			shopIds: this.ctx.shopIds,
			shopNomes: this.ctx.shopNomes
		};
	}

	#omniboxInput(event) {
		const value = event.value ?? '';
		this.ui.omnibox.inputValue = value;
		this.ui.omnibox.highlightIdx = -1;
		this.ui.omnibox.aberto = true;

		// Parsing e roteamento
		const tokens = parsearInput(value);
		const ultimoToken = tokens[tokens.length - 1];

		if (ultimoToken && ultimoToken.tipo !== 'keyword') {
			// Token com prefixo -> modo prefixo (sugestoes contextuais)
			this.ui.omnibox.modo = 'sugestoes';
			this.ui.omnibox.opcoes = this.#gerarSugestoesPrefixo(ultimoToken);
		} else {
			// Texto livre -> detecao de intencao
			this.ui.omnibox.modo = 'intencao';
			const textoLivre = tokens.filter((t) => t.tipo === 'keyword').map((t) => t.valor).join(' ');
			this.ui.omnibox.opcoes = detectarIntencao(textoLivre, this.#intencaoCtx);
		}

		// Keyword para a engine (debounce)
		const kw = tokens.filter((t) => t.tipo === 'keyword').map((t) => t.valor).join(' ');
		this.ctx.keyword = kw;
		this.#debounce();
	}

	#omniboxKeydown(event) {
		const { key, idx } = event;
		const opcoes = this.ui.omnibox.opcoes;
		const n = opcoes.length;

		if (key === 'highlight' && idx != null) {
			// Mouse hover highlight
			this.ui.omnibox.highlightIdx = idx;
		} else if (key === 'ArrowDown') {
			this.ui.omnibox.aberto = true;
			// -1 -> 0, then cycles 0..n-1
			const cur = this.ui.omnibox.highlightIdx;
			this.ui.omnibox.highlightIdx = n ? (cur + 1) % n : -1;
		} else if (key === 'ArrowUp') {
			this.ui.omnibox.aberto = true;
			const cur = this.ui.omnibox.highlightIdx;
			// -1 -> n-1 (last), then cycles n-1..0
			this.ui.omnibox.highlightIdx = n ? (cur <= 0 ? n - 1 : cur - 1) : -1;
		} else if (key === 'Enter') {
			this.#omniboxExecutar();
		} else if (key === 'Escape') {
			this.ui.omnibox.aberto = false;
			this.ui.omnibox.highlightIdx = -1;
		}
	}

	#omniboxSelecionar(event) {
		const { indice } = event;
		const opcoes = this.ui.omnibox.opcoes;
		if (indice >= 0 && indice < opcoes.length) {
			this.ui.omnibox.aberto = false;
			this.ui.omnibox.highlightIdx = -1;
			const opcao = opcoes[indice];
			if (this.ui.omnibox.modo === 'intencao') {
				this.#executarIntencao(opcao);
			} else {
				this.#executarSugestaoPrefixo(opcao);
			}
		}
	}

	#omniboxBlur() {
		this.ui.omnibox.aberto = false;
		this.ui.omnibox.highlightIdx = -1;
	}

	#omniboxExecutar() {
		const { opcoes, highlightIdx, modo } = this.ui.omnibox;
		if (!opcoes.length) return;

		const idx = highlightIdx >= 0 ? highlightIdx : 0;
		const opcao = opcoes[idx];

		this.ui.omnibox.aberto = false;
		this.ui.omnibox.highlightIdx = -1;

		if (modo === 'intencao') {
			this.#executarIntencao(opcao);
		} else {
			this.#executarSugestaoPrefixo(opcao);
		}
	}

	#executarIntencao(opcao) {
		switch (opcao.tipo) {
			case 'produtos':
				this.ctx.keyword = opcao.payload.keyword ?? '';
				this.ui.resultados.modo = 'produtos';
				this.#executarBusca();
				break;
			case 'lojas':
				this.#buscarLojas({ termo: opcao.payload.termo });
				break;
			case 'categoria':
				this.#adicionarCategoria({ nome: opcao.payload.categoria });
				this.ctx.keyword = '';
				this.ui.omnibox.inputValue = '';
				this.#executarBusca();
				break;
			case 'resolver_link':
				this.#adicionarLoja({ value: opcao.payload.url });
				this.ui.resultados.modo = 'lojas';
				break;
		}
	}

	#executarSugestaoPrefixo(sug) {
		if (!sug) return;
		switch (sug.tipo) {
			case 'loja':
				this.send({ type: 'ADICIONAR_LOJA', loja: sug.meta });
				break;
			case 'categoria':
				this.send({ type: 'ADICIONAR_CATEGORIA', nome: sug.label, categoria: sug.meta });
				break;
			case 'marketplace': {
				const mkts = [...new Set([...this.ctx.marketplacesFiltro, sug.meta.marketplace])];
				this.send({ type: 'MUDAR_MARKETPLACES', marketplaces: mkts });
				break;
			}
			case 'busca_salva':
				this.send({ type: 'CARREGAR_SALVA', config: sug.meta.config });
				this.ui.omnibox.inputValue = (sug.meta.config.keywords ?? [])[0] ?? '';
				break;
		}
		// Remove token ativo do input
		const tokens = parsearInput(this.ui.omnibox.inputValue);
		tokens.pop();
		this.ui.omnibox.inputValue = tokens.map((t) => {
			const pfx = t.tipo === 'keyword' ? '' : { loja: '@', categoria: '#', marketplace: '!' }[t.tipo] ?? '';
			return pfx + t.valor;
		}).join(' ');
	}

	/** Gera sugestoes por prefixo (@loja, #categoria, !marketplace) para o dropdown. */
	#gerarSugestoesPrefixo(ultimoToken) {
		const ctx = {
			lojasMonitoradas: this.ctx.lojasDisponiveis,
			categoriasDisponiveis: this.ctx.categoriasDisponiveis,
			marketplaces: MARKETPLACES?.suportados ?? ['shopee', 'mercado_livre', 'amazon'],
			buscasSalvas: this.ctx.buscasSalvas
		};
		const cfg = { minChars: OMNIBOX?.minChars ?? 2, maxSugestoes: OMNIBOX?.maxSugestoes ?? 7, matchBuscaSalva: OMNIBOX?.matchBuscaSalva ?? true };
		const sugestoesMap = gerarSugestoes(ultimoToken, ctx, cfg);
		// Flatten map to array for rendering
		return [...sugestoesMap.values()].flat();
	}

	async #buscarLojas(event) {
		const termo = (event.termo ?? '').trim();
		if (termo.length < 2) return;

		this.ui.resultados.modo = 'lojas';
		this.status = STATES.SEARCHING;

		try {
			const resultado = await this.#effects.buscarLojasPorNome(termo);
			this.ui.resultados.lojas = resultado?.lojas ?? [];
			this.status = STATES.RESULTS;
		} catch (e) {
			this.ctx.error = e?.message ?? 'Falha ao buscar lojas';
			this.ui.resultados.lojas = [];
			this.status = STATES.ERROR;
		}
	}

	async #monitorarLoja(event) {
		const loja = event.loja;
		if (!loja?.id) return;

		this.ctx.resolucaoLoja = { status: 'resolvendo' };
		try {
			// Reutiliza logica de adicionar loja conhecida
			await this.#adicionarLojaConhecida(loja);
			// Atualiza o card no resultados lojas para refletir o novo status
			this.ui.resultados.lojas = this.ui.resultados.lojas.map((l) =>
				String(l.id) === String(loja.id) ? { ...l, monitorada: true } : l
			);
			this.ctx.resolucaoLoja = { status: 'idle' };
		} catch (e) {
			this.ctx.resolucaoLoja = { status: 'erro', erro: e?.message ?? 'Falha ao monitorar loja' };
		}
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
