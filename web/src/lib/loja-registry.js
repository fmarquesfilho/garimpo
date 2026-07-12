import { LOJA_REGISTRO } from './busca-config.js';

/** Mínimo de caracteres para matching — de rules.lojaRegistro (fonte da verdade). */
const MATCH_MIN_CHARS = LOJA_REGISTRO?.matchMinChars ?? 2;

/**
 * Normaliza nome para matching: NFD → remove diacríticos → remove non-[a-z0-9] → lowercase.
 * Idêntica à implementação C# `Loja.Normalizar()`.
 * @param {string} nome
 * @returns {string} nome normalizado (apenas [a-z0-9])
 */
export function normalizarNome(nome) {
	if (!nome) return '';
	return nome
		.normalize('NFD')
		.replace(/[\u0300-\u036f]/g, '') // remove combining marks
		.replace(/[^a-z0-9]/gi, '') // keep only alphanum
		.toLowerCase();
}

/**
 * Faz matching de um input (já normalizado) contra a lista de lojas do registro.
 * Retorna lojas que casam por substring no NomeNormalizado OU no Nome canônico (lowercase).
 *
 * @param {string} inputNormalizado — input do usuário já normalizado
 * @param {Array<{id:string, nome:string, nome_normalizado:string, marketplace:string}>} lojas
 * @param {number} [max=7]
 * @returns {Array} lojas que casam
 */
export function matchLojas(inputNormalizado, lojas, max = 7) {
	if (!inputNormalizado || inputNormalizado.length < MATCH_MIN_CHARS) return [];
	return (lojas ?? [])
		.filter(
			(l) =>
				(l.nome_normalizado && l.nome_normalizado.includes(inputNormalizado)) ||
				(l.nome && l.nome.toLowerCase().includes(inputNormalizado))
		)
		.slice(0, max);
}
