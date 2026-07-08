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
		todos = todos.filter((r) => !r.comissao || r.comissao >= comissaoMin);
	}
	if (vendasMin > 0) {
		todos = todos.filter((r) => !r.vendas || r.vendas >= vendasMin);
	}

	return todos;
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
