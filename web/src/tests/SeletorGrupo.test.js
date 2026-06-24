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

	it('renderiza input de busca quando tem grupos', () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });
		expect(screen.getByPlaceholderText('Digite para buscar um grupo…')).toBeInTheDocument();
	});
});

describe('SeletorGrupo — seleção múltipla', () => {
	it('selecionar um grupo chama onselect com o ID', async () => {
		const onselect = vi.fn();
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);

		const botao = screen.getByText('#1 Garimpo Hoje');
		await fireEvent.click(botao);

		expect(onselect).toHaveBeenCalledWith('120363430000000000@g.us');
	});

	it('selecionar dois grupos emite IDs separados por vírgula', async () => {
		const onselect = vi.fn();
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);
		await fireEvent.click(screen.getByText('#1 Garimpo Hoje'));

		// Segundo grupo
		const input2 = screen.getByPlaceholderText('Adicionar outro grupo…');
		await fireEvent.focus(input2);
		await fireEvent.click(screen.getByText('#08 AVANÇADO VOE'));

		expect(onselect).toHaveBeenLastCalledWith(
			'120363430000000000@g.us,120363410893012870@g.us'
		);
	});

	it('grupo já selecionado não aparece na lista', async () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);
		await fireEvent.click(screen.getByText('#1 Garimpo Hoje'));

		// Reabre o dropdown
		const input2 = screen.getByPlaceholderText('Adicionar outro grupo…');
		await fireEvent.focus(input2);

		// Garimpo Hoje não deve estar na lista
		const items = screen.queryAllByRole('button').filter(b => b.closest('ul'));
		const nomes = items.map(b => b.textContent);
		expect(nomes).not.toContain('#1 Garimpo Hoje');
		expect(nomes).toContain('#08 AVANÇADO VOE');
	});

	it('mostra chips dos grupos selecionados', async () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);
		await fireEvent.click(screen.getByText('#1 Garimpo Hoje'));

		expect(screen.getByText('#1 Garimpo Hoje')).toBeInTheDocument();
	});

	it('remover chip remove grupo e emite nova lista', async () => {
		const onselect = vi.fn();
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);
		await fireEvent.click(screen.getByText('#1 Garimpo Hoje'));

		// Remove o chip
		const removeBtn = screen.getByTitle('Remover');
		await fireEvent.click(removeBtn);

		expect(onselect).toHaveBeenLastCalledWith('');
	});

	it('limita a 5 grupos', async () => {
		const onselect = vi.fn();
		// Cria 6 grupos para testar o limite
		const muitos = [
			...gruposMock,
			{ id: '999@g.us', nome: 'Grupo Extra' }
		];
		render(SeletorGrupo, { props: { grupos: muitos, onselect } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');

		// Seleciona 5 grupos
		for (let i = 0; i < 5; i++) {
			const inp = screen.queryByPlaceholderText('Digite para buscar um grupo…')
				|| screen.queryByPlaceholderText('Adicionar outro grupo…');
			await fireEvent.focus(inp);
			const items = screen.getAllByRole('button').filter(b => b.closest('ul'));
			await fireEvent.click(items[0]);
		}

		// Deve mostrar mensagem de limite
		expect(screen.getByText(/Limite de 5 grupos/)).toBeInTheDocument();
	});
});

describe('SeletorGrupo — filtro', () => {
	it('digitar filtra a lista', async () => {
		render(SeletorGrupo, { props: { grupos: gruposMock, onselect: () => {} } });

		const input = screen.getByPlaceholderText('Digite para buscar um grupo…');
		await fireEvent.focus(input);
		await fireEvent.input(input, { target: { value: 'pipa' } });

		const items = screen.getAllByRole('button').filter(b => b.closest('ul'));
		expect(items.length).toBe(1);
		expect(items[0].textContent).toBe('Famílias da Pipa');
	});
});
