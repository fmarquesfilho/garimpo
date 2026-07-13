import { render, screen, cleanup, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, afterEach, vi } from 'vitest';
import Omnibox from '$lib/components/Omnibox.svelte';

afterEach(() => cleanup());

/**
 * Engine mock: the new Omnibox is a pure renderer — reads engine.omnibox and
 * emits events via engine.send(). We provide the omnibox state and spy on send.
 */
function fakeEngine(omniboxOverrides = {}) {
	return {
		omnibox: {
			inputValue: '',
			aberto: false,
			highlightIdx: -1,
			modo: 'intencao',
			opcoes: [],
			placeholder: 'Buscar produtos, lojas ou categorias…',
			...omniboxOverrides
		},
		lojaCards: [],
		categoriaCards: [],
		send: vi.fn()
	};
}

describe('Omnibox — pure renderer', () => {
	it('renders a combobox input', () => {
		const engine = fakeEngine();
		render(Omnibox, { props: { engine } });
		const input = screen.getByRole('combobox');
		expect(input).toBeInTheDocument();
	});

	it('displays the placeholder from engine.omnibox', () => {
		const engine = fakeEngine({ placeholder: 'Custom placeholder' });
		render(Omnibox, { props: { engine } });
		const input = screen.getByRole('combobox');
		expect(input.getAttribute('placeholder')).toBe('Custom placeholder');
	});

	it('displays inputValue from engine.omnibox', () => {
		const engine = fakeEngine({ inputValue: 'serum' });
		render(Omnibox, { props: { engine } });
		const input = /** @type {HTMLInputElement} */ (screen.getByRole('combobox'));
		expect(input.value).toBe('serum');
	});

	it('aria-expanded reflects aberto + opcoes.length', () => {
		const engine = fakeEngine({ aberto: true, opcoes: [{ tipo: 'produtos', label: 'Test', icone: '🔎' }] });
		render(Omnibox, { props: { engine } });
		const input = screen.getByRole('combobox');
		expect(input.getAttribute('aria-expanded')).toBe('true');
	});

	it('aria-expanded is false when closed', () => {
		const engine = fakeEngine({ aberto: false });
		render(Omnibox, { props: { engine } });
		const input = screen.getByRole('combobox');
		expect(input.getAttribute('aria-expanded')).toBe('false');
	});
});

describe('Omnibox — events dispatched', () => {
	it('oninput dispatches OMNIBOX_INPUT', async () => {
		const engine = fakeEngine();
		render(Omnibox, { props: { engine } });
		const input = screen.getByRole('combobox');
		await fireEvent.input(input, { target: { value: 'serum' } });
		expect(engine.send).toHaveBeenCalledWith({ type: 'OMNIBOX_INPUT', value: 'serum' });
	});

	it('Enter dispatches OMNIBOX_KEYDOWN', async () => {
		const engine = fakeEngine({ aberto: true, opcoes: [{ tipo: 'produtos', label: 'T', icone: '🔎' }] });
		render(Omnibox, { props: { engine } });
		const input = screen.getByRole('combobox');
		await fireEvent.keyDown(input, { key: 'Enter' });
		expect(engine.send).toHaveBeenCalledWith({ type: 'OMNIBOX_KEYDOWN', key: 'Enter' });
	});

	it('ArrowDown dispatches OMNIBOX_KEYDOWN', async () => {
		const engine = fakeEngine();
		render(Omnibox, { props: { engine } });
		const input = screen.getByRole('combobox');
		await fireEvent.keyDown(input, { key: 'ArrowDown' });
		expect(engine.send).toHaveBeenCalledWith({ type: 'OMNIBOX_KEYDOWN', key: 'ArrowDown' });
	});

	it('Escape dispatches OMNIBOX_KEYDOWN', async () => {
		const engine = fakeEngine();
		render(Omnibox, { props: { engine } });
		const input = screen.getByRole('combobox');
		await fireEvent.keyDown(input, { key: 'Escape' });
		expect(engine.send).toHaveBeenCalledWith({ type: 'OMNIBOX_KEYDOWN', key: 'Escape' });
	});

	it('onfocus dispatches OMNIBOX_INPUT with current value', async () => {
		const engine = fakeEngine({ inputValue: 'hello' });
		render(Omnibox, { props: { engine } });
		const input = screen.getByRole('combobox');
		await fireEvent.focus(input);
		expect(engine.send).toHaveBeenCalledWith({ type: 'OMNIBOX_INPUT', value: 'hello' });
	});
});

describe('Omnibox — dropdown rendering', () => {
	it('renders listbox when aberto and opcoes exist', () => {
		const engine = fakeEngine({
			aberto: true,
			opcoes: [
				{ tipo: 'produtos', label: 'Pesquisar "serum" em Produtos', icone: '🔎' },
				{ tipo: 'lojas', label: 'Pesquisar "serum" em Lojas', icone: '🏪' }
			]
		});
		render(Omnibox, { props: { engine } });
		expect(screen.getByRole('listbox')).toBeInTheDocument();
		expect(screen.getAllByRole('option')).toHaveLength(2);
	});

	it('does not render listbox when closed', () => {
		const engine = fakeEngine({ aberto: false, opcoes: [{ tipo: 'produtos', label: 'T', icone: '🔎' }] });
		render(Omnibox, { props: { engine } });
		expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
	});

	it('clicking an option dispatches OMNIBOX_SELECIONAR with index', async () => {
		const engine = fakeEngine({
			aberto: true,
			opcoes: [
				{ tipo: 'produtos', label: 'Prod', icone: '🔎' },
				{ tipo: 'lojas', label: 'Lojas', icone: '🏪' }
			]
		});
		render(Omnibox, { props: { engine } });
		const buttons = screen.getAllByRole('option');
		await fireEvent.click(buttons[1].querySelector('button'));
		expect(engine.send).toHaveBeenCalledWith({ type: 'OMNIBOX_SELECIONAR', indice: 1 });
	});

	it('highlights active option visually', () => {
		const engine = fakeEngine({
			aberto: true,
			highlightIdx: 0,
			opcoes: [{ tipo: 'produtos', label: 'Prod', icone: '🔎' }]
		});
		render(Omnibox, { props: { engine } });
		const option = screen.getByRole('option');
		expect(option.getAttribute('aria-selected')).toBe('true');
	});

	it('sets aria-activedescendant when highlighted', () => {
		const engine = fakeEngine({
			aberto: true,
			highlightIdx: 0,
			opcoes: [{ tipo: 'produtos', label: 'Prod', icone: '🔎' }]
		});
		render(Omnibox, { props: { engine } });
		const input = screen.getByRole('combobox');
		expect(input.getAttribute('aria-activedescendant')).toBe('omnibox-opt-0');
	});

	it('announces option count via aria-live', () => {
		const engine = fakeEngine({
			aberto: true,
			opcoes: [
				{ tipo: 'produtos', label: 'P', icone: '🔎' },
				{ tipo: 'lojas', label: 'L', icone: '🏪' }
			]
		});
		render(Omnibox, { props: { engine } });
		const liveRegion = document.querySelector('[aria-live="polite"]');
		expect(liveRegion.textContent).toContain('2');
		expect(liveRegion.textContent).toContain('opções');
	});
});
