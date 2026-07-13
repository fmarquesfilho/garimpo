/**
 * BuscaEngine — Headless UI Controller da pagina Garimpar.
 *
 * Responsabilidade unica: orquestrar estado reativo + despachar eventos.
 * Logica de dominio delegada para modulos coesos:
 *   - busca-engine-omnibox.js (smart search, keyboard, intencao)
 *   - busca-engine-lojas.js (adicionar/remover/resolver lojas)
 *   - busca-engine-persistencia.js (salvar/carregar/editar buscas)
 *   - busca-engine-state.js (estado inicial, guards)
 *   - busca-engine-effects.js (side effects — API calls)
 */

import { montarResultados } from './descobrir-logic.js';
import {
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
	MARKETPLACES
} from './busca-config.js';
import { criarContextoInicial, criarUIInicial, guards, STATES, MODOS } from './busca-engine-state.js';
import {
	processarOmniboxInput,
	processarOmniboxKeydown,
	resolverOpcaoAtiva,
	classificarIntencao,
	classificarSugestaoPrefixo
} from './busca-engine-omnibox.js';
import { classificarAdicionarLoja, prepararAdicionarLojaConhecida, prepararRemoverLoja } from './busca-engine-lojas.js';
import { restaurarConfig, montarPayloadSalvar, validarAntesDeSalvar } from './busca-engine-persistencia.js';
import { trace } from '@opentelemetry/api';

const tracer = trace.getTracer('busca-engine');

export { STATES, MODOS, guards, gerarLabelBusca, cronLabel, gerarResumo };

// ── Engine ────────────────────────────────────────────────────────────────
export class BuscaEngine {
	status = $state(STATES.IDLE);
	ctx = $state(criarContextoInicial());
	ui = $state(criarUIInicial());
	colapsado = $state(false);

	// Paineis (accessors para compatibilidade com componentes)
	get filtrosAberto() {
		return this.ui.paineis.filtrosAberto;
	}
	set filtrosAberto(v) {
		this.ui.paineis.filtrosAberto = v;
	}
	get salvarAberto() {
		return this.ui.paineis.salvarAberto;
	}
	set salvarAberto(v) {
		this.ui.paineis.salvarAberto = v;
	}
	get buscasPainelAberto() {
		return this.ui.paineis.buscasSalvasAberto;
	}
	set buscasPainelAberto(v) {
		this.ui.paineis.buscasSalvasAberto = v;
	}

	// ── Derivados ─────────────────────────────────────────────────────────
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
	get intent() {
		return intentBusca(this.ctx);
	}
	get omnibox() {
		return this.ui.omnibox;
	}
	get modoResultados() {
		return this.ui.resultados.modo;
	}
	get resultadosLojas() {
		return this.ui.resultados.lojas;
	}
	get modo() {
		return this.ctx.modo;
	}
	get contadorLojas() {
		return this.ctx.shopIds.length;
	}
	get contadorBuscas() {
		return this.ctx.buscasSalvas.length;
	}
	get buscaDuplicada() {
		if (!BUSCA_DUPLICADA.feedbackReativo) return null;
		return buscarDuplicada(this.ctx, this.ctx.buscasSalvas, this.ctx.editandoId);
	}
	get categoriaCards() {
		return this.ctx.categorias.map((nome) => ({ nome, marketplaces: this.ctx.categoriaMeta[nome] ?? [] }));
	}
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
	get contadorFiltros() {
		let n = this.ctx.categorias.length + this.ctx.marketplacesFiltro.length;
		if (this.ctx.comissaoMin !== DEFAULTS.comissaoMin) n++;
		if (this.ctx.vendasMin > 0) n++;
		for (const [k, v] of Object.entries(this.ctx.fontes)) {
			if (v !== DEFAULTS.fontes[k]) n++;
		}
		return n;
	}

	#effects;
	#debounceTimer = null;
	constructor(effects) {
		this.#effects = effects;
	}

	// ── Dispatch ──────────────────────────────────────────────────────────
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
			OMNIBOX_INPUT: (e) => this.#omniboxInput(e),
			OMNIBOX_KEYDOWN: (e) => this.#omniboxKeydown(e),
			OMNIBOX_SELECIONAR: (e) => this.#omniboxSelecionar(e),
			OMNIBOX_BLUR: () => this.#omniboxBlur(),
			BUSCAR_LOJAS: (e) => this.#buscarLojas(e),
			MONITORAR_LOJA: (e) => this.#monitorarLoja(e)
		};
	}

	async send(event) {
		const span = tracer.startSpan(`engine.${event.type}`, {
			attributes: { 'engine.event_type': event.type, 'engine.modo': this.ctx.modo, 'engine.status': this.status }
		});
		try {
			const novoModo = proximoModo(this.ctx.modo, event.type);
			if (novoModo !== this.ctx.modo) {
				this.ctx.modo = novoModo;
				if (novoModo === MODOS.EXPLORANDO && event.type !== 'SALVAR') {
					this.ctx.buscaSelecionadaId = null;
					this.ctx.editandoId = null;
				}
			}
			this.ctx.erroDuplicata = null;
			const result = await this.#handlers[event.type]?.(event);
			span.setStatus({ code: 0 });
			return result;
		} catch (e) {
			span.setStatus({ code: 2, message: e?.message });
			throw e;
		} finally {
			span.end();
		}
	}

	// ── Inicializacao ─────────────────────────────────────────────────────
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
			await this.#effects.sincronizarStoreExterno();
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

	// ── Busca (keyword/filtros) ───────────────────────────────────────────
	#digitar(event) {
		this.ctx.keyword = event.value ?? '';
		if (this.ui.resultados.modo === 'lojas') this.ui.resultados.modo = 'produtos';
		this.#debounce();
	}

	#mudarFiltro(event) {
		if ('comissaoMin' in event) this.ctx.comissaoMin = normalizarComissao(event.comissaoMin);
		if ('vendasMin' in event) this.ctx.vendasMin = normalizarVendas(event.vendasMin);
		if ('categorias' in event) this.ctx.categorias = event.categorias ?? [];
		this.#refiltrar();
	}

	#mudarFontes(event) {
		this.ctx.fontes = event.fontes;
		this.#debounce();
	}
	#mudarMarketplaces(event) {
		this.ctx.marketplacesFiltro = event.marketplaces ?? [];
		this.#debounce();
	}

	#adicionarCategoria(event) {
		const nome = event.nome ?? event.categoria?.nome;
		if (!nome || this.ctx.categorias.includes(nome)) return;
		this.ctx.categorias = [...this.ctx.categorias, nome];
		if (event.categoria?.marketplaces)
			this.ctx.categoriaMeta = { ...this.ctx.categoriaMeta, [nome]: event.categoria.marketplaces };
		this.#executarBusca();
	}

	#removerCategoria(event) {
		this.ctx.categorias = this.ctx.categorias.filter((c) => c !== event.nome);
		const meta = { ...this.ctx.categoriaMeta };
		delete meta[event.nome];
		this.ctx.categoriaMeta = meta;
		this.ui.omnibox.chipRemovalMessage = event.nome ? `Categoria ${event.nome} removida do escopo` : '';
		this.#debounce();
	}

	// ── Lojas (delega para busca-engine-lojas.js) ─────────────────────────
	#adicionarLojaConhecida(loja) {
		const mutacoes = prepararAdicionarLojaConhecida(loja, this.ctx);
		if (!mutacoes) return;
		Object.assign(this.ctx, mutacoes);
		return this.#executarBusca();
	}

	async #adicionarLoja(event) {
		const decisao = classificarAdicionarLoja(event, this.ctx, this.ctx.lojasDisponiveis);
		if (decisao.tipo === 'conhecida') return this.#adicionarLojaConhecida(decisao.loja);
		if (decisao.tipo === 'resolver') return this.#resolverLojaRemota(decisao.input);
	}

	async #resolverLojaRemota(input) {
		if (!guards.resolucaoPermitida(this.ctx)) return;
		this.ctx.resolucaoLoja = { status: 'resolvendo' };
		if (this.ctx.resolucaoLoja.status === 'erro') this.ctx.resolucaoLoja = { status: 'idle' };
		const controller = new AbortController();
		const timer = setTimeout(() => controller.abort(), 10000);
		try {
			const r = await this.#effects.resolverLoja(input, controller.signal);
			if (!this.ctx.lojasDisponiveis.some((l) => String(l.id) === String(r.id))) {
				this.ctx.lojasDisponiveis = [...this.ctx.lojasDisponiveis, r];
			}
			this.ctx.resolucaoLoja = { status: 'idle' };
			return this.#adicionarLojaConhecida(r);
		} catch (e) {
			const abortou = e?.name === 'AbortError' || controller.signal.aborted;
			this.ctx.resolucaoLoja = {
				status: 'erro',
				erro: abortou ? 'Timeout na resolucao da loja (10s).' : (e?.message ?? 'Falha ao resolver loja')
			};
		} finally {
			clearTimeout(timer);
		}
	}

	#removerLoja(event) {
		const nome = this.ctx.shopNomes[event.shopId] || '';
		Object.assign(this.ctx, prepararRemoverLoja(event.shopId, this.ctx));
		this.ui.omnibox.chipRemovalMessage = nome ? `Loja ${nome} removida do escopo` : '';
		this.#debounce();
	}

	// ── Buscas salvas (delega para busca-engine-persistencia.js) ──────────
	async #salvar() {
		const validacao = validarAntesDeSalvar(this.ctx);
		if (!validacao.ok) {
			this.ctx.erroDuplicata = validacao.erro ?? null;
			return;
		}
		this.status = STATES.SAVING;
		try {
			const payload = montarPayloadSalvar(this.ctx, this.fontesAtivas);
			await this.#effects.salvarBusca(payload);
			await this.#effects.sincronizarStoreExterno();
			this.ctx.buscasSalvas = ((await this.#effects.carregarBuscasSalvas()) ?? []).map(payloadToConfig);
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
		Object.assign(this.ctx, restaurarConfig(event.config));
		this.ctx.buscaSelecionadaId = event.config.id ?? null;
		this.ctx.editandoId = null;
		this.#executarBusca();
	}

	#editarSalva(event) {
		Object.assign(this.ctx, restaurarConfig(event.config));
		this.ctx.buscaSelecionadaId = event.config.id ?? null;
		this.ctx.editandoId = event.config.id ?? null;
		this.salvarAberto = true;
		this.#executarBusca();
	}

	#cancelarEdicao() {
		this.ctx.editandoId = null;
		this.ctx.buscaSelecionadaId = null;
		this.salvarAberto = false;
	}

	async #removerSalva(event) {
		await this.#effects.removerBusca(event.config);
		await this.#effects.sincronizarStoreExterno();
		this.ctx.buscasSalvas = ((await this.#effects.carregarBuscasSalvas()) ?? []).map(payloadToConfig);
	}

	#limpar() {
		const keep = { categoriasDisponiveis: this.ctx.categoriasDisponiveis, buscasSalvas: this.ctx.buscasSalvas };
		Object.assign(this.ctx, criarContextoInicial(), keep);
		this.status = STATES.IDLE;
	}

	// ── Omnibox (delega para busca-engine-omnibox.js) ─────────────────────
	get #intencaoCtx() {
		return {
			categoriasDisponiveis: this.ctx.categoriasDisponiveis,
			marketplacesFiltro: this.ctx.marketplacesFiltro,
			shopIds: this.ctx.shopIds,
			shopNomes: this.ctx.shopNomes
		};
	}
	get #sugestoesCtx() {
		return {
			lojasMonitoradas: this.ctx.lojasDisponiveis,
			categoriasDisponiveis: this.ctx.categoriasDisponiveis,
			marketplaces: MARKETPLACES?.suportados ?? [],
			buscasSalvas: this.ctx.buscasSalvas
		};
	}

	#omniboxInput(event) {
		const r = processarOmniboxInput(event.value ?? '', this.#intencaoCtx, this.#sugestoesCtx);
		Object.assign(this.ui.omnibox, {
			inputValue: r.inputValue,
			aberto: r.aberto,
			highlightIdx: r.highlightIdx,
			modo: r.modo,
			opcoes: r.opcoes
		});
		this.ctx.keyword = r.keyword;
		this.#debounce();
	}

	#omniboxKeydown(event) {
		const r = processarOmniboxKeydown(
			event.key,
			event.idx,
			this.ui.omnibox.highlightIdx,
			this.ui.omnibox.opcoes.length
		);
		if (r.highlightIdx != null) this.ui.omnibox.highlightIdx = r.highlightIdx;
		if (r.aberto != null) this.ui.omnibox.aberto = r.aberto;
		if (r.executar) this.#omniboxExecutar();
	}

	#omniboxSelecionar(event) {
		const opcoes = this.ui.omnibox.opcoes;
		if (event.indice >= 0 && event.indice < opcoes.length) {
			this.ui.omnibox.aberto = false;
			this.ui.omnibox.highlightIdx = -1;
			this.#executarOpcao(opcoes[event.indice]);
		}
	}

	#omniboxBlur() {
		this.ui.omnibox.aberto = false;
		this.ui.omnibox.highlightIdx = -1;
	}

	#omniboxExecutar() {
		const opcao = resolverOpcaoAtiva(this.ui.omnibox.opcoes, this.ui.omnibox.highlightIdx);
		if (!opcao) return;
		this.ui.omnibox.aberto = false;
		this.ui.omnibox.highlightIdx = -1;
		this.#executarOpcao(opcao);
	}

	#executarOpcao(opcao) {
		if (this.ui.omnibox.modo === 'intencao') {
			const { action, payload } = classificarIntencao(opcao);
			if (action === 'BUSCAR_PRODUTOS') {
				this.ctx.keyword = payload.keyword;
				this.ui.resultados.modo = 'produtos';
				this.#executarBusca();
			} else if (action === 'BUSCAR_LOJAS') {
				this.#buscarLojas(payload);
			} else if (action === 'ADICIONAR_CATEGORIA') {
				this.#adicionarCategoria({ nome: payload.nome });
				this.ctx.keyword = '';
				this.ui.omnibox.inputValue = '';
				this.#executarBusca();
			} else if (action === 'RESOLVER_LINK') {
				this.#adicionarLoja({ value: payload.url });
				this.ui.resultados.modo = 'lojas';
			}
		} else {
			const { action, payload, novoInput } = classificarSugestaoPrefixo(opcao, this.ui.omnibox.inputValue);
			this.ui.omnibox.inputValue = novoInput;
			if (action === 'ADICIONAR_LOJA') this.send({ type: 'ADICIONAR_LOJA', loja: payload.loja });
			else if (action === 'ADICIONAR_CATEGORIA')
				this.send({ type: 'ADICIONAR_CATEGORIA', nome: payload.nome, categoria: payload.categoria });
			else if (action === 'MUDAR_MARKETPLACES')
				this.send({
					type: 'MUDAR_MARKETPLACES',
					marketplaces: [...new Set([...this.ctx.marketplacesFiltro, payload.marketplace])]
				});
			else if (action === 'CARREGAR_SALVA') this.send({ type: 'CARREGAR_SALVA', config: payload.config });
		}
	}

	async #buscarLojas(event) {
		const termo = (event.termo ?? '').trim();
		if (termo.length < 2) return;
		this.ui.resultados.modo = 'lojas';
		this.status = STATES.SEARCHING;
		try {
			const r = await this.#effects.buscarLojasPorNome(termo);
			this.ui.resultados.lojas = r?.lojas ?? [];
			this.status = STATES.RESULTS;
		} catch (e) {
			this.ctx.error = e?.message ?? 'Falha ao buscar lojas';
			this.ui.resultados.lojas = [];
			this.status = STATES.ERROR;
		}
	}

	async #monitorarLoja(event) {
		if (!event.loja?.id) return;
		this.ctx.resolucaoLoja = { status: 'resolvendo' };
		try {
			await this.#adicionarLojaConhecida(event.loja);
			this.ui.resultados.lojas = this.ui.resultados.lojas.map((l) =>
				String(l.id) === String(event.loja.id) ? { ...l, monitorada: true } : l
			);
			this.ctx.resolucaoLoja = { status: 'idle' };
		} catch (e) {
			this.ctx.resolucaoLoja = { status: 'erro', erro: e?.message ?? 'Falha ao monitorar loja' };
		}
	}

	// ── Core (busca + debounce + refiltrar) ───────────────────────────────
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
			this.ctx.dadosBrutos = await this.#effects.executarBusca(this.ctx);
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
