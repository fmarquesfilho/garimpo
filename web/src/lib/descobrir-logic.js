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
	return buscasComLojas.find(b => {
		const nome = (b.nome || b.id || '').toLowerCase();
		return nome.includes(t) || t.includes(nome);
	}) ?? null;
}

/**
 * Monta a lista final de resultados aplicando todos os filtros client-side.
 */
export function montarResultados({ fontes, dadosCuradoria, dadosQuedas, dadosNovos, favoritos, busca, categorias, comissaoMin, vendasMin, notaMin }) {
	let todos = [];
	if (fontes.curadoria) todos.push(...dadosCuradoria);
	if (fontes.quedas) todos.push(...dadosQuedas);
	if (fontes.novos) todos.push(...dadosNovos);
	if (fontes.favoritos) {
		const favs = (favoritos ?? []).map(f => ({ ...f, id: f.produto_id, _fonte: 'favorito' }));
		todos.push(...favs);
	}

	const termo = (busca ?? '').trim().toLowerCase();
	if (termo) {
		todos = todos.filter(r =>
			(r.nome ?? '').toLowerCase().includes(termo) ||
			(r.loja ?? '').toLowerCase().includes(termo)
		);
	}

	const cats = (categorias ?? []).map(c => c.toLowerCase());
	if (cats.length > 0) {
		todos = todos.filter(r =>
			!r.categoria || cats.some(c => (r.categoria ?? '').toLowerCase().includes(c))
		);
	}

	if (comissaoMin > 0) {
		todos = todos.filter(r => !r.comissao || r.comissao >= comissaoMin);
	}
	if (vendasMin > 0) {
		todos = todos.filter(r => !r.vendas || r.vendas >= vendasMin);
	}
	if (notaMin > 0) {
		todos = todos.filter(r => !r.avaliacao || r.avaliacao >= notaMin);
	}

	return todos;
}
