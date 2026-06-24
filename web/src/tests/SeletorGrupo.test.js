import { render, fireEvent, screen, cleanup } from '@testing-library/svelte';
import { describe, it, expect, vi, afterEach } from 'vitest';
import SeletorGrupo from '$lib/SeletorGrupo.svelte';

afterEach(() => cleanup());

const gruposMock = [
	{ id: '120363430000000000@g.us', nome: '#1 Garimpo Hoje' },
	{ id: '120363410893012870@g.us', nome: '#08 AVANÇADO VOE' },
	{ id: '120363426313232441@g.us', nome: '#96 NOSSO GRUPINHO' },
	{ id: '120363156757082979@g.us', nome: 'Ofertas | Beleza na Web' },
	{ id: '558491629647-1486926372@g.us', nome: 'Famílias da Pipa' }
];

describe('SeletorGrupo — renderização', () => {
	it('mostra "Carregando…" quando carregando=true', () => {
		render(SeletorGrupo, { props: { grupos: [], carregando: true, onselect: () => {} } });
		expect(screen.getByText('Carregando grupos…')).toBeInTheDocument();
	});

	it('mostra "Nenhum grupo encontrado" quando lista vazia', () => {
		render(SeletorGrupo, { props: { grupos: [], carregando: false, onselect: () => {} } });
		expect(screen.getByText('Nenhum grupo encontrado')).toBeInTheDocument();
	});

	it('mostra mensagem de erro e input manual quando há erro', () => {
		render(SeletorGrupo, { props: { grupos: [], erro: 'API falhou', onselect: () => {} } });
		expect(screen.getByText('API falhou')).toBeInTheDocument();
		expect(screen.getByPlaceholderText(/ID do grupo/)).toBeInTheDocument();
	});

	it('renderiza select com todos os grupos', () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });
		const select = screen.getByRole('combobox');
		// 5 grupos + 1 placeholder
		expect(select.options).toHaveLength(6);
		expect(select.options[1].textContent).toBe('#1 Garimpo Hoje');
	});

	it('renderiza campo de filtro', () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });
		expect(screen.getByPlaceholderText('Filtrar grupos…')).toBeInTheDocument();
	});
});

describe('SeletorGrupo — seleção propaga valor', () => {
	it('chama onselect com o ID do grupo ao selecionar', async () => {
		const onselect = vi.fn();
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect } });

		const select = screen.getByRole('combobox');
		await fireEvent.change(select, { target: { value: '120363430000000000@g.us' } });

		expect(onselect).toHaveBeenCalledTimes(1);
		expect(onselect).toHaveBeenCalledWith('120363430000000000@g.us');
	});

	it('chama onselect com string vazia ao voltar pro placeholder', async () => {
		const onselect = vi.fn();
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect } });

		const select = screen.getByRole('combobox');
		// Primeiro seleciona um grupo
		await fireEvent.change(select, { target: { value: '120363430000000000@g.us' } });
		// Depois volta pro placeholder
		await fireEvent.change(select, { target: { value: '' } });

		expect(onselect).toHaveBeenLastCalledWith('');
	});

	it('chama onselect várias vezes ao trocar seleção', async () => {
		const onselect = vi.fn();
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect } });

		const select = screen.getByRole('combobox');

		await fireEvent.change(select, { target: { value: '120363430000000000@g.us' } });
		await fireEvent.change(select, { target: { value: '120363410893012870@g.us' } });
		await fireEvent.change(select, { target: { value: '558491629647-1486926372@g.us' } });

		expect(onselect).toHaveBeenCalledTimes(3);
		expect(onselect).toHaveBeenLastCalledWith('558491629647-1486926372@g.us');
	});

	it('o valor selecionado NÃO é resetado por re-render (o bug original)', async () => {
		const onselect = vi.fn();
		const { rerender } = render(SeletorGrupo, {
			props: { grupos: gruposMock, onselect }
		});

		const select = screen.getByRole('combobox');

		// Seleciona um grupo
		await fireEvent.change(select, { target: { value: '120363430000000000@g.us' } });
		expect(onselect).toHaveBeenCalledWith('120363430000000000@g.us');

		// Re-render com novos props (simula o parent re-renderizando)
		await rerender({ grupos: gruposMock, onselect });

		// A option selecionada deve ainda estar marcada (estado interno do componente)
		const selectedOption = select.querySelector('option[selected]');
		expect(selectedOption).not.toBeNull();
		expect(selectedOption.value).toBe('120363430000000000@g.us');
	});
});

describe('SeletorGrupo — filtro', () => {
	it('filtrar reduz as opções visíveis', async () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });

		const filtro = screen.getByPlaceholderText('Filtrar grupos…');
		await fireEvent.input(filtro, { target: { value: 'garimpo' } });

		const select = screen.getByRole('combobox');
		// 1 grupo filtrado + 1 placeholder
		expect(select.options).toHaveLength(2);
		expect(select.options[1].textContent).toBe('#1 Garimpo Hoje');
	});

	it('limpar filtro mostra todos os grupos', async () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });

		const filtro = screen.getByPlaceholderText('Filtrar grupos…');

		// Filtra
		await fireEvent.input(filtro, { target: { value: 'garimpo' } });
		let select = screen.getByRole('combobox');
		expect(select.options).toHaveLength(2);

		// Limpa
		await fireEvent.input(filtro, { target: { value: '' } });
		select = screen.getByRole('combobox');
		expect(select.options).toHaveLength(6); // 5 + placeholder
	});

	it('filtrar e selecionar propaga o valor correto', async () => {
		const onselect = vi.fn();
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect } });

		// Filtra por "pipa"
		const filtro = screen.getByPlaceholderText('Filtrar grupos…');
		await fireEvent.input(filtro, { target: { value: 'pipa' } });

		// Seleciona o resultado filtrado
		const select = screen.getByRole('combobox');
		await fireEvent.change(select, { target: { value: '558491629647-1486926372@g.us' } });

		expect(onselect).toHaveBeenCalledWith('558491629647-1486926372@g.us');
	});

	it('seleção persiste após digitar e apagar filtro', async () => {
		const onselect = vi.fn();
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect } });

		// Seleciona um grupo primeiro
		const select = screen.getByRole('combobox');
		await fireEvent.change(select, { target: { value: '120363430000000000@g.us' } });

		const filtro = screen.getByPlaceholderText('Filtrar grupos…');

		// Digita algo e apaga
		await fireEvent.input(filtro, { target: { value: 'xyz' } });
		await fireEvent.input(filtro, { target: { value: '' } });

		// A option com selected deve ainda estar correta
		const selectedOption = select.querySelector('option[selected]');
		expect(selectedOption).not.toBeNull();
		expect(selectedOption.value).toBe('120363430000000000@g.us');
	});
});

describe('SeletorGrupo — modo erro (input manual)', () => {
	it('input manual chama onselect ao digitar', async () => {
		const onselect = vi.fn();
		render(SeletorGrupo, { props: { grupos: [], erro: 'falhou', onselect } });

		const input = screen.getByPlaceholderText(/ID do grupo/);
		await fireEvent.input(input, { target: { value: '123-456@g.us' } });

		expect(onselect).toHaveBeenCalledWith('123-456@g.us');
	});
});
