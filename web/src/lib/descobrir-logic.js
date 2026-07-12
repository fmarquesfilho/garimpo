/**
 * Lógica pura da página Descobrir — funções de filtragem e detecção de loja.
 * Sem dependências de API/Firebase — testável com Vitest diretamente.
 */


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
