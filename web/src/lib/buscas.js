// Buscas salvas (perfis de coleta): conjuntos nomeados de filtros, reusáveis
// manualmente e candidatos à coleta periódica. Vivem no localStorage (uso
// imediato) e são sincronizadas no servidor/BigQuery (para a coleta agendada não
// depender do navegador) via /api/buscas.
import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { listarBuscasServidor, sincronizarBusca } from './api.js';

const CHAVE = 'garimpo:buscas:v1';

function inicial() {
	if (browser) {
		try {
			const bruto = localStorage.getItem(CHAVE);
			if (bruto) return JSON.parse(bruto);
		} catch {
			/* ignora */
		}
	}
	return [];
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
		/** Salva/atualiza um perfil (por nome) localmente e sincroniza no servidor. */
		salvar(busca) {
			const b = { ...busca, salvo_em: new Date().toISOString() };
			update((lista) => {
				const i = lista.findIndex((x) => x.nome === b.nome);
				if (i >= 0) {
					const copia = [...lista];
					copia[i] = b;
					return copia;
				}
				return [...lista, b];
			});
			sincronizarBusca(b); // best-effort
		},
		/** Remove um perfil localmente e grava o tombstone no servidor. */
		remover(nome) {
			update((lista) => lista.filter((x) => x.nome !== nome));
			sincronizarBusca({ nome }, { remover: true });
		},
		/** Puxa do servidor (BigQuery) e funde com o local — servidor vence por nome. */
		async sincronizarDoServidor() {
			try {
				const r = await listarBuscasServidor();
				const doServidor = r?.buscas ?? [];
				if (doServidor.length === 0) return;
				update((local) => {
					const porNome = new Map(local.map((b) => [b.nome, b]));
					for (const b of doServidor) porNome.set(b.nome, b);
					return [...porNome.values()];
				});
			} catch {
				/* offline ou sem servidor: fica só com o local */
			}
		}
	};
}

export const buscasSalvas = criar();
