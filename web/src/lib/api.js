// Cliente da API do Garimpo.
// Em produção o front é estático e o nginx faz proxy de /api -> Go (mesma origem),
// então a base é vazia. Em dev, aponta para o Go local. Dá pra sobrescrever com
// VITE_API_BASE se precisar (ex.: front e API em hosts diferentes).
import { getIdToken } from './firebase.js';

const BASE = import.meta.env.VITE_API_BASE ?? (import.meta.env.PROD ? '' : 'http://localhost:8080');

/** Headers com token de auth (se logado). */
async function authHeaders() {
	const token = await getIdToken();
	if (token) return { Authorization: `Bearer ${token}` };
	return {};
}

async function pegar(caminho) {
	const headers = await authHeaders();
	const resp = await fetch(`${BASE}${caminho}`, { headers });
	if (!resp.ok) {
		throw await parseProblem(resp, caminho);
	}
	return resp.json();
}

/**
 * Lista priorizada de uma estratégia.
 * @param {Object} opts
 * @param {string} [opts.estrategia]
 * @param {number} [opts.top]
 * @param {string} [opts.keyword]
 * @param {string} [opts.categoria]
 * @param {string} [opts.cat]
 * @param {number} [opts.comissaoMin]
 * @param {number} [opts.vendasMin]
 * @param {number} [opts.notaMin]
 * @param {boolean} [opts.exploracao]
 * @param {string} [opts.fonte]
 * @param {string|string[]} [opts.shopIds]
 * @param {boolean} [opts.semFiltro]
 */
export function buscarCandidatos({
	estrategia = 'nicho',
	top = 10,
	keyword,
	categoria,
	cat,
	comissaoMin,
	vendasMin,
	notaMin,
	exploracao,
	fonte,
	shopIds,
	semFiltro
} = {}) {
	const p = new URLSearchParams({ estrategia, top: String(top) });
	if (keyword) p.set('keyword', keyword);
	if (categoria) p.set('categoria', categoria);
	if (cat != null) p.set('cat', String(cat));
	if (comissaoMin != null) p.set('comissao_min', String(comissaoMin));
	if (vendasMin != null) p.set('vendas_min', String(vendasMin));
	if (notaMin != null) p.set('nota_min', String(notaMin));
	if (exploracao != null) p.set('exploracao', String(exploracao));
	if (fonte) p.set('fonte', fonte);
	if (shopIds) p.set('shop_ids', Array.isArray(shopIds) ? shopIds.join(',') : String(shopIds));
	if (semFiltro) p.set('sem_filtro', 'true');
	return pegar(`/api/candidatos?${p}`);
}

async function postar(caminho, corpo) {
	const headers = { 'Content-Type': 'application/json', ...(await authHeaders()) };
	const resp = await fetch(`${BASE}${caminho}`, {
		method: 'POST',
		headers,
		body: JSON.stringify(corpo)
	});
	if (!resp.ok) {
		throw await parseProblem(resp, caminho);
	}
	return resp.json();
}

/**
 * Parseia uma resposta de erro no formato RFC 9457 (Problem Details).
 * Retorna um Error enriquecido com campos úteis para o frontend.
 * @returns {Promise<Error & {status: number, problem: object, retry: boolean, code: string, endpoint: string}>}
 */
async function parseProblem(resp, caminho) {
	let problem = {};
	try {
		problem = await resp.json();
	} catch {
		/* corpo não-JSON */
	}

	// Mensagem amigável: prefere detail > erro > title > status text
	const mensagem = problem.detail || problem.erro || problem.title || `Erro ${resp.status}`;

	/** @type {any} */
	const err = new Error(mensagem);
	err.status = resp.status;
	err.problem = problem; // RFC 9457 completo
	err.retry = problem.retry ?? false;
	err.code = problem.code ?? '';
	err.endpoint = caminho;
	return err;
}

/**
 * Publica a oferta no canal (Telegram/WhatsApp/Mock) e devolve o Resultado.
 * @param {Object} candidato
 * @param {Object} [opts]
 * @param {string} [opts.destinoId]
 * @param {string} [opts.templateId]
 */
export function publicar(candidato, { destinoId, templateId } = {}) {
	const corpo = { ...candidato };
	if (destinoId) corpo.destino_id = destinoId;
	if (templateId) corpo.template_id = templateId;
	return postar('/api/publicar', corpo);
}

/** Lista os destinos de publicação cadastrados (Telegram, WhatsApp, etc.). */
export function listarDestinos() {
	return pegar('/api/destinos');
}

/** Salva (cria/atualiza) um destino de publicação. */
export function salvarDestino(destino) {
	return postar('/api/destinos', destino);
}

/** Remove um destino por ID. */
export async function deletarDestino(id) {
	const headers = { ...(await authHeaders()) };
	const resp = await fetch(`${BASE}/api/destinos?id=${encodeURIComponent(id)}`, {
		method: 'DELETE',
		headers
	});
	if (!resp.ok) {
		throw await parseProblem(resp, '/api/destinos');
	}
	return resp.json();
}

/** Lista os templates de mensagem disponíveis. */
export function listarTemplates() {
	return pegar('/api/templates');
}

/** Salva (cria/atualiza) um template de mensagem. */
export function salvarTemplate(template) {
	return postar('/api/templates', template);
}

/** Remove um template por ID. */
export async function deletarTemplate(id) {
	const headers = { ...(await authHeaders()) };
	const resp = await fetch(`${BASE}/api/templates?id=${encodeURIComponent(id)}`, {
		method: 'DELETE',
		headers
	});
	if (!resp.ok) {
		throw await parseProblem(resp, '/api/templates');
	}
	return resp.json();
}

/** Renderiza um preview de template com dados do produto. */
export function previewTemplate(dados) {
	return postar('/api/templates/preview', dados);
}

/** Relatório de conversões (publicações por canal/destino). */
export function buscarConversoes({ dias = 30 } = {}) {
	return pegar(`/api/conversoes?dias=${dias}`);
}

/** Logs recentes para o dashboard de admin. */
/** Verifica se o usuário logado é admin. */
export function verificarAdmin() {
	return pegar('/api/admin/me');
}

/** Lista publicações por status (agendada|enviada|erro; vazio = todas). */
export function listarPublicacoes({ status = '' } = {}) {
	const qs = status ? `?status=${status}` : '';
	return pegar(`/api/publicacoes${qs}`);
}

/** Agenda ou envia imediatamente uma publicação. */
export function agendarPublicacao(pub) {
	return postar('/api/publicacoes', pub);
}

/** Resumo descritivo dos snapshots coletados (por categoria), janela em dias. */
export function buscarEstatisticas({ dias = 30 } = {}) {
	return pegar(`/api/estatisticas?dias=${dias}`);
}

/** Novidades de lojas monitoradas (produtos novos + variações de preço). */
export function buscarNovidades({ buscaId = '', dias = 7 } = {}) {
	const p = new URLSearchParams({ dias: String(dias) });
	if (buscaId) p.set('busca_id', buscaId);
	return pegar(`/api/lojas/novidades?${p}`);
}

/** Evolução de preço das lojas monitoradas ao longo do tempo. */
export function buscarEvolucaoLojas({ dias = 30 } = {}) {
	return pegar(`/api/lojas/evolucao?dias=${dias}`);
}

/** Configuração atual dos alertas de preço. */
export function buscarAlertasConfig() {
	return pegar('/api/alertas');
}

/**
 * Envia um alerta de teste (verifica bot + chat_id).
 * @param {Object} [opts]
 * @param {string} [opts.buscaId]
 */
export function testarAlertas({ buscaId } = {}) {
	const corpo = buscaId ? { busca_id: buscaId } : {};
	return postar('/api/alertas/testar', corpo);
}

/**
 * Atualiza configuração de alertas em runtime.
 * @param {Object} opts
 * @param {string} [opts.chatId]
 * @param {number} [opts.threshold]
 * @param {boolean} [opts.apenasQuedas]
 */
export function configurarAlertas({ chatId, threshold, apenasQuedas } = {}) {
	const corpo = {};
	if (chatId != null) corpo.chat_id = chatId;
	if (threshold != null) corpo.threshold = threshold;
	if (apenasQuedas != null) corpo.apenas_quedas = apenasQuedas;
	return postar('/api/alertas/configurar', corpo);
}

/**
 * Adiciona uma loja ao monitoramento (aceita URL ou ID numérico).
 * @param {Object} opts
 * @param {string} opts.input
 * @param {string} [opts.cron]
 * @param {string} [opts.origemPadrao]
 */
export function adicionarLoja({ input, cron, origemPadrao }) {
	const corpo = { input };
	if (cron) corpo.cron = cron;
	if (origemPadrao) corpo.origem_padrao = origemPadrao;
	return postar('/api/lojas', corpo);
}

/** Remove uma loja do monitoramento. */
export async function removerLoja(id) {
	const headers = { ...(await authHeaders()) };
	const resp = await fetch(`${BASE}/api/lojas?id=${encodeURIComponent(id)}`, {
		method: 'DELETE',
		headers
	});
	if (!resp.ok) {
		throw await parseProblem(resp, '/api/lojas');
	}
	return resp.json();
}

/** Resolve um link curto da Shopee para obter URL final + dados do produto. */
export function resolverLinkShopee(url) {
	return postar('/api/resolver-link', { url });
}

/** Histórico de coletas executadas (snapshots por execução), janela em dias. */
export function buscarColetas({ dias = 30 } = {}) {
	return pegar(`/api/coletas?dias=${dias}`);
}

// ── Onboarding / Tenant ──────────────────────────────────────────────────

/** Status do onboarding do tenant atual. */
export function onboardingStatus() {
	return pegar('/api/onboarding/status');
}

/** Step 1: Aceitar termos de uso. */
export function onboardingTermos() {
	return postar('/api/onboarding/termos', { aceito: true });
}

/** Step 2: Salvar credenciais Shopee. */
export function onboardingShopee({ appId, secret }) {
	return postar('/api/onboarding/shopee', { app_id: appId, secret });
}

/**
 * Step 3: Configurar Telegram (ou pular).
 * @param {Object} opts
 * @param {string} [opts.token]
 * @param {string} [opts.chatId]
 * @param {boolean} [opts.pular]
 */
export function onboardingTelegram({ token, chatId, pular = false } = {}) {
	if (pular) return postar('/api/onboarding/telegram', { pular: true });
	return postar('/api/onboarding/telegram', { token, chat_id: chatId });
}

/**
 * Step 3 (alternativo): Configurar WhatsApp Meta (ou pular).
 * @param {Object} opts
 * @param {string} [opts.phoneNumberId]
 * @param {string} [opts.accessToken]
 * @param {boolean} [opts.pular]
 */
export function onboardingWhatsapp({ phoneNumberId, accessToken, pular = false } = {}) {
	if (pular) return postar('/api/onboarding/whatsapp', { pular: true });
	return postar('/api/onboarding/whatsapp', { phone_number_id: phoneNumberId, access_token: accessToken });
}

/** Step 4: Validar credenciais Shopee com chamada de teste. */
export function onboardingValidar() {
	return postar('/api/onboarding/validar', {});
}

/** Excluir conta e dados (LGPD). */
export function excluirConta() {
	return postar('/api/onboarding/excluir-conta', { confirmar: true });
}

/** Conversões reais da Shopee (conversionReport). */
export async function buscarConversoesReais({ dias = 30 } = {}) {
	const controller = new AbortController();
	const timeout = setTimeout(() => controller.abort(), 15000);
	try {
		const headers = await authHeaders();
		const resp = await fetch(`${BASE}/api/conversoes/reais?dias=${dias}`, {
			headers,
			signal: controller.signal
		});
		clearTimeout(timeout);
		if (!resp.ok) throw await parseProblem(resp, '/api/conversoes/reais');
		return resp.json();
	} catch (err) {
		clearTimeout(timeout);
		if (err.name === 'AbortError') throw new Error('A Shopee não respondeu a tempo. Tente novamente.', { cause: err });
		throw err;
	}
}

/** Lista os perfis de busca sincronizados no servidor (BigQuery). */
export function listarBuscasServidor() {
	return pegar('/api/buscas');
}

/** Salva (sync) um perfil de busca no servidor. Best-effort. */
export async function sincronizarBusca(busca, { remover = false } = {}) {
	const qs = remover ? '?remover' : '';
	const corpo = remover ? { id: busca.id, keywords: busca.keywords ?? [] } : busca;
	const headers = { 'Content-Type': 'application/json', ...(await authHeaders()) };
	return fetch(`${BASE}/api/buscas${qs}`, {
		method: 'POST',
		headers,
		body: JSON.stringify(corpo)
	}).catch(() => {
		/* sync não pode travar o uso local */
	});
}

// ── Favoritos ─────────────────────────────────────────────────────────────
/** Lista os favoritos do usuário logado. */
export function listarFavoritos() {
	return pegar('/api/favoritos');
}

/** Salva um produto como favorito. */
export function salvarFavorito(produto) {
	return postar('/api/favoritos', produto);
}

/** Remove um produto dos favoritos. */
export async function removerFavorito(produtoId) {
	const headers = { ...(await authHeaders()) };
	const resp = await fetch(`${BASE}/api/favoritos?produto_id=${encodeURIComponent(produtoId)}`, {
		method: 'DELETE',
		headers
	});
	if (!resp.ok) {
		throw await parseProblem(resp, '/api/favoritos');
	}
	return resp.json();
}
