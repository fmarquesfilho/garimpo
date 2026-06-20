// Cliente da API do Garimpo. Base configurável via VITE_API_BASE.
const BASE = import.meta.env.VITE_API_BASE ?? 'http://localhost:8080';

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
	notaMin
} = {}) {
	const p = new URLSearchParams({ estrategia, top: String(top) });
	if (keyword) p.set('keyword', keyword);
	if (categoria) p.set('categoria', categoria);
	if (cat != null) p.set('cat', String(cat));
	if (comissaoMin != null) p.set('comissao_min', String(comissaoMin));
	if (vendasMin != null) p.set('vendas_min', String(vendasMin));
	if (notaMin != null) p.set('nota_min', String(notaMin));
	return pegar(`/api/candidatos?${p}`);
}

/** Os dois rankings lado a lado. */
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
