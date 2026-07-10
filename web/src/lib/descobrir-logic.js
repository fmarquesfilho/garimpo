/**
 * Lógica pura da página Descobrir — funções de filtragem e detecção de loja.
 * Sem dependências de API/Firebase — testável com Vitest diretamente.
 */

/**
 * Detecta se o termo de busca é nome de uma loja monitorada.
 */
export function encontrarLojaPorNome(termo, buscasComLojas) {
	if (!termo) return null;
	const t = termo.toLowerCase();
	return (
		buscasComLojas.find((b) => {
			const nome = (b.nome || b.id || '').toLowerCase();
			return nome.includes(t) || t.includes(nome);
		}) ?? null
	);
}

/**
 * Monta a lista final de resultados aplicando todos os filtros client-side.
 * @param {Object} opts
 * @param {{curadoria: boolean, quedas: boolean, novos: boolean, favoritos: boolean, lojas?: boolean}} opts.fontes
 * @param {any[]} opts.dadosCuradoria
 * @param {any[]} opts.dadosQuedas
 * @param {any[]} opts.dadosNovos
 * @param {any[]} [opts.dadosLojas]
 * @param {any[]} [opts.favoritos]
 * @param {string} [opts.busca]
 * @param {string[]} [opts.categorias]
 * @param {number} [opts.comissaoMin]
 * @param {number} [opts.vendasMin]
 */
export function montarResultados({
	fontes,
	dadosCuradoria,
	dadosQuedas,
	dadosNovos,
	dadosLojas,
	favoritos,
	busca,
	categorias,
	comissaoMin,
	vendasMin
}) {
	let todos = [];
	if (fontes.curadoria) todos.push(...dadosCuradoria);
	if (fontes.quedas) todos.push(...dadosQuedas);
	if (fontes.novos) todos.push(...dadosNovos);
	if (fontes.lojas && dadosLojas?.length) {
		todos.push(...dadosLojas.map((p) => (p._fonte ? p : { ...p, _fonte: 'loja' })));
	}
	if (fontes.favoritos) {
		const favs = (favoritos ?? []).map((f) => ({ ...f, id: f.produto_id, _fonte: 'favorito' }));
		todos.push(...favs);
	}

	const termo = (busca ?? '').trim().toLowerCase();
	if (termo) {
		todos = todos.filter(
			(r) => (r.nome ?? '').toLowerCase().includes(termo) || (r.loja ?? '').toLowerCase().includes(termo)
		);
	}

	const cats = (categorias ?? []).map((c) => c.toLowerCase());
	if (cats.length > 0) {
		todos = todos.filter((r) => !r.categoria || cats.some((c) => (r.categoria ?? '').toLowerCase().includes(c)));
	}

	if (comissaoMin > 0) {
		todos = todos.filter((r) => r.comissao == null || r.comissao >= comissaoMin);
	}
	if (vendasMin > 0) {
		todos = todos.filter((r) => r.vendas == null || r.vendas >= vendasMin);
	}

	return todos;
}

/**
 * Agrupa a lista crua de categorias (`{ id, nome, slug, marketplace }[]` de
 * /api/categorias) em itens de autocomplete `{ nome, marketplaces[] }`, unindo
 * uma mesma categoria presente em mais de um marketplace.
 * @param {Array<{nome?:string, marketplace?:string}|string>} [categorias]
 * @returns {{nome:string, marketplaces:string[]}[]}
 */
export function agruparCategoriasPorMarketplace(categorias) {
	const mapa = new Map();
	for (const c of categorias ?? []) {
		const obj = typeof c === 'string' ? { nome: c } : c;
		const nome = obj.nome;
		if (!nome) continue;
		if (!mapa.has(nome)) mapa.set(nome, new Set());
		if (obj.marketplace) mapa.get(nome).add(obj.marketplace);
	}
	return [...mapa.entries()]
		.map(([nome, mkts]) => ({ nome, marketplaces: [...mkts].sort() }))
		.sort((a, b) => a.nome.localeCompare(b.nome));
}

/**
 * Deriva a lista de lojas monitoradas para o autocomplete, a partir das buscas
 * salvas (uma busca pode conter várias lojas). Dedup por shopId.
 * @param {any[]} buscasSalvas — registros de /api/buscas
 * @returns {{id:string, nome:string, marketplace:string, origem:string|null, monitorada:boolean, cron:string}[]}
 */
export function listarLojasMonitoradas(buscasSalvas) {
	const vistos = new Map();
	for (const b of buscasSalvas ?? []) {
		const ids = b.shop_ids ?? b.shopIds ?? [];
		const nomes = b.shop_names ?? b.shopNomes ?? {};
		for (const id of ids) {
			if (vistos.has(id)) continue;
			vistos.set(id, buildLojaEntry(id, nomes, b));
		}
	}
	return [...vistos.values()].sort((a, b) => a.nome.localeCompare(b.nome));
}

function buildLojaEntry(id, nomes, busca) {
	return {
		id: String(id),
		nome: nomes[String(id)] || nomes[id] || busca.nome || String(id),
		marketplace: busca.marketplaces || busca.marketplace || 'shopee',
		origem: busca.origem_padrao ?? busca.origemPadrao ?? null,
		monitorada: Boolean(busca.cron),
		cron: busca.cron ?? ''
	};
}

/**
 * Gera opções para o ToggleGroup de fontes com badges de contagem.
 */
export function buildFonteOpcoes({ contagemCuradoria, contagemQuedas, contagemNovos, contagemLojas, totalFavoritos }) {
	return [
		{
			value: 'curadoria',
			label: '🔍 Busca',
			badge: contagemCuradoria,
			badgeColor: 'ouro',
			title: 'Busca por palavra-chave na API de afiliados Shopee'
		},
		{
			value: 'quedas',
			label: '📉 Quedas',
			badge: contagemQuedas,
			badgeColor: 'sucesso',
			title: 'Produtos que caíram de preço nas lojas monitoradas'
		},
		{
			value: 'novos',
			label: '🆕 Novos',
			badge: contagemNovos,
			badgeColor: 'rosa',
			title: 'Produtos novos detectados nas lojas monitoradas'
		},
		{
			value: 'favoritos',
			label: '⭐ Favoritos',
			badge: totalFavoritos,
			badgeColor: 'ouro',
			title: 'Produtos que você salvou como favorito'
		},
		{
			value: 'lojas',
			label: '🏪 Lojas',
			badge: contagemLojas,
			badgeColor: 'ouro',
			title: 'Produtos das lojas que você monitora'
		}
	];
}
