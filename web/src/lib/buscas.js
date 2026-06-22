// Buscas salvas: cada busca é identificada por um ID gerado da keyword principal.
// Uma busca pode ter múltiplas keywords (ex.: "kenzo", "shiseido" numa busca
// "perfumaria japonesa"). Os filtros são persistidos no servidor (BigQuery) para
// sobreviver à troca de dispositivo; o localStorage serve de cache imediato.
import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { listarBuscasServidor, sincronizarBusca } from './api.js';

const CHAVE = 'garimpo:buscas:v2';

/** Gera um ID slug a partir de uma string (mesma lógica do Go). */
export function slugificar(s) {
	return (s ?? '')
		.toLowerCase()
		.trim()
		.normalize('NFD')
		.replace(/[\u0300-\u036f]/g, '') // remove acentos
		.replace(/[^a-z0-9\s-]/g, '')
		.replace(/[\s_]+/g, '-')
		.replace(/-+/g, '-')
		.replace(/^-|-$/g, '')
		|| 'busca';
}

function inicial() {
	if (browser) {
		try {
			const bruto = localStorage.getItem(CHAVE);
			if (bruto) return JSON.parse(bruto);
		} catch {
			/* ignora */
		}
		// migração da chave antiga (v1 usava nome + keyword string)
		try {
			const velho = localStorage.getItem('garimpo:buscas:v1');
			if (velho) {
				const lista = JSON.parse(velho);
				return lista.map((b) => migrarBuscaLegada(b));
			}
		} catch {
			/* ignora */
		}
	}
	return [];
}

/** Converte um registro do formato antigo (nome + keyword string) para o novo. */
function migrarBuscaLegada(b) {
	if (b.id) return b; // já no novo formato
	const keywords = b.keywords ?? (b.keyword ? [b.keyword] : []);
	const id = b.id ?? slugificar(b.nome ?? keywords[0] ?? 'busca');
	return {
		id,
		keywords,
		categoria: b.categoria ?? 'cosméticos',
		estrategia: b.estrategia ?? 'nicho',
		comissao_min: b.comissao_min ?? 0.07,
		vendas_min: b.vendas_min ?? 5,
		nota_min: b.nota_min ?? 0,
		top: b.top ?? 9,
		cron: b.cron ?? '',
		salvo_em: b.salvo_em ?? new Date().toISOString()
	};
}

function criar() {
	const { subscribe, set, update } = writable(inicial());

	if (browser) {
		subscribe((lista) => {
			try {
				localStorage.setItem(CHAVE, JSON.stringify(lista));
			} catch {
				/* ignora */
			}
		});
	}

	return {
		subscribe,
		set,

		/**
		 * Salva/atualiza um perfil (por ID) localmente e sincroniza no servidor.
		 * @param {Object} busca - objeto com id, keywords[], categoria, estrategia, etc.
		 */
		salvar(busca) {
			const b = { ...busca, salvo_em: new Date().toISOString() };
			// garante que keywords é sempre array
			if (!Array.isArray(b.keywords)) {
				b.keywords = b.keywords ? [b.keywords] : [];
			}
			// gera ID da primeira keyword se não tiver
			if (!b.id) {
				b.id = slugificar(b.keywords[0] ?? 'busca');
			}
			update((lista) => {
				const i = lista.findIndex((x) => x.id === b.id);
				if (i >= 0) {
					const copia = [...lista];
					copia[i] = b;
					return copia;
				}
				return [...lista, b];
			});
			sincronizarBusca(b); // best-effort
		},

		/** Adiciona uma keyword a uma busca existente. */
		adicionarKeyword(id, keyword) {
			update((lista) => {
				const i = lista.findIndex((x) => x.id === id);
				if (i < 0) return lista;
				const b = { ...lista[i] };
				const kw = keyword.trim();
				if (!kw || b.keywords.includes(kw)) return lista;
				b.keywords = [...b.keywords, kw];
				b.salvo_em = new Date().toISOString();
				const copia = [...lista];
				copia[i] = b;
				sincronizarBusca(b);
				return copia;
			});
		},

		/** Remove uma keyword de uma busca. Se ficar vazia, remove a busca. */
		removerKeyword(id, keyword) {
			update((lista) => {
				const i = lista.findIndex((x) => x.id === id);
				if (i < 0) return lista;
				const b = { ...lista[i] };
				b.keywords = b.keywords.filter((k) => k !== keyword);
				if (b.keywords.length === 0) {
					// sem keywords → remove a busca inteira
					sincronizarBusca({ id }, { remover: true });
					return lista.filter((x) => x.id !== id);
				}
				b.salvo_em = new Date().toISOString();
				const copia = [...lista];
				copia[i] = b;
				sincronizarBusca(b);
				return copia;
			});
		},

		/** Remove um perfil localmente e grava o tombstone no servidor. */
		remover(id) {
			update((lista) => lista.filter((x) => x.id !== id));
			sincronizarBusca({ id }, { remover: true });
		},

		/** Puxa do servidor (BigQuery) e substitui o local — servidor é a verdade.
		 *  Se o servidor retorna lista vazia, o local fica vazio. */
		async sincronizarDoServidor() {
			try {
				const r = await listarBuscasServidor();
				const doServidor = (r?.buscas ?? []).map(migrarBuscaLegada);
				// Servidor é a fonte de verdade: substitui o local inteiro.
				set(doServidor);
			} catch {
				/* offline ou sem servidor: fica só com o local */
			}
		}
	};
}

export const buscasSalvas = criar();
