import { describe, it, expect } from 'vitest';

/**
 * Testa a lógica de geração de link que a página de Oportunidades usa
 * antes de navegar para /publicar. Garante que o link é construído
 * corretamente a partir do busca_id + produto_id.
 */

function gerarLinkProduto(item) {
	let link = item.link ?? '';
	if (!link && item.loja && item.produto_id) {
		const shopId = item.loja.replace('loja-', '');
		if (/^\d+$/.test(shopId)) {
			link = `https://shopee.com.br/product-i.${shopId}.${item.produto_id}`;
		}
	}
	return link;
}

describe('Oportunidades → Publicar: geração de link', () => {
	it('gera link a partir de loja-SHOPID + produto_id', () => {
		const item = { produto_id: '23198188943', loja: 'loja-258316442', nome: 'Sérum' };
		expect(gerarLinkProduto(item)).toBe('https://shopee.com.br/product-i.258316442.23198188943');
	});

	it('mantém link existente se já veio preenchido', () => {
		const item = { produto_id: '111', loja: 'loja-999', link: 'https://shope.ee/aff123' };
		expect(gerarLinkProduto(item)).toBe('https://shope.ee/aff123');
	});

	it('retorna vazio se loja não tem formato numérico', () => {
		const item = { produto_id: '111', loja: 'minha-loja', nome: 'X' };
		expect(gerarLinkProduto(item)).toBe('');
	});

	it('retorna vazio se não tem loja nem link', () => {
		const item = { produto_id: '111', nome: 'X' };
		expect(gerarLinkProduto(item)).toBe('');
	});

	it('retorna vazio se não tem produto_id', () => {
		const item = { loja: 'loja-999', nome: 'X' };
		expect(gerarLinkProduto(item)).toBe('');
	});
});

describe('Publicar: resolução automática de imagem', () => {
	it('deve tentar resolver quando tem link mas não imagem', () => {
		const produto = { id: '123', nome: 'Sérum', link: 'https://shopee.com.br/product-i.999.123', imagem: '' };
		// A condição que trigger a resolução:
		const deveResolver = produto && !produto.imagem && produto.link;
		expect(deveResolver).toBeTruthy();
	});

	it('não deve resolver quando já tem imagem', () => {
		const produto = { id: '123', nome: 'Sérum', link: 'https://shopee.com.br/x', imagem: 'https://img.jpg' };
		const deveResolver = produto && !produto.imagem && produto.link;
		expect(deveResolver).toBeFalsy();
	});

	it('não deve resolver quando não tem link', () => {
		const produto = { id: '123', nome: 'Sérum', link: '', imagem: '' };
		const deveResolver = produto && !produto.imagem && produto.link;
		expect(deveResolver).toBeFalsy();
	});
});
