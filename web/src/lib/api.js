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
		let detalhe = '';
		try {
			const corpo = await resp.json();
			detalhe = corpo?.erro ?? '';
		} catch {
			/* corpo não-JSON */
		}
		throw new Error(detalhe || `Falha ${resp.status}`);
	}
	return resp.json();
}

/** Lista priorizada de uma estratégia. */
export function buscarCandidatos({
	estrategia = 'nicho',
	top = 10,
	keyword,
	categoria,
	cat,
	comissaoMin,
	vendasMin,
	notaMin,
	exploracao
} = {}) {
	const p = new URLSearchParams({ estrategia, top: String(top) });
	if (keyword) p.set('keyword', keyword);
	if (categoria) p.set('categoria', categoria);
	if (cat != null) p.set('cat', String(cat));
	if (comissaoMin != null) p.set('comissao_min', String(comissaoMin));
	if (vendasMin != null) p.set('vendas_min', String(vendasMin));
	if (notaMin != null) p.set('nota_min', String(notaMin));
	if (exploracao != null) p.set('exploracao', String(exploracao));
	return pegar(`/api/candidatos?${p}`);
}

/** Registra uma decisão de curadoria (seleção) para análise. Best-effort. */
export async function registrarSelecao(candidato) {
	const headers = { 'Content-Type': 'application/json', ...(await authHeaders()) };
	return fetch(`${BASE}/api/eventos`, {
		method: 'POST',
		headers,
		body: JSON.stringify({ tipo: 'selecao', ...candidato })
	}).catch(() => {
		/* telemetria não pode atrapalhar o uso */
	});
}

async function postar(caminho, corpo) {
	const headers = { 'Content-Type': 'application/json', ...(await authHeaders()) };
	const resp = await fetch(`${BASE}${caminho}`, {
		method: 'POST',
		headers,
		body: JSON.stringify(corpo)
	});
	if (!resp.ok) {
		let detalhe = '';
		try {
			detalhe = (await resp.json())?.erro ?? '';
		} catch {
			/* corpo não-JSON */
		}
		throw new Error(detalhe || `Falha ${resp.status}`);
	}
	return resp.json();
}

/** Publica a oferta no canal (Telegram/Mock) e devolve o Resultado. */
export function publicar(candidato) {
	return postar('/api/publicar', candidato);
}

/** Resumo descritivo dos snapshots coletados (por categoria), janela em dias. */
export function buscarEstatisticas({ dias = 30 } = {}) {
	return pegar(`/api/estatisticas?dias=${dias}`);
}

/** Lista os perfis de busca sincronizados no servidor (BigQuery). */
export function listarBuscasServidor() {
	return pegar('/api/buscas');
}

/** Salva (sync) um perfil de busca no servidor. Best-effort. */
export async function sincronizarBusca(busca, { remover = false } = {}) {
	const qs = remover ? '?remover' : '';
	const corpo = remover
		? { id: busca.id, keywords: busca.keywords ?? [] }
		: busca;
	const headers = { 'Content-Type': 'application/json', ...(await authHeaders()) };
	return fetch(`${BASE}/api/buscas${qs}`, {
		method: 'POST',
		headers,
		body: JSON.stringify(corpo)
	}).catch(() => {
		/* sync não pode travar o uso local */
	});
}

export function compararEstrategias({
	top = 8,
	keyword,
	categoria,
	cat,
	comissaoMin,
	vendasMin,
	notaMin
} = {}) {
	const p = new URLSearchParams({ top: String(top) });
	if (keyword) p.set('keyword', keyword);
	if (categoria) p.set('categoria', categoria);
	if (cat != null) p.set('cat', String(cat));
	if (comissaoMin != null) p.set('comissao_min', String(comissaoMin));
	if (vendasMin != null) p.set('vendas_min', String(vendasMin));
	if (notaMin != null) p.set('nota_min', String(notaMin));
	return pegar(`/api/comparar?${p}`);
}
