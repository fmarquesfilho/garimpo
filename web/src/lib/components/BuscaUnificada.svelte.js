/**
 * Estado reativo e handlers do componente BuscaUnificada.
 * Módulo .svelte.js — runes ($state, $derived) funcionam aqui.
 * Dividido em funções menores para respeitar max-lines-per-function (80).
 */
import { buscasSalvas } from '$lib/buscas.js';
import { favoritos } from '$lib/favoritos.js';
import { carregarFontes } from '$lib/descobrir.js';
import { montarResultados, buildFonteOpcoes } from '$lib/descobrir-logic.js';
import { adicionarLoja, sincronizarBusca, listarBuscasServidor } from '$lib/api.js';
import {
	configToPayload,
	payloadToConfig,
	gerarResumo,
	contarFiltrosAtivos,
	gerarLabelBusca,
	cronLabel
} from '$lib/busca-unificada-logic.js';

export { gerarLabelBusca, cronLabel };

/** Opções estáticas de comissão para o Select. */
export const comissaoOpcoes = [0.05, 0.07, 0.1, 0.15].map((c) => ({ value: String(c), label: `${c * 100}%` }));

/** Cria o estado reativo da busca unificada. */
export function criarEstado() {
	let busca = $state('');
	let keywords = $state([]);
	let shopIds = $state([]);
	let shopNomes = $state({});
	let comissaoMin = $state(0.07);
	let vendasMin = $state(0);
	let categorias = $state([]);
	let fontes = $state({ curadoria: true, quedas: true, novos: true, favoritos: false, lojas: false });
	let cron = $state('');
	let colapsado = $state(false);
	let filtrosAberto = $state(false);
	let salvarAberto = $state(false);
	let lojaInput = $state('');
	let lojaResolvendo = $state(false);
	let lojaErro = $state('');
	let buscasSalvasLista = $state([]);

	// Contagens de resultados por fonte (atualizadas após cada execução)
	let contagens = $state({ curadoria: 0, quedas: 0, novos: 0, lojas: 0 });

	return {
		get busca() {
			return busca;
		},
		set busca(v) {
			busca = v;
		},
		get keywords() {
			return keywords;
		},
		set keywords(v) {
			keywords = v;
		},
		get shopIds() {
			return shopIds;
		},
		set shopIds(v) {
			shopIds = v;
		},
		get shopNomes() {
			return shopNomes;
		},
		set shopNomes(v) {
			shopNomes = v;
		},
		get comissaoMin() {
			return comissaoMin;
		},
		set comissaoMin(v) {
			comissaoMin = v;
		},
		get vendasMin() {
			return vendasMin;
		},
		set vendasMin(v) {
			vendasMin = v;
		},
		get categorias() {
			return categorias;
		},
		set categorias(v) {
			categorias = v;
		},
		get fontes() {
			return fontes;
		},
		set fontes(v) {
			fontes = v;
		},
		get cron() {
			return cron;
		},
		set cron(v) {
			cron = v;
		},
		get colapsado() {
			return colapsado;
		},
		set colapsado(v) {
			colapsado = v;
		},
		get filtrosAberto() {
			return filtrosAberto;
		},
		set filtrosAberto(v) {
			filtrosAberto = v;
		},
		get salvarAberto() {
			return salvarAberto;
		},
		set salvarAberto(v) {
			salvarAberto = v;
		},
		get lojaInput() {
			return lojaInput;
		},
		set lojaInput(v) {
			lojaInput = v;
		},
		get lojaResolvendo() {
			return lojaResolvendo;
		},
		set lojaResolvendo(v) {
			lojaResolvendo = v;
		},
		get lojaErro() {
			return lojaErro;
		},
		set lojaErro(v) {
			lojaErro = v;
		},
		get buscasSalvasLista() {
			return buscasSalvasLista;
		},
		set buscasSalvasLista(v) {
			buscasSalvasLista = v;
		},
		get contagens() {
			return contagens;
		},
		set contagens(v) {
			contagens = v;
		}
	};
}

/** Cria derivados reativos a partir do estado. */
export function criarDerivados(s, getBuscasSalvas, getFavoritos) {
	const buscasComLojas = $derived((getBuscasSalvas() ?? []).filter((b) => b.shop_ids?.length > 0));
	const nomesLojas = $derived(Object.fromEntries(buscasComLojas.map((b) => [b.id, b.nome || b.id])));
	const filtrosAtivos = $derived(
		contarFiltrosAtivos({ comissaoMin: s.comissaoMin, vendasMin: s.vendasMin, categorias: s.categorias })
	);
	const fontesAtivas = $derived(
		Object.entries(s.fontes)
			.filter(([, v]) => v)
			.map(([k]) => k)
	);
	const resumo = $derived(
		gerarResumo({
			keywords: s.keywords,
			shopIds: s.shopIds,
			comissaoMin: s.comissaoMin,
			vendasMin: s.vendasMin,
			categorias: s.categorias,
			cron: s.cron
		})
	);
	const keywordStr = $derived(s.keywords.length > 0 ? s.keywords[0] : s.busca);
	const fonteOpcoes = $derived(
		buildFonteOpcoes({
			contagemCuradoria: s.contagens.curadoria,
			contagemQuedas: s.contagens.quedas,
			contagemNovos: s.contagens.novos,
			contagemLojas: s.contagens.lojas,
			totalFavoritos: (getFavoritos() ?? []).length
		})
	);

	return {
		get buscasComLojas() {
			return buscasComLojas;
		},
		get nomesLojas() {
			return nomesLojas;
		},
		get filtrosAtivos() {
			return filtrosAtivos;
		},
		get fontesAtivas() {
			return fontesAtivas;
		},
		get resumo() {
			return resumo;
		},
		get keywordStr() {
			return keywordStr;
		},
		get fonteOpcoes() {
			return fonteOpcoes;
		},
		get favoritos() {
			return getFavoritos() ?? [];
		}
	};
}

/** Cria handlers de ação que operam sobre estado + derivados. */
export function criarHandlers(s, d, { onresultados, oncarregando, onerro }) {
	async function executar() {
		oncarregando?.(true);
		onerro?.(null);
		try {
			const kw = d.keywordStr.trim();
			const r = await carregarFontes({
				fontes: s.fontes,
				busca: kw,
				comissaoMin: s.comissaoMin,
				categorias: s.categorias,
				buscasComLojas: d.buscasComLojas,
				nomesLojas: d.nomesLojas
			});
			const resultados = montarResultados({
				fontes: s.fontes,
				dadosCuradoria: r.curadoria,
				dadosQuedas: r.quedas,
				dadosNovos: r.novos,
				dadosLojas: r.lojas,
				favoritos: d.favoritos,
				busca: kw,
				categorias: s.categorias,
				comissaoMin: s.comissaoMin,
				vendasMin: s.vendasMin
			});
			onresultados?.(resultados);
			s.contagens = {
				curadoria: resultados.filter((r) => r._fonte === 'curadoria').length,
				quedas: resultados.filter((r) => r._fonte === 'queda').length,
				novos: resultados.filter((r) => r._fonte === 'novo').length,
				lojas: resultados.filter((r) => r._fonte === 'loja').length
			};
		} catch (e) {
			onerro?.(e);
		} finally {
			oncarregando?.(false);
		}
	}

	async function inicializar() {
		await buscasSalvas.sincronizarDoServidor();
		favoritos.sincronizar();
		await carregarBuscasSalvas();
		executar();
	}

	async function carregarBuscasSalvas() {
		try {
			const r = await listarBuscasServidor();
			s.buscasSalvasLista = (r?.buscas ?? []).map(payloadToConfig);
		} catch {
			/* offline */
		}
	}

	function handleFontesChange(v) {
		s.fontes = {
			curadoria: v.includes('curadoria'),
			quedas: v.includes('quedas'),
			novos: v.includes('novos'),
			favoritos: v.includes('favoritos'),
			lojas: v.includes('lojas')
		};
	}

	async function handleAdicionarLoja() {
		if (!s.lojaInput.trim()) return;
		s.lojaResolvendo = true;
		s.lojaErro = '';
		try {
			const r = await adicionarLoja({ input: s.lojaInput.trim() });
			if (r.shop_ids?.length) {
				s.shopIds = [...s.shopIds, ...r.shop_ids];
				s.shopNomes = { ...s.shopNomes, [r.shop_ids[0]]: r.keyword };
			}
			s.lojaInput = '';
			await buscasSalvas.sincronizarDoServidor();
		} catch (e) {
			s.lojaErro = e?.message || 'Falha ao resolver loja.';
		} finally {
			s.lojaResolvendo = false;
		}
	}

	function handleRemoverLoja(id) {
		s.shopIds = s.shopIds.filter((x) => x !== id);
		const n = { ...s.shopNomes };
		delete n[id];
		s.shopNomes = n;
	}

	async function handleSalvar() {
		const payload = configToPayload({
			keywords: s.keywords,
			shopIds: s.shopIds,
			comissaoMin: s.comissaoMin,
			vendasMin: s.vendasMin,
			categorias: s.categorias,
			fontes: d.fontesAtivas,
			cron: s.cron,
			marketplaces: 'shopee'
		});
		try {
			await sincronizarBusca(payload);
			await carregarBuscasSalvas();
			s.salvarAberto = false;
		} catch {
			/* best-effort */
		}
	}

	function handleCarregarSalva(config) {
		s.keywords = config.keywords ?? [];
		s.busca = config.keywords?.[0] ?? '';
		s.shopIds = config.shopIds ?? [];
		s.shopNomes = config.shopNomes ?? {};
		s.comissaoMin = config.comissaoMin || 0.07;
		s.vendasMin = config.vendasMin || 0;
		s.categorias = config.categorias ?? [];
		s.cron = config.cron ?? '';
		if (config.fontes?.length) {
			s.fontes = {
				curadoria: config.fontes.includes('curadoria'),
				quedas: config.fontes.includes('quedas'),
				novos: config.fontes.includes('novos'),
				favoritos: config.fontes.includes('favoritos'),
				lojas: config.fontes.includes('lojas')
			};
		}
	}

	async function handleRemoverSalva(config) {
		await sincronizarBusca({ keywords: config.keywords }, { remover: true });
		await carregarBuscasSalvas();
	}

	return {
		executar,
		inicializar,
		handleFontesChange,
		handleAdicionarLoja,
		handleRemoverLoja,
		handleSalvar,
		handleCarregarSalva,
		handleRemoverSalva
	};
}
