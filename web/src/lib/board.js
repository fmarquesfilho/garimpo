// Quadro de operação (Kanban) persistido no navegador.
// Colunas fixas; cards movem entre elas. Limites de WIP tornam o gargalo visível.
import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export const COLUNAS = [
	{ id: 'selecionados', titulo: 'Selecionados', wip: 5 },
	{ id: 'producao', titulo: 'Em produção', wip: 2 },
	{ id: 'publicado', titulo: 'Publicado', wip: null },
	{ id: 'analise', titulo: 'Em análise', wip: null }
];

const CHAVE = 'garimpo:quadro:v1';

function estadoInicial() {
	if (browser) {
		try {
			const bruto = localStorage.getItem(CHAVE);
			if (bruto) return JSON.parse(bruto);
		} catch {
			/* ignora storage corrompido */
		}
	}
	return { selecionados: [], producao: [], publicado: [], analise: [] };
}

function criarQuadro() {
	const { subscribe, update, set } = writable(estadoInicial());

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
		/** Adiciona um candidato à coluna "Selecionados" (sem duplicar por id). */
		selecionar(candidato) {
			update((estado) => {
				const jaExiste = Object.values(estado)
					.flat()
					.some((c) => c.id === candidato.id);
				if (jaExiste) return estado;
				return { ...estado, selecionados: [{ ...candidato, estrategia: candidato.estrategia ?? null }, ...estado.selecionados] };
			});
		},
		/** Move um card de uma coluna para outra. */
		mover(id, de, para) {
			update((estado) => {
				const card = estado[de].find((c) => c.id === id);
				if (!card) return estado;
				return {
					...estado,
					[de]: estado[de].filter((c) => c.id !== id),
					[para]: [card, ...estado[para]]
				};
			});
		},
		/** Remove um card de qualquer coluna. */
		remover(id, coluna) {
			update((estado) => ({ ...estado, [coluna]: estado[coluna].filter((c) => c.id !== id) }));
		},
		limpar() {
			set({ selecionados: [], producao: [], publicado: [], analise: [] });
		}
	};
}

export const quadro = criarQuadro();
