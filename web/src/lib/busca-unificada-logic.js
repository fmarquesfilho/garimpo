/**
 * Lógica pura do componente BuscaUnificada.
 * Conversão entre estado interno do componente e payloads de API.
 * Testável com Vitest sem dependências de DOM/API.
 */

/**
 * Converte o estado interno do componente para o payload do POST /api/buscas.
 * @param {object} config
 * @returns {object} payload compatível com SyncBuscaRequest
 */
export function configToPayload(config) {
	const payload = {};
	if (config.id) payload.id = config.id;

	const kws = (config.keywords ?? []).filter((k) => k.trim());
	if (kws.length > 0) payload.keywords = kws;

	// Campos opcionais: incluídos apenas quando têm valor significativo
	const optionalFields = [
		['shop_ids', config.shopIds, (v) => v?.length > 0],
		['shop_names', config.shopNomes, (v) => v && Object.keys(v).length > 0],
		['cron', config.cron, Boolean],
		['comissao_min', config.comissaoMin, (v) => v > 0],
		['vendas_min', config.vendasMin, (v) => v > 0],
		['categorias', config.categorias, (v) => v?.length > 0],
		['fontes', config.fontes, (v) => v?.length > 0],
		['marketplaces', config.marketplaces, Boolean]
	];

	for (const [key, value, predicate] of optionalFields) {
		if (predicate(value)) payload[key] = value;
	}

	return payload;
}

/**
 * Converte um item do response GET /api/buscas para o estado interno do componente.
 * @param {object} busca — item de buscas[]
 * @returns {object} config
 */
export function payloadToConfig(busca) {
	return {
		id: busca.id,
		keywords: busca.keywords ?? [],
		shopIds: busca.shop_ids ?? [],
		shopNomes: busca.shop_names ?? {},
		comissaoMin: busca.comissao_min ?? 0,
		vendasMin: busca.vendas_min ?? 0,
		categorias: busca.categorias ?? [],
		fontes: busca.fontes ?? [],
		cron: busca.cron ?? null,
		marketplaces: busca.marketplaces ?? 'shopee'
	};
}

/**
 * Gera string compacta para modo colapsado do componente.
 * Ex: "sérum" · 2 lojas · 2 filtros · ⏱ a cada 8h
 */
export function gerarResumo(config) {
	const partes = [];
	const kw = (config.keywords ?? []).join(', ');
	if (kw) partes.push(`"${kw}"`);
	const lojas = config.shopIds?.length ?? 0;
	if (lojas > 0) partes.push(`${lojas} ${lojas === 1 ? 'loja' : 'lojas'}`);
	const filtros = contarFiltrosAtivos(config);
	if (filtros > 0) partes.push(`${filtros} ${filtros === 1 ? 'filtro' : 'filtros'}`);
	if (config.cron) partes.push(`⏱ ${cronLabel(config.cron)}`);
	return partes.join(' · ') || 'Nenhum filtro ativo';
}

/**
 * Conta filtros não-default ativos na configuração.
 */
export function contarFiltrosAtivos(config) {
	let count = 0;
	if (config.comissaoMin > 0.07) count++;
	if (config.vendasMin > 0) count++;
	if (config.categorias?.length > 0) count++;
	return count;
}

/**
 * Converte expressão cron para label legível.
 */
export function cronLabel(cron) {
	if (!cron) return '';
	if (cron === '0 */8 * * *') return 'a cada 8h';
	if (cron === '0 */12 * * *') return 'a cada 12h';
	if (cron === '0 */6 * * *') return 'a cada 6h';
	if (cron === '0 9 * * *') return 'diária 9h';
	if (cron === '0 0 * * *') return 'diária 0h';
	return cron;
}

/**
 * Gera label para chip de busca salva.
 */
export function gerarLabelBusca(config) {
	const kw = (config.keywords ?? []).slice(0, 2).join(', ');
	const lojas = config.shopIds?.length ?? 0;
	const cats = config.categorias?.length ?? 0;
	let label = kw || (cats > 0 ? `${cats} ${cats === 1 ? 'categoria' : 'categorias'}` : '(sem keywords)');
	if (lojas > 0) label += ` + ${lojas} ${lojas === 1 ? 'loja' : 'lojas'}`;
	return label;
}
