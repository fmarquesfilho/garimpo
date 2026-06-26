// Filtros de curadoria persistidos no navegador. Resolve a perda da palavra-chave
// (e dos demais filtros) quando se troca de tela: o estado vive aqui, não só no
// componente da página. É também a base das "buscas salvas".
import { writable } from 'svelte/store';
import { browser } from '$app/environment';

const CHAVE = 'garimpo:filtros:v1';

export const FILTROS_PADRAO = {
	modo: 'nicho', // nicho | diversificada | comparar
	busca: '',
	categoria: 'cosméticos',
	comissaoMin: 0.07,
	vendasMin: 5,
	notaMin: 0,
	quantos: 20,
	explorar: false
};

function inicial() {
	if (browser) {
		try {
			const bruto = localStorage.getItem(CHAVE);
			if (bruto) return { ...FILTROS_PADRAO, ...JSON.parse(bruto) };
		} catch {
			/* storage corrompido: usa o padrão */
		}
	}
	return { ...FILTROS_PADRAO };
}

function criar() {
	const { subscribe, set, update } = writable(inicial());
	if (browser) {
		subscribe((estado) => {
			try {
				localStorage.setItem(CHAVE, JSON.stringify(estado));
			} catch {
				/* cota cheia: ignora */
			}
		});
	}
	return {
		subscribe,
		set,
		update,
		/** Aplica um conjunto de filtros (ex.: ao carregar uma busca salva). */
		aplicar(parcial) {
			update((e) => ({ ...e, ...parcial }));
		},
		repor() {
			set({ ...FILTROS_PADRAO });
		}
	};
}

export const filtros = criar();
