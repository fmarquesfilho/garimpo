/**
 * Módulo de categorias — busca e cache das categorias de marketplace.
 *
 * As categorias vêm da API (/api/categorias) que retorna a lista oficial
 * por marketplace. O frontend faz cache local e usa para autocomplete.
 *
 * Arquitetura multi-marketplace:
 * - Cada marketplace tem sua taxonomia de categorias
 * - O frontend mostra todas as categorias disponíveis (unificadas)
 * - O filtro client-side usa match parcial case-insensitive
 * - Futuramente: mapeamento cross-marketplace (ex: "Beleza" Shopee ≈ "Beauty" Amazon)
 */

let cache = null;
let cacheTimestamp = 0;
const CACHE_TTL = 1000 * 60 * 60; // 1 hora

/**
 * Busca categorias da API com cache de 1h.
 * Retorna array de { id, nome, slug, marketplace }.
 */
export async function buscarCategorias() {
	if (cache && Date.now() - cacheTimestamp < CACHE_TTL) {
		return cache;
	}

	try {
		const resp = await fetch('/api/categorias');
		if (!resp.ok) return cache ?? getFallback();
		const data = await resp.json();

		const todas = [];
		for (const mp of data.marketplaces ?? []) {
			for (const cat of mp.categorias ?? []) {
				todas.push({ ...cat, marketplace: mp.marketplace });
			}
		}

		cache = todas;
		cacheTimestamp = Date.now();
		return todas;
	} catch {
		return cache ?? getFallback();
	}
}

/**
 * Filtra categorias por termo (para autocomplete).
 * Match parcial case-insensitive no nome.
 */
export function filtrarCategorias(categorias, termo) {
	if (!termo?.trim()) return categorias;
	const t = termo.trim().toLowerCase();
	return categorias.filter(c => c.nome.toLowerCase().includes(t));
}

/**
 * Fallback hardcoded — usado se a API não responder.
 * Baseado nas categorias nível 1 da Shopee Brasil.
 */
function getFallback() {
	return [
		{ id: 100630, nome: 'Beleza', slug: 'beleza', marketplace: 'shopee' },
		{ id: 100640, nome: 'Perfumaria', slug: 'perfumaria', marketplace: 'shopee' },
		{ id: 100664, nome: 'Cuidados com a Pele', slug: 'cuidados-pele', marketplace: 'shopee' },
		{ id: 100663, nome: 'Maquiagem', slug: 'maquiagem', marketplace: 'shopee' },
		{ id: 100659, nome: 'Cuidados com o Cabelo', slug: 'cuidados-cabelo', marketplace: 'shopee' },
		{ id: 100636, nome: 'Casa & Decoração', slug: 'casa-decoracao', marketplace: 'shopee' },
		{ id: 100637, nome: 'Moda', slug: 'moda', marketplace: 'shopee' },
		{ id: 100631, nome: 'Saúde & Bem-estar', slug: 'saude-bem-estar', marketplace: 'shopee' },
		{ id: 100644, nome: 'Áudio & Eletrônicos', slug: 'audio-eletronicos', marketplace: 'shopee' },
		{ id: 100011, nome: 'Roupas Femininas', slug: 'roupas-femininas', marketplace: 'shopee' },
		{ id: 100017, nome: 'Roupas Masculinas', slug: 'roupas-masculinas', marketplace: 'shopee' },
		{ id: 100012, nome: 'Calçados', slug: 'calcados', marketplace: 'shopee' },
		{ id: 100632, nome: 'Brinquedos & Bebês', slug: 'brinquedos-bebes', marketplace: 'shopee' },
		{ id: 100643, nome: 'Papelaria & Livros', slug: 'papelaria-livros', marketplace: 'shopee' },
		{ id: 100001, nome: 'Alimentos', slug: 'alimentos', marketplace: 'shopee' },
		{ id: 100658, nome: 'Manicure & Pedicure', slug: 'manicure-pedicure', marketplace: 'shopee' },
		{ id: 100633, nome: 'Acessórios & Bolsas', slug: 'acessorios-bolsas', marketplace: 'shopee' },
		{ id: 100009, nome: 'Celulares', slug: 'celulares', marketplace: 'shopee' },
	];
}
