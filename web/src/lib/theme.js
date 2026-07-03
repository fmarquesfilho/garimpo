/**
 * Theme Engine — gerencia detecção, persistência e aplicação do tema.
 *
 * Funciona como módulo puro JS (para blocking script) e como Svelte store
 * (para reatividade em componentes).
 *
 * @typedef {'light' | 'dark' | 'system'} ThemePreference
 * @typedef {'light' | 'dark'} ResolvedTheme
 */

import { writable } from 'svelte/store';

const browser = typeof window !== 'undefined';

const STORAGE_KEY = 'theme';
const VALID = ['light', 'dark', 'system'];

/**
 * Lê a preferência salva no localStorage.
 * @returns {ThemePreference | null}
 */
export function getStoredTheme() {
	if (!browser) return null;
	try {
		const stored = localStorage.getItem(STORAGE_KEY);
		return VALID.includes(stored) ? /** @type {ThemePreference} */ (stored) : null;
	} catch {
		return null;
	}
}

/**
 * Detecta a preferência do sistema operacional.
 * @returns {ResolvedTheme}
 */
export function getSystemTheme() {
	if (!browser) return 'light';
	return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

/**
 * Resolve o tema efetivo (override manual ou sistema).
 * @returns {ResolvedTheme}
 */
export function resolveTheme() {
	const stored = getStoredTheme();
	if (stored === 'light' || stored === 'dark') return stored;
	return getSystemTheme();
}

/**
 * Aplica o tema ao documento root.
 * @param {ResolvedTheme} theme
 */
export function applyTheme(theme) {
	if (!browser) return;
	document.documentElement.setAttribute('data-theme', theme);
}

/**
 * Define a preferência do usuário, persiste e aplica.
 * @param {ThemePreference} preference
 */
export function setTheme(preference) {
	if (!browser) return;
	try {
		if (preference === 'system') {
			localStorage.removeItem(STORAGE_KEY);
		} else {
			localStorage.setItem(STORAGE_KEY, preference);
		}
	} catch {
		// localStorage indisponível — aplica sem persistir
	}
	applyTheme(preference === 'system' ? getSystemTheme() : preference);
	_store.set(preference);
}

/**
 * Escuta mudanças na preferência do sistema.
 * @param {(theme: ResolvedTheme) => void} callback
 * @returns {() => void} unsubscribe
 */
export function onSystemChange(callback) {
	if (!browser) return () => {};
	const mq = window.matchMedia('(prefers-color-scheme: dark)');
	const handler = () => {
		const stored = getStoredTheme();
		if (!stored || stored === 'system') {
			const resolved = mq.matches ? 'dark' : 'light';
			applyTheme(resolved);
			callback(resolved);
		}
	};
	mq.addEventListener('change', handler);
	return () => mq.removeEventListener('change', handler);
}

// ── Svelte Store ──────────────────────────────────────────────────────────────

const _store = writable(/** @type {ThemePreference} */ ('system'));

/** Store reativo do tema. set() aplica e persiste. */
export const theme = {
	subscribe: _store.subscribe,
	/** @param {ThemePreference} value */
	set: setTheme
};

/**
 * Inicializa o theme engine (chama uma vez no layout).
 * Sincroniza o store com o estado atual e escuta mudanças do sistema.
 * @returns {() => void} cleanup
 */
export function initTheme() {
	const stored = getStoredTheme();
	const preference = stored ?? 'system';
	_store.set(preference);
	applyTheme(resolveTheme());

	return onSystemChange((resolved) => {
		_store.set('system');
		applyTheme(resolved);
	});
}
