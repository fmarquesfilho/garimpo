/**
 * Tests for StoreCard.svelte — card de resultado de busca por loja.
 */
import { render, screen, cleanup, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, afterEach, vi } from 'vitest';
import StoreCard from '$lib/components/StoreCard.svelte';

afterEach(() => cleanup());

function fakeEngine(resolucaoOverrides = {}) {
	return {
		ctx: {
			resolucaoLoja: { status: 'idle', ...resolucaoOverrides }
		},
		send: vi.fn()
	};
}

const lojaCompleta = {
	id: '100',
	nome: 'Glory of Seoul',
	marketplace: 'shopee',
	monitorada: false,
	origem: '🇰🇷',
	imagem: 'https://img.test/glory.jpg',
	seguidores: 12000,
	total_produtos: 340,
	avaliacao: 4.8
};

const lojaMinima = {
	id: '200',
	nome: 'Basic Store',
	marketplace: 'amazon',
	monitorada: false,
	origem: null,
	imagem: null,
	seguidores: null,
	total_produtos: null,
	avaliacao: null
};

describe('StoreCard — rendering', () => {
	it('renders loja name', () => {
		render(StoreCard, { props: { loja: lojaCompleta, engine: fakeEngine() } });
		expect(screen.getByText('Glory of Seoul')).toBeInTheDocument();
	});

	it('renders marketplace', () => {
		render(StoreCard, { props: { loja: lojaCompleta, engine: fakeEngine() } });
		expect(screen.getByText(/shopee/)).toBeInTheDocument();
	});

	it('renders bandeira de origem when available', () => {
		render(StoreCard, { props: { loja: lojaCompleta, engine: fakeEngine() } });
		expect(screen.getByTitle('Origem')).toHaveTextContent('🇰🇷');
	});

	it('does not render bandeira when origem is null', () => {
		render(StoreCard, { props: { loja: lojaMinima, engine: fakeEngine() } });
		expect(screen.queryByTitle('Origem')).not.toBeInTheDocument();
	});

	it('renders avatar when imagem available', () => {
		render(StoreCard, { props: { loja: lojaCompleta, engine: fakeEngine() } });
		const img = screen.getByRole('img');
		expect(img).toHaveAttribute('src', 'https://img.test/glory.jpg');
	});

	it('renders fallback icon when no imagem', () => {
		render(StoreCard, { props: { loja: lojaMinima, engine: fakeEngine() } });
		expect(screen.queryByRole('img')).not.toBeInTheDocument();
		// Fallback icon (marketplace icon or default)
		expect(screen.getByText('🟡')).toBeInTheDocument(); // amazon icon
	});

	it('renders seguidores when available', () => {
		render(StoreCard, { props: { loja: lojaCompleta, engine: fakeEngine() } });
		expect(screen.getByText(/12.*000.*seguidores/i)).toBeInTheDocument();
	});

	it('renders total_produtos when available', () => {
		render(StoreCard, { props: { loja: lojaCompleta, engine: fakeEngine() } });
		expect(screen.getByText(/340 produtos/)).toBeInTheDocument();
	});

	it('renders avaliacao when available', () => {
		render(StoreCard, { props: { loja: lojaCompleta, engine: fakeEngine() } });
		expect(screen.getByText(/4\.8/)).toBeInTheDocument();
	});

	it('does not render optional fields when null', () => {
		render(StoreCard, { props: { loja: lojaMinima, engine: fakeEngine() } });
		expect(screen.queryByText(/seguidores/)).not.toBeInTheDocument();
		expect(screen.queryByText(/produtos/)).not.toBeInTheDocument();
		expect(screen.queryByText(/⭐/)).not.toBeInTheDocument();
	});
});

describe('StoreCard — monitoramento', () => {
	it('shows Monitorar button when not monitorada', () => {
		render(StoreCard, { props: { loja: lojaCompleta, engine: fakeEngine() } });
		expect(screen.getByRole('button', { name: /monitorar/i })).toBeInTheDocument();
	});

	it('shows check indicator when monitorada', () => {
		const loja = { ...lojaCompleta, monitorada: true };
		render(StoreCard, { props: { loja, engine: fakeEngine() } });
		expect(screen.queryByRole('button', { name: /monitorar/i })).not.toBeInTheDocument();
		expect(screen.getByTitle('Monitorada')).toBeInTheDocument();
	});

	it('click Monitorar dispatches MONITORAR_LOJA', async () => {
		const engine = fakeEngine();
		render(StoreCard, { props: { loja: lojaCompleta, engine } });
		await fireEvent.click(screen.getByRole('button', { name: /monitorar/i }));
		expect(engine.send).toHaveBeenCalledWith({ type: 'MONITORAR_LOJA', loja: lojaCompleta });
	});

	it('button disabled when resolvendo', () => {
		const engine = fakeEngine({ status: 'resolvendo' });
		render(StoreCard, { props: { loja: lojaCompleta, engine } });
		expect(screen.getByRole('button', { name: /monitorar/i })).toBeDisabled();
	});

	it('shows error message when resolucao fails', () => {
		const engine = fakeEngine({ status: 'erro', erro: 'Falha de rede' });
		render(StoreCard, { props: { loja: lojaCompleta, engine } });
		expect(screen.getByText('Falha de rede')).toBeInTheDocument();
	});
});

describe('StoreCard — accessibility', () => {
	it('button has aria-label with loja name', () => {
		render(StoreCard, { props: { loja: lojaCompleta, engine: fakeEngine() } });
		const btn = screen.getByRole('button', { name: /monitorar loja glory of seoul/i });
		expect(btn).toBeInTheDocument();
	});
});
