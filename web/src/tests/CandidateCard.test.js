import { render, screen, cleanup } from '@testing-library/svelte';
import { describe, it, expect, afterEach } from 'vitest';
import CandidateCard from '$lib/components/CandidateCard.svelte';

afterEach(() => cleanup());

const candidatoBase = {
	id: 'P1',
	nome: 'Sérum Vitamina C SKIN1004',
	categoria: 'cosméticos',
	loja: 'SKIN1004 Official',
	preco: 89.9,
	comissao: 0.12,
	vendas: 150,
	avaliacao: 4.8,
	score: 0.75,
	link: 'https://shope.ee/aff123',
	imagem: 'https://img.shopee.com/thumb.jpg',
	componentes: { comissao: 0.3, vendas: 0.25, nota: 0.2 },
	suspeito: false
};

describe('CandidateCard — renderização básica', () => {
	it('mostra nome do produto', () => {
		render(CandidateCard, { props: { candidato: candidatoBase } });
		expect(screen.getByText('Sérum Vitamina C SKIN1004')).toBeInTheDocument();
	});

	it('mostra nome da loja', () => {
		render(CandidateCard, { props: { candidato: candidatoBase } });
		expect(screen.getByText('🏪 SKIN1004 Official')).toBeInTheDocument();
	});

	it('mostra categoria', () => {
		render(CandidateCard, { props: { candidato: candidatoBase } });
		expect(screen.getByText('cosméticos')).toBeInTheDocument();
	});

	it('mostra preço formatado', () => {
		render(CandidateCard, { props: { candidato: candidatoBase } });
		// BRL format: R$ 89,90
		expect(screen.getByText(/R\$\s*89,90/)).toBeInTheDocument();
	});

	it('mostra comissão como percentual', () => {
		render(CandidateCard, { props: { candidato: candidatoBase } });
		expect(screen.getByText('12%')).toBeInTheDocument();
	});

	it('mostra vendas', () => {
		render(CandidateCard, { props: { candidato: candidatoBase } });
		expect(screen.getByText(/150 vendas/)).toBeInTheDocument();
	});
});

describe('CandidateCard — badge de origem', () => {
	it('mostra badge 🇰🇷 Coreia quando origem=Coreia', () => {
		const c = { ...candidatoBase, origem: 'Coreia' };
		render(CandidateCard, { props: { candidato: c } });
		expect(screen.getByText(/🇰🇷/)).toBeInTheDocument();
		expect(screen.getByText(/Coreia/)).toBeInTheDocument();
	});

	it('mostra badge 🇯🇵 Japão quando origem=Japão', () => {
		const c = { ...candidatoBase, origem: 'Japão' };
		render(CandidateCard, { props: { candidato: c } });
		expect(screen.getByText(/🇯🇵/)).toBeInTheDocument();
	});

	it('mostra badge 🇨🇳 China quando origem=China', () => {
		const c = { ...candidatoBase, origem: 'China' };
		render(CandidateCard, { props: { candidato: c } });
		expect(screen.getByText(/🇨🇳/)).toBeInTheDocument();
	});

	it('não mostra badge de origem quando vazio', () => {
		const c = { ...candidatoBase, origem: '' };
		render(CandidateCard, { props: { candidato: c } });
		expect(screen.queryByText(/🇰🇷/)).not.toBeInTheDocument();
		expect(screen.queryByText(/🇯🇵/)).not.toBeInTheDocument();
	});
});

describe('CandidateCard — badge de desconto', () => {
	it('mostra badge 🔥 30% OFF quando desconto=0.30', () => {
		const c = { ...candidatoBase, desconto: 0.3 };
		render(CandidateCard, { props: { candidato: c } });
		expect(screen.getByText(/🔥 30% OFF/)).toBeInTheDocument();
	});

	it('não mostra badge de desconto quando desconto=0', () => {
		const c = { ...candidatoBase, desconto: 0 };
		render(CandidateCard, { props: { candidato: c } });
		expect(screen.queryByText(/% OFF/)).not.toBeInTheDocument();
	});
});

describe('CandidateCard — expiração de oferta', () => {
	it('mostra badge ⏳ com tempo restante quando oferta_expira no futuro', () => {
		const futuro = new Date(Date.now() + 5 * 24 * 3600000).toISOString(); // 5 dias
		const c = { ...candidatoBase, oferta_expira: futuro };
		render(CandidateCard, { props: { candidato: c } });
		// Deve mostrar ⏳ com algum número de dias (4d ou 5d dependendo da hora)
		expect(screen.getByText(/⏳ \d+d/)).toBeInTheDocument();
	});

	it('não mostra badge de expiração quando oferta_expira vazio', () => {
		const c = { ...candidatoBase, oferta_expira: '' };
		render(CandidateCard, { props: { candidato: c } });
		expect(screen.queryByText(/⏳/)).not.toBeInTheDocument();
	});
});

describe('CandidateCard — suspeito', () => {
	it('mostra badge ⚠ suspeito quando suspeito=true', () => {
		const c = { ...candidatoBase, suspeito: true };
		render(CandidateCard, { props: { candidato: c } });
		expect(screen.getByText(/⚠ suspeito/)).toBeInTheDocument();
	});

	it('não mostra badge suspeito quando suspeito=false', () => {
		render(CandidateCard, { props: { candidato: candidatoBase } });
		expect(screen.queryByText(/suspeito/)).not.toBeInTheDocument();
	});
});
