import { render, screen, cleanup } from '@testing-library/svelte';
import { describe, it, expect, afterEach } from 'vitest';
import HeroProduto from '$lib/components/HeroProduto.svelte';
import PublicarPreview from '$lib/components/PublicarPreview.svelte';

afterEach(() => cleanup());

// ── HeroProduto ───────────────────────────────────────────────────────────

describe('HeroProduto — renderização', () => {
	const produto = {
		nome: 'Sérum Vitamina C',
		preco: 89.9,
		categoria: 'skincare',
		imagem: 'https://img.test/thumb.jpg',
		link: 'https://shopee.com.br/product-test'
	};

	it('mostra imagem do produto', () => {
		render(HeroProduto, { props: { produto } });
		const img = screen.getByAltText('Sérum Vitamina C');
		expect(img).toBeInTheDocument();
		expect(img.src).toBe('https://img.test/thumb.jpg');
	});

	it('mostra placeholder quando sem imagem', () => {
		render(HeroProduto, { props: { produto: { ...produto, imagem: '' } } });
		expect(screen.getByText('📦')).toBeInTheDocument();
	});

	it('mostra input com nome editável', () => {
		render(HeroProduto, { props: { produto } });
		const input = screen.getByPlaceholderText('Nome do produto');
		expect(input.value).toBe('Sérum Vitamina C');
	});

	it('mostra link do produto truncado', () => {
		render(HeroProduto, { props: { produto } });
		expect(screen.getByText(/shopee\.com\.br/)).toBeInTheDocument();
	});

	it('não mostra link quando vazio', () => {
		render(HeroProduto, { props: { produto: { ...produto, link: '' } } });
		expect(screen.queryByRole('link')).not.toBeInTheDocument();
	});
});

// ── PublicarPreview ───────────────────────────────────────────────────────

describe('PublicarPreview — renderização', () => {
	it('mostra imagem quando disponível', () => {
		render(PublicarPreview, { props: { imagem: 'https://img.test/1.jpg', legenda: 'Teste', link: '' } });
		const img = screen.getByAltText('preview');
		expect(img).toBeInTheDocument();
	});

	it('não mostra imagem quando vazia', () => {
		render(PublicarPreview, { props: { imagem: '', legenda: 'Teste', link: '' } });
		expect(screen.queryByAltText('preview')).not.toBeInTheDocument();
	});

	it('renderiza legenda como HTML', () => {
		render(PublicarPreview, { props: { imagem: '', legenda: '✨ <b>Produto</b>', link: '' } });
		expect(screen.getByText('Produto')).toBeInTheDocument();
	});

	it('mostra botão Comprar quando tem link', () => {
		render(PublicarPreview, { props: { imagem: '', legenda: 'x', link: 'https://shopee.com.br/x' } });
		expect(screen.getByText('🛒 Comprar')).toBeInTheDocument();
	});

	it('não mostra botão Comprar sem link', () => {
		render(PublicarPreview, { props: { imagem: '', legenda: 'x', link: '' } });
		expect(screen.queryByText('🛒 Comprar')).not.toBeInTheDocument();
	});

	it('mostra label "Preview"', () => {
		render(PublicarPreview, { props: { imagem: '', legenda: 'x', link: '' } });
		expect(screen.getByText('Preview')).toBeInTheDocument();
	});
});

// ── Lógica de publicação (unit tests puros) ───────────────────────────────

describe('Publicar — lógica de legenda', () => {
	function gerarLegendaLocal(produto) {
		let t = '';
		if (produto.nome) t += `✨ <b>${produto.nome}</b>\n`;
		if (produto.categoria) t += `📂 <i>${produto.categoria}</i>\n`;
		if (produto.preco > 0) t += `💸 <b>R$ ${produto.preco.toFixed(2)}</b>`;
		return t.trimEnd();
	}

	it('gera legenda com nome, categoria e preço', () => {
		const leg = gerarLegendaLocal({ nome: 'Sérum', categoria: 'skincare', preco: 49.9 });
		expect(leg).toContain('Sérum');
		expect(leg).toContain('skincare');
		expect(leg).toContain('49.90');
	});

	it('omite categoria se vazia', () => {
		const leg = gerarLegendaLocal({ nome: 'X', categoria: '', preco: 10 });
		expect(leg).not.toContain('📂');
	});

	it('omite preço se zero', () => {
		const leg = gerarLegendaLocal({ nome: 'X', categoria: 'y', preco: 0 });
		expect(leg).not.toContain('💸');
	});
});

describe('Publicar — validação de envio', () => {
	function podeEnviar({ destinoId, publicando }) {
		return !publicando && !!destinoId;
	}

	it('pode enviar quando tem destino e não está publicando', () => {
		expect(podeEnviar({ destinoId: 'telegram-1', publicando: false })).toBe(true);
	});

	it('não pode enviar sem destino', () => {
		expect(podeEnviar({ destinoId: '', publicando: false })).toBe(false);
	});

	it('não pode enviar enquanto publicando', () => {
		expect(podeEnviar({ destinoId: 'x', publicando: true })).toBe(false);
	});
});
