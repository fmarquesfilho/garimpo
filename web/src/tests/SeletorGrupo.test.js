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
		expect(screen.getByPlaceholderText('Carregando grupos…')).toBeInTheDocument();
	});

	it('mostra "Nenhum grupo encontrado" quando lista vazia', () => {
		render(SeletorGrupo, { props: { grupos: [], carregando: false, onselect: () => {} } });
		expect(screen.getByPlaceholderText('Nenhum grupo encontrado')).toBeInTheDocument();
	});

	it('mostra mensagem de erro e input manual quando há erro', () => {
		render(SeletorGrupo, { props: { grupos: [], erro: 'API falhou', onselect: () => {} } });
		expect(screen.getByText('API falhou')).toBeInTheDocument();
		expect(screen.getByPlaceholderText(/ID do grupo/)).toBeInTheDocument();
	});

	it('renderiza input de busca quando tem grupos', () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });
		expect(screen.getByPlaceholderText('Digite para buscar um grupo…')).toBeInTheDocument();
	});
});

describe('SeletorGrupo — dropdown e seleção', () => {
	it('mostra dropdown ao focar no input', async () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);

		// Deve mostrar todos os grupos no dropdown
		const items = screen.getAllByRole('button').filter(b => b.closest('ul'));
		expect(items.length).toBe(5);
	});

	it('clicar num grupo chama onselect com o ID', async () => {
		const onselect = vi.fn();
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);

		// Clica no primeiro grupo
		const botaoGrupo = screen.getByText('#1 Garimpo Hoje');
		await fireEvent.click(botaoGrupo);

		expect(onselect).toHaveBeenCalledWith('120363430000000000@g.us');
	});

	it('após selecionar, o input mostra o nome do grupo', async () => {
		const onselect = vi.fn();
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);

		const botaoGrupo = screen.getByText('Famílias da Pipa');
		await fireEvent.click(botaoGrupo);

		expect(input.value).toBe('Famílias da Pipa');
	});

	it('após selecionar, input fica com estilo "selecionado"', async () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);

		const botaoGrupo = screen.getByText('#1 Garimpo Hoje');
		await fireEvent.click(botaoGrupo);

		expect(input.classList.contains('selecionado')).toBe(true);
	});
});

describe('SeletorGrupo — filtragem', () => {
	it('digitar filtra a lista de grupos', async () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);
		await fireEvent.input(input, { target: { value: 'garimpo' } });

		const items = screen.getAllByRole('button').filter(b => b.closest('ul'));
		expect(items.length).toBe(1);
		expect(items[0].textContent).toBe('#1 Garimpo Hoje');
	});

	it('filtro sem resultado mostra mensagem', async () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);
		await fireEvent.input(input, { target: { value: 'xyzabc123' } });

		expect(screen.getByText('Nenhum grupo encontrado')).toBeInTheDocument();
	});

	it('digitar após selecionar limpa a seleção e chama onselect vazio', async () => {
		const onselect = vi.fn();
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);

		// Seleciona
		const botaoGrupo = screen.getByText('#1 Garimpo Hoje');
		await fireEvent.click(botaoGrupo);
		expect(onselect).toHaveBeenCalledWith('120363430000000000@g.us');

		// Edita o texto
		await fireEvent.focus(input);
		await fireEvent.input(input, { target: { value: 'outro texto' } });

		// Deve ter chamado onselect com '' (limpou seleção)
		expect(onselect).toHaveBeenLastCalledWith('');
	});
});

describe('SeletorGrupo — limpar seleção', () => {
	it('botão limpar aparece após seleção', async () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);

		const botaoGrupo = screen.getByText('#1 Garimpo Hoje');
		await fireEvent.click(botaoGrupo);

		expect(screen.getByTitle('Limpar')).toBeInTheDocument();
	});

	it('clicar limpar reseta tudo e chama onselect vazio', async () => {
		const onselect = vi.fn();
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);

		const botaoGrupo = screen.getByText('#1 Garimpo Hoje');
		await fireEvent.click(botaoGrupo);

		const limpar = screen.getByTitle('Limpar');
		await fireEvent.click(limpar);

		expect(input.value).toBe('');
		expect(onselect).toHaveBeenLastCalledWith('');
	});
});
