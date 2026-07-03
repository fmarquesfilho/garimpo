/**
 * Store reativo de favoritos. Persiste em localStorage com sync para o servidor.
 * Exporta um store Svelte-compatível ($favoritos) e funções de manipulação.
 */
import { writable, get } from 'svelte/store';
import { listarFavoritos, salvarFavorito, removerFavorito } from './api.js';

const STORAGE_KEY = 'garimpei_favoritos';

function carregarLocal() {
	if (typeof localStorage === 'undefined') return [];
	try {
		return JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]');
	} catch {
		return [];
	}
}

function salvarLocal(favoritos) {
	if (typeof localStorage === 'undefined') return;
	localStorage.setItem(STORAGE_KEY, JSON.stringify(favoritos));
}

const { subscribe, set, update } = writable(carregarLocal());

// Mantém localStorage sincronizado
subscribe((valor) => salvarLocal(valor));

/** Sincroniza favoritos do servidor (substitui o local). */
async function sincronizar() {
	try {
		const r = await listarFavoritos();
		const lista = r?.favoritos ?? [];
		set(lista);
	} catch {
		// Falha silenciosa — usa o que tem no localStorage
	}
}

/** Verifica se um produto está favoritado (por produto_id ou id). */
function isFavorito(produtoId) {
	const lista = get({ subscribe });
	return lista.some((f) => f.produto_id === produtoId || f.id === produtoId);
}

/** Adiciona um produto aos favoritos. */
async function adicionar(produto) {
	const fav = {
		produto_id: produto.produto_id || produto.id || '',
		nome: produto.nome || '',
		preco: produto.preco || 0,
		comissao: produto.comissao || 0,
		link: produto.link || '',
		imagem: produto.imagem || '',
		loja: produto.loja || '',
		categoria: produto.categoria || '',
		origem: produto.origem || ''
	};

	// Otimista: adiciona localmente primeiro
	update((lista) => {
		const existe = lista.some((f) => f.produto_id === fav.produto_id);
		if (existe) return lista;
		return [{ ...fav, salvo_em: new Date().toISOString() }, ...lista];
	});

	// Sync com servidor (best-effort)
	try {
		await salvarFavorito(fav);
	} catch {
		// Já está salvo localmente — sync acontecerá depois
	}
}

/** Remove um produto dos favoritos. */
async function remover(produtoId) {
	// Otimista: remove localmente primeiro
	update((lista) => lista.filter((f) => f.produto_id !== produtoId && f.id !== produtoId));

	// Sync com servidor (best-effort)
	try {
		await removerFavorito(produtoId);
	} catch {
		// Já removido localmente
	}
}

/** Toggle: favorita se não está, desfavorita se já está. */
async function toggle(produto) {
	const id = produto.produto_id || produto.id || '';
	if (isFavorito(id)) {
		await remover(id);
	} else {
		await adicionar(produto);
	}
}

export const favoritos = {
	subscribe,
	sincronizar,
	isFavorito,
	adicionar,
	remover,
	toggle
};
