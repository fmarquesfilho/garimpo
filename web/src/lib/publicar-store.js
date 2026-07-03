/**
 * Store temporário para passar dados do produto à página /publicar.
 * Usa sessionStorage para sobreviver à navegação client-side sem poluir a URL.
 */
import { browser } from '$app/environment';

const STORAGE_KEY = 'garimpei:publicar:produto';

/** Salva o produto no sessionStorage e retorna a URL de navegação. */
export function prepararPublicacao(produto) {
	if (browser) {
		sessionStorage.setItem(STORAGE_KEY, JSON.stringify(produto));
	}
	return '/publicar';
}

/** Recupera e limpa o produto do sessionStorage. */
export function recuperarProduto() {
	if (!browser) return null;
	const raw = sessionStorage.getItem(STORAGE_KEY);
	if (!raw) return null;
	sessionStorage.removeItem(STORAGE_KEY);
	try {
		return JSON.parse(raw);
	} catch {
		return null;
	}
}
