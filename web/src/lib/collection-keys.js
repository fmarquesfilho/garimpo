/**
 * Derivação determinística de collection_keys a partir de uma Busca.
 * Lógica idêntica em Go, Python, C# e TypeScript.
 *
 * Regras:
 *   - Cada shop_id vira sua representação string
 *   - Cada keyword é trimmed e lowercased
 *   - Cada categoria é trimmed e lowercased (fallback para tipo=categoria)
 *   - Strings vazias após normalização são descartadas
 *   - Resultado é sorted lexicograficamente e sem duplicatas
 *
 * @param {number[]} shopIds
 * @param {string[]} keywords
 * @param {string[]} [categorias]
 * @returns {string[]}
 */
export function deriveCollectionKeys(shopIds, keywords, categorias) {
	const seen = new Set();
	const keys = [];

	addShopIds(shopIds, seen, keys);
	addNormalized(keywords, seen, keys);

	// Categorias only become keys when shop_ids and keywords are both empty
	if (!shopIds?.length && !keywords?.length) {
		addNormalized(categorias, seen, keys);
	}

	keys.sort();
	return keys;
}

/** @param {number[]|null|undefined} ids */
function addShopIds(ids, seen, keys) {
	for (const id of ids ?? []) {
		const s = String(id);
		if (!seen.has(s)) {
			seen.add(s);
			keys.push(s);
		}
	}
}

/** @param {string[]|null|undefined} items */
function addNormalized(items, seen, keys) {
	for (const item of items ?? []) {
		const normalized = item.trim().toLowerCase();
		if (normalized && !seen.has(normalized)) {
			seen.add(normalized);
			keys.push(normalized);
		}
	}
}
