/**
 * API client — Dashboard v2 + Realtime endpoints.
 * Separado de api.js para manter o arquivo principal dentro do limite de 400 linhas.
 */
import { getIdToken } from './firebase.js';

const BASE = import.meta.env.VITE_API_BASE ?? (import.meta.env.PROD ? '' : 'http://localhost:8080');

async function authHeaders() {
	const token = await getIdToken();
	if (token) return { Authorization: `Bearer ${token}` };
	return {};
}

async function pegar(caminho) {
	const headers = await authHeaders();
	const resp = await fetch(`${BASE}${caminho}`, { headers });
	if (!resp.ok) {
		/** @type {any} */
		const err = new Error(`Erro ${resp.status}`);
		err.status = resp.status;
		throw err;
	}
	return resp.json();
}

/** Saúde das coletas (última execução, atrasos, keywords paradas). */
export function buscarSaudeColetas() {
	return pegar('/api/coletas/saude');
}

/** Oportunidades agora: top quedas + novos + alto-valor não publicados. */
export function buscarOportunidadesAgora({ dias = 7 } = {}) {
	return pegar(`/api/oportunidades/agora?dias=${dias}`);
}

/** Resumo de conversões/receita por canal. */
export function buscarResumoConversoes({ dias = 30 } = {}) {
	return pegar(`/api/conversoes/resumo?dias=${dias}`);
}

/** Eficácia dos alertas: quedas → alertas → conversões. */
export function buscarEficaciaAlertas({ dias = 30 } = {}) {
	return pegar(`/api/alertas/eficacia?dias=${dias}`);
}

/** Change detection para smart polling do dashboard. */
export function buscarDashboardChanges() {
	return pegar('/api/dashboard/changes');
}
