/**
 * omnibox-intencao — detecção de intenção a partir de texto livre (sem prefixo).
 *
 * Função pura que analisa o texto digitado no Omnibox e gera opções de intenção
 * para o Smart Dropdown. Não depende de DOM, runes ou estado global.
 *
 * Regras:
 * - URL detectada → opção "Resolver Link" (exclusiva, sem outras opções)
 * - Texto >= minChars → opções base ("Pesquisar em Produtos" + "Pesquisar em Lojas")
 * - Match de categoria de 1º nível → opções adicionais por marketplace
 * - Texto < minChars → sem opções (dropdown fecha)
 *
 * Config-driven: lê de rules.omnibox.intencao para minChars, ordemOpcoes,
 * maxCategorias, urlPatterns.
 */

import { OMNIBOX, MARKETPLACES } from './busca-config.js';

// ── Config do JSON (com fallbacks seguros) ────────────────────────────────

const INTENCAO_CONFIG = OMNIBOX?.intencao ?? {
	habilitado: true,
	minChars: 2,
	ordemOpcoes: ['produtos', 'lojas', 'categorias'],
	maxCategorias: 3,
	urlPatterns: ['^https?://'],
	enterSemSelecao: 'primeira_opcao',
	navegacaoCiclica: true
};

const MIN_CHARS = INTENCAO_CONFIG.minChars ?? 2;
const MAX_CATEGORIAS = INTENCAO_CONFIG.maxCategorias ?? 3;
const URL_PATTERNS = (INTENCAO_CONFIG.urlPatterns ?? ['^https?://']).map((p) => new RegExp(p, 'i')); // nosemgrep: detect-non-literal-regexp
const ORDEM = INTENCAO_CONFIG.ordemOpcoes ?? ['produtos', 'lojas', 'categorias'];

// ── Ícones por tipo de intenção ──────────────────────────────────────────

const ICONES = {
	produtos: '🔎',
	lojas: '🏪',
	categoria: '📂',
	resolver_link: '🔗'
};

// ── API pública ───────────────────────────────────────────────────────────

/**
 * @typedef {Object} IntencaoOption
 * @property {string} tipo
 * @property {string} label
 * @property {string} labelAcessivel
 * @property {string} icone
 * @property {Object} payload
 */

/**
 * @typedef {Object} IntencaoCtx
 * @property {Array<{nome:string, marketplaces?:string[]}>} [categoriasDisponiveis]
 * @property {string[]} [marketplacesFiltro] - marketplaces ativos no filtro (vazio = todos)
 * @property {number[]} [shopIds] - lojas ativas no escopo
 * @property {Record<number, string>} [shopNomes] - id -> nome da loja
 */

/**
 * Detecta a intenção do usuário a partir do texto livre e contexto.
 * Retorna lista ordenada de opções para o Smart Dropdown.
 *
 * @param {string} texto - texto completo do Omnibox (sem prefixo no ultimo token)
 * @param {IntencaoCtx} [ctx] - contexto da engine
 * @returns {IntencaoOption[]}
 */
export function detectarIntencao(texto, ctx = {}) {
	if (!INTENCAO_CONFIG.habilitado) return [];
	const trimmed = (texto ?? '').trim();

	// URL detectada → opção exclusiva "Resolver Link"
	if (isUrl(trimmed)) {
		return [
			{
				tipo: 'resolver_link',
				label: 'Resolver Link',
				labelAcessivel: `Resolver link ${trimmed}`,
				icone: ICONES.resolver_link,
				payload: { url: trimmed }
			}
		];
	}

	// Texto curto demais → sem opções
	if (trimmed.length < MIN_CHARS) return [];

	/** @type {IntencaoOption[]} */
	const opcoes = [];

	// Gera opções na ORDEM configurada
	for (const tipo of ORDEM) {
		if (tipo === 'produtos') {
			opcoes.push(buildProdutosOption(trimmed, ctx));
		} else if (tipo === 'lojas') {
			opcoes.push(buildLojasOption(trimmed));
		} else if (tipo === 'categorias') {
			opcoes.push(...buildCategoriasOptions(trimmed, ctx));
		}
	}

	return opcoes;
}

/**
 * Verifica se o texto é uma URL.
 * @param {string} texto
 * @returns {boolean}
 */
export function isUrl(texto) {
	if (!texto) return false;
	return URL_PATTERNS.some((re) => re.test(texto));
}

// ── Builders internos ─────────────────────────────────────────────────────

/** @returns {IntencaoOption} */
function buildProdutosOption(texto, ctx) {
	const sufixo = buildSufixoContexto(ctx);
	return {
		tipo: 'produtos',
		label: `Pesquisar "${texto}" em Produtos${sufixo}`,
		labelAcessivel: `Pesquisar ${texto} em Produtos${sufixo}`,
		icone: ICONES.produtos,
		payload: { keyword: texto }
	};
}

/** @returns {IntencaoOption} */
function buildLojasOption(texto) {
	return {
		tipo: 'lojas',
		label: `Pesquisar "${texto}" em Lojas`,
		labelAcessivel: `Pesquisar ${texto} em Lojas`,
		icone: ICONES.lojas,
		payload: { termo: texto }
	};
}

/** @returns {IntencaoOption[]} */
function buildCategoriasOptions(texto, ctx) {
	const categorias = ctx?.categoriasDisponiveis ?? [];
	const marketplacesAtivos = ctx?.marketplacesFiltro?.length
		? ctx.marketplacesFiltro
		: (MARKETPLACES?.suportados ?? []);

	const q = texto.toLowerCase();
	const matches = categorias
		.filter((c) => {
			const nome = (c.nome ?? c).toLowerCase();
			return nome.includes(q);
		})
		.slice(0, MAX_CATEGORIAS);

	return matches.map((cat) => buildCategoriaOption(cat, marketplacesAtivos, ctx)).filter(Boolean);
}

/** @returns {IntencaoOption|null} */
function buildCategoriaOption(cat, marketplacesAtivos, ctx) {
	const nome = cat.nome ?? cat;
	const catMarketplaces = cat.marketplaces ?? marketplacesAtivos;
	const mktsRelevantes = catMarketplaces.filter((m) => marketplacesAtivos.includes(m));

	if (mktsRelevantes.length === 0) return null;

	const shopIds = ctx?.shopIds ?? [];
	const shopNomes = ctx?.shopNomes ?? {};
	const label = buildCategoriaLabel(nome, shopIds, shopNomes, mktsRelevantes);
	const labelAcessivel = buildCategoriaLabelAcessivel(nome, shopIds, shopNomes);

	return {
		tipo: 'categoria',
		label,
		labelAcessivel,
		icone: ICONES.categoria,
		payload: { categoria: nome, marketplaces: mktsRelevantes }
	};
}

function buildCategoriaLabel(nome, shopIds, shopNomes, mktsRelevantes) {
	if (shopIds.length === 1) return `Pesquisar em #${nome} na ${shopNomes[shopIds[0]] || String(shopIds[0])}`;
	if (shopIds.length > 1) return `Pesquisar em #${nome} nas lojas selecionadas`;
	if (mktsRelevantes.length === 1) return `Pesquisar em #${nome} na ${mktsRelevantes[0]}`;
	return `Pesquisar em #${nome}`;
}

function buildCategoriaLabelAcessivel(nome, shopIds, shopNomes) {
	if (shopIds.length === 1)
		return `Pesquisar por categoria ${nome} na loja ${shopNomes[shopIds[0]] || String(shopIds[0])}`;
	if (shopIds.length > 1) return `Pesquisar por categoria ${nome} nas lojas selecionadas`;
	return `Pesquisar por categoria ${nome} em todos os marketplaces`;
}

/**
 * Constrói sufixo de contexto para a opção de Produtos.
 * Ex: " na shopee", " na Glory of Seoul"
 */
function buildSufixoContexto(ctx) {
	const shopIds = ctx?.shopIds ?? [];
	const shopNomes = ctx?.shopNomes ?? {};
	const mkts = ctx?.marketplacesFiltro ?? [];

	if (shopIds.length === 1) {
		const nome = shopNomes[shopIds[0]] || String(shopIds[0]);
		return ` na ${nome}`;
	}
	if (shopIds.length > 1) return ' nas lojas selecionadas';
	if (mkts.length === 1) return ` na ${mkts[0]}`;
	return '';
}
