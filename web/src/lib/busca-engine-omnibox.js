/**
 * busca-engine-omnibox — lógica do Smart Search extraída da engine.
 *
 * Funções puras que processam eventos OMNIBOX_* e retornam mutações de estado.
 * A engine chama essas funções e aplica as mudanças em ui/ctx.
 *
 * Separação: a engine orquestra (estado reativo + effects); este módulo decide
 * (parsing, roteamento, keyboard, execução de intenção).
 */

import { detectarIntencao } from './omnibox-intencao.js';
import { parsearInput } from './omnibox-parser.js';
import { gerarSugestoes } from './omnibox-sugestoes.js';
import { OMNIBOX } from './busca-config.js';

/**
 * Processa OMNIBOX_INPUT: parseia tokens, roteia entre intencao e prefixo,
 * gera opcoes para o dropdown.
 *
 * @param {string} value - texto do input
 * @param {object} intencaoCtx - {categoriasDisponiveis, marketplacesFiltro, shopIds, shopNomes}
 * @param {object} sugestoesCtx - {lojasDisponiveis, categoriasDisponiveis, marketplaces, buscasSalvas}
 * @returns {{inputValue: string, aberto: boolean, highlightIdx: number, modo: string, opcoes: any[], keyword: string}}
 */
export function processarOmniboxInput(value, intencaoCtx, sugestoesCtx) {
	const tokens = parsearInput(value);
	const ultimoToken = tokens[tokens.length - 1];

	let modo, opcoes;
	if (ultimoToken && ultimoToken.tipo !== 'keyword') {
		modo = 'sugestoes';
		opcoes = gerarSugestoesPrefixo(ultimoToken, sugestoesCtx);
	} else {
		modo = 'intencao';
		const textoLivre = tokens
			.filter((t) => t.tipo === 'keyword')
			.map((t) => t.valor)
			.join(' ');
		opcoes = detectarIntencao(textoLivre, intencaoCtx);
	}

	const keyword = tokens
		.filter((t) => t.tipo === 'keyword')
		.map((t) => t.valor)
		.join(' ');

	return { inputValue: value, aberto: true, highlightIdx: -1, modo, opcoes, keyword };
}

/**
 * Processa OMNIBOX_KEYDOWN: navegação cíclica e ações.
 *
 * @param {string} key
 * @param {number|undefined} idx - para 'highlight' (mouse hover)
 * @param {number} currentHighlight
 * @param {number} opcoesLength
 * @returns {{highlightIdx?: number, aberto?: boolean, executar?: boolean}}
 */
export function processarOmniboxKeydown(key, idx, currentHighlight, opcoesLength) {
	const n = opcoesLength;

	if (key === 'highlight' && idx != null) {
		return { highlightIdx: idx };
	}
	if (key === 'ArrowDown') {
		return { highlightIdx: n ? (currentHighlight + 1) % n : -1, aberto: true };
	}
	if (key === 'ArrowUp') {
		return { highlightIdx: n ? (currentHighlight <= 0 ? n - 1 : currentHighlight - 1) : -1, aberto: true };
	}
	if (key === 'Enter') {
		return { executar: true };
	}
	if (key === 'Escape') {
		return { highlightIdx: -1, aberto: false };
	}
	return {};
}

/**
 * Resolve qual opção executar (pela highlightIdx ou primeira).
 *
 * @param {any[]} opcoes
 * @param {number} highlightIdx
 * @returns {any|null}
 */
export function resolverOpcaoAtiva(opcoes, highlightIdx) {
	if (!opcoes.length) return null;
	const idx = highlightIdx >= 0 ? highlightIdx : 0;
	return opcoes[idx] ?? null;
}

/**
 * Classifica a ação a tomar a partir de uma intenção selecionada.
 *
 * @param {object} opcao - IntencaoOption
 * @returns {{action: string, payload: object}}
 */
export function classificarIntencao(opcao) {
	switch (opcao.tipo) {
		case 'produtos':
			return { action: 'BUSCAR_PRODUTOS', payload: { keyword: opcao.payload.keyword ?? '' } };
		case 'lojas':
			return { action: 'BUSCAR_LOJAS', payload: { termo: opcao.payload.termo } };
		case 'categoria':
			return { action: 'ADICIONAR_CATEGORIA', payload: { nome: opcao.payload.categoria } };
		case 'resolver_link':
			return { action: 'RESOLVER_LINK', payload: { url: opcao.payload.url } };
		default:
			return { action: 'NOOP', payload: {} };
	}
}

/**
 * Classifica a ação a tomar a partir de uma sugestão de prefixo selecionada.
 *
 * @param {object} sug - Sugestao
 * @param {string} inputValue - valor atual do input (para remover token ativo)
 * @returns {{action: string, payload: object, novoInput: string}}
 */
export function classificarSugestaoPrefixo(sug, inputValue) {
	const novoInput = removerTokenAtivo(inputValue);

	switch (sug.tipo) {
		case 'loja':
			return { action: 'ADICIONAR_LOJA', payload: { loja: sug.meta }, novoInput };
		case 'categoria':
			return { action: 'ADICIONAR_CATEGORIA', payload: { nome: sug.label, categoria: sug.meta }, novoInput };
		case 'marketplace':
			return { action: 'MUDAR_MARKETPLACES', payload: { marketplace: sug.meta.marketplace }, novoInput };
		case 'busca_salva':
			return {
				action: 'CARREGAR_SALVA',
				payload: { config: sug.meta.config },
				novoInput: (sug.meta.config.keywords ?? [])[0] ?? ''
			};
		default:
			return { action: 'NOOP', payload: {}, novoInput };
	}
}

// ── Helpers internos ──────────────────────────────────────────────────────────

function gerarSugestoesPrefixo(ultimoToken, ctx) {
	const cfg = {
		minChars: OMNIBOX?.minChars ?? 2,
		maxSugestoes: OMNIBOX?.maxSugestoes ?? 7,
		matchBuscaSalva: OMNIBOX?.matchBuscaSalva ?? true
	};
	const sugestoesMap = gerarSugestoes(ultimoToken, ctx, cfg);
	return [...sugestoesMap.values()].flat();
}

function removerTokenAtivo(inputValue) {
	const tokens = parsearInput(inputValue);
	tokens.pop();
	return tokens
		.map((t) => {
			const pfx = t.tipo === 'keyword' ? '' : ({ loja: '@', categoria: '#', marketplace: '!' }[t.tipo] ?? '');
			return pfx + t.valor;
		})
		.join(' ');
}
