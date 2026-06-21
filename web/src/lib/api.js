// Cliente da API do Garimpo.
// Em produção o front é estático e o nginx faz proxy de /api -> Go (mesma origem),
// então a base é vazia. Em dev, aponta para o Go local. Dá pra sobrescrever com
// VITE_API_BASE se precisar (ex.: front e API em hosts diferentes).
const BASE = import.meta.env.VITE_API_BASE ?? (import.meta.env.PROD ? '' : 'http://localhost:8080');

async function pegar(caminho) {
	const resp = await fetch(`${BASE}${caminho}`);
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
export function registrarSelecao(candidato) {
	return fetch(`${BASE}/api/eventos`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ tipo: 'selecao', ...candidato })
	}).catch(() => {
		/* telemetria não pode atrapalhar o uso */
	});
}

async function postar(caminho, corpo) {
	const resp = await fetch(`${BASE}${caminho}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
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
