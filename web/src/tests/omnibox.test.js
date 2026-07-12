import { render, screen, cleanup, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, afterEach, vi } from 'vitest';
import Omnibox from '$lib/components/Omnibox.svelte';

afterEach(() => cleanup());

/** Engine falso: só o que o Omnibox lê + um spy em send(). */
function fakeEngine(overrides = {}) {
	return {
		ctx: {
			categoriasDisponiveis: [{ nome: 'Beleza' }, { nome: 'Casa' }],
			buscasSalvas: [],
			marketplacesFiltro: [],
			shopIds: [],
			categorias: [],
			...overrides
		},
		send: vi.fn()
	};
}

const lojas = [
	{ id: '920', nome: 'Glory of Seoul', marketplace: 'shopee' },
	{ id: '281', nome: 'Le Botanic', marketplace: 'shopee' }
];

function setup(overrides) {
	const engine = fakeEngine(overrides);
	render(Omnibox, { props: { engine, lojasMonitoradas: lojas } });
	return { engine, input: screen.getByRole('combobox') };
}

describe('Omnibox — render e dropdown', () => {
	it('renderiza um único campo combobox', () => {
		const { input } = setup();
		expect(input).toBeInTheDocument();
		expect(input.getAttribute('aria-expanded')).toBe('false');
	});

	it('digitar 2+ chars abre dropdown com sugestões agrupadas', async () => {
		const { input } = setup();
		await fireEvent.input(input, { target: { value: 'glo' } });
		expect(await screen.findByRole('listbox')).toBeInTheDocument();
		expect(screen.getByText('Glory of Seoul')).toBeInTheDocument();
		expect(input.getAttribute('aria-expanded')).toBe('true');
	});

	it('1 char não abre dropdown (minChars)', async () => {
		const { input } = setup();
		await fireEvent.input(input, { target: { value: 'g' } });
		expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
	});

	it('digitar keyword emite DIGITAR com a keyword resolvida', async () => {
		const { engine, input } = setup();
		await fireEvent.input(input, { target: { value: 'serum' } });
		expect(engine.send).toHaveBeenCalledWith({ type: 'DIGITAR', value: 'serum' });
	});

	it('token de loja não polui a keyword enviada à engine', async () => {
		const { engine, input } = setup();
		await fireEvent.input(input, { target: { value: 'serum @glo' } });
		expect(engine.send).toHaveBeenLastCalledWith({ type: 'DIGITAR', value: 'serum' });
	});
});

describe('Omnibox — seleção emite eventos', () => {
	it('clicar numa loja emite ADICIONAR_LOJA', async () => {
		const { engine, input } = setup();
		await fireEvent.input(input, { target: { value: '@glo' } });
		await fireEvent.click(screen.getByText('Glory of Seoul'));
		expect(engine.send).toHaveBeenCalledWith(expect.objectContaining({ type: 'ADICIONAR_LOJA' }));
		const call = engine.send.mock.calls.find((c) => c[0].type === 'ADICIONAR_LOJA');
		expect(call[0].loja.id).toBe('920');
	});

	it('selecionar categoria emite ADICIONAR_CATEGORIA', async () => {
		const { engine, input } = setup();
		await fireEvent.input(input, { target: { value: '#bel' } });
		await fireEvent.click(screen.getByText('Beleza'));
		expect(engine.send).toHaveBeenCalledWith(expect.objectContaining({ type: 'ADICIONAR_CATEGORIA', nome: 'Beleza' }));
	});
});

describe('Omnibox — teclado', () => {
	it('Escape fecha o dropdown', async () => {
		const { input } = setup();
		await fireEvent.input(input, { target: { value: 'glo' } });
		expect(screen.getByRole('listbox')).toBeInTheDocument();
		await fireEvent.keyDown(input, { key: 'Escape' });
		expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
	});

	it('ArrowDown destaca a primeira sugestão (aria-activedescendant)', async () => {
		const { input } = setup();
		await fireEvent.input(input, { target: { value: 'glo' } });
		await fireEvent.keyDown(input, { key: 'ArrowDown' });
		expect(input.getAttribute('aria-activedescendant')).toBe('omnibox-opt-0');
	});

	it('Enter sem sugestão destacada executa busca (DIGITAR)', async () => {
		const { engine, input } = setup();
		await fireEvent.input(input, { target: { value: 'serum' } });
		engine.send.mockClear();
		await fireEvent.keyDown(input, { key: 'Enter' });
		expect(engine.send).toHaveBeenCalledWith({ type: 'DIGITAR', value: 'serum' });
	});

	it('Enter com sugestão destacada seleciona (ADICIONAR_LOJA)', async () => {
		const { engine, input } = setup();
		await fireEvent.input(input, { target: { value: '@glo' } });
		await fireEvent.keyDown(input, { key: 'ArrowDown' });
		await fireEvent.keyDown(input, { key: 'Enter' });
		expect(engine.send).toHaveBeenCalledWith(expect.objectContaining({ type: 'ADICIONAR_LOJA' }));
	});
});

describe('Omnibox — degradação graceful', () => {
	it('sem lojas monitoradas, keyword ainda funciona', async () => {
		const engine = fakeEngine();
		render(Omnibox, { props: { engine, lojasMonitoradas: [] } });
		const input = screen.getByRole('combobox');
		await fireEvent.input(input, { target: { value: 'serum' } });
		expect(engine.send).toHaveBeenCalledWith({ type: 'DIGITAR', value: 'serum' });
	});
});
