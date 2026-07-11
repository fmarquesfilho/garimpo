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

	for (const id of shopIds ?? []) {
		const s = String(id);
		if (!seen.has(s)) {
			seen.add(s);
			keys.push(s);
		}
	}

	for (const kw of keywords ?? []) {
		const normalized = kw.trim().toLowerCase();
		if (normalized && !seen.has(normalized)) {
			seen.add(normalized);
			keys.push(normalized);
		}
	}

	// Categorias are used as collection keys ONLY when shop_ids and keywords are both empty
	if (!(shopIds?.length) && !(keywords?.length)) {
		for (const cat of categorias ?? []) {
			const normalized = cat.trim().toLowerCase();
			if (normalized && !seen.has(normalized)) {
				seen.add(normalized);
				keys.push(normalized);
			}
		}
	}

	keys.sort();
	return keys;
}
