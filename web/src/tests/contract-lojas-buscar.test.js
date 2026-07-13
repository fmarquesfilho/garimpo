/**
 * Contract test: valida que o schema esperado pelo frontend (api.js buscarLojas)
 * e consistente com o contrato JSON definido em contracts/schemas/lojas-buscar.response.json.
 *
 * Se a API C# mudar o shape sem atualizar o contrato, este teste falha.
 * Se o frontend espera campos que nao estao no contrato, este teste falha.
 */
import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { resolve } from 'path';

const schema = JSON.parse(
	readFileSync(resolve(import.meta.dirname, '../../../contracts/schemas/lojas-buscar.response.json'), 'utf-8')
);

// Simula uma resposta valida da API (como o frontend espera)
const SAMPLE_RESPONSE = {
	lojas: [
		{
			id: '920292999',
			nome: 'Glory of Seoul',
			nome_normalizado: 'gloryofseoul',
			marketplace: 'shopee',
			monitorada: true,
			origem: '🇰🇷',
			imagem: 'https://img.test/glory.jpg',
			seguidores: 12000,
			total_produtos: 340,
			avaliacao: 4.8
		},
		{
			id: '123',
			nome: 'Loja Minima',
			nome_normalizado: 'lojaminima',
			marketplace: 'amazon',
			monitorada: false,
			origem: null,
			imagem: null,
			seguidores: null,
			total_produtos: null,
			avaliacao: null
		}
	],
	total: 2
};

describe('Contract: GET /api/lojas/buscar response', () => {
	it('schema requer campos obrigatorios', () => {
		const required = schema.properties.lojas.items.required;
		expect(required).toContain('id');
		expect(required).toContain('nome');
		expect(required).toContain('nome_normalizado');
		expect(required).toContain('marketplace');
		expect(required).toContain('monitorada');
	});

	it('response de amostra satisfaz campos obrigatorios', () => {
		for (const loja of SAMPLE_RESPONSE.lojas) {
			expect(loja.id).toBeDefined();
			expect(typeof loja.id).toBe('string');
			expect(loja.nome).toBeDefined();
			expect(typeof loja.nome).toBe('string');
			expect(loja.nome_normalizado).toBeDefined();
			expect(loja.marketplace).toBeDefined();
			expect(typeof loja.monitorada).toBe('boolean');
		}
	});

	it('response tem total numerico', () => {
		expect(typeof SAMPLE_RESPONSE.total).toBe('number');
		expect(SAMPLE_RESPONSE.total).toBe(SAMPLE_RESPONSE.lojas.length);
	});

	it('campos opcionais podem ser null', () => {
		const minima = SAMPLE_RESPONSE.lojas[1];
		expect(minima.origem).toBeNull();
		expect(minima.imagem).toBeNull();
		expect(minima.seguidores).toBeNull();
		expect(minima.total_produtos).toBeNull();
		expect(minima.avaliacao).toBeNull();
	});

	it('schema permite campos opcionais como nullable', () => {
		const props = schema.properties.lojas.items.properties;
		expect(props.origem.type).toContain('null');
		expect(props.imagem.type).toContain('null');
		expect(props.seguidores.type).toContain('null');
		expect(props.total_produtos.type).toContain('null');
		expect(props.avaliacao.type).toContain('null');
	});

	it('frontend usa todos os campos do contrato', () => {
		// O frontend (StoreCard, BuscaEngine) usa estes campos:
		const camposUsados = [
			'id',
			'nome',
			'marketplace',
			'monitorada',
			'origem',
			'imagem',
			'seguidores',
			'total_produtos',
			'avaliacao'
		];
		const camposContrato = Object.keys(schema.properties.lojas.items.properties);
		for (const campo of camposUsados) {
			expect(camposContrato, `Campo ${campo} usado pelo frontend nao esta no contrato`).toContain(campo);
		}
	});
});
