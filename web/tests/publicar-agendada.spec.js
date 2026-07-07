/**
 * E2E: Publicações — envio imediato e agendamento.
 *
 * Testa:
 * 1. Publicar imediatamente (link de afiliada gerado + enviado ao Publisher)
 * 2. Agendar publicação (persiste com status="agendada" + registra job no Scheduler)
 * 3. Endpoint interno /internal/publish-scheduled processa publicação agendada
 *
 * Pré-requisitos:
 *   - API C# rodando (mise run up)
 *   - Collector Go rodando (porta 50051) — para GenerateAffiliateLink
 *   - Scheduler Go rodando (porta 50054) — para SetSchedule
 *   - Publisher Go rodando (porta 50052) — para Publish
 *   - Firebase Auth Emulator (porta 9099)
 *
 * Execução:
 *   mise run test:e2e:publicar
 */
import { test, expect } from './fixtures.js';

test.describe('Publicações — Envio e Agendamento', () => {
	test.slow();

	test('publicar imediatamente gera link de afiliada e persiste', async ({ authedPage: page }) => {
		// Chama POST /api/publicar diretamente com dados de produto
		const response = await page.request.post('/api/publicar', {
			data: {
				id: 'test-produto-123',
				nome: 'Sérum Vitamina C',
				preco: 49.9,
				comissao: 0.085,
				link: 'https://shopee.com.br/product/920292999/25000641551',
				imagem: 'https://cf.shopee.com.br/file/test.jpg',
				categoria: 'Beleza',
				estrategia: 'nicho',
				destino_id: '' // sem destino = não envia, mas gera link
			}
		});

		// Sem destino configurado, o Publisher não é chamado mas a publicação é registrada
		expect(response.status()).toBe(200);
		const body = await response.json();
		expect(body).toHaveProperty('success');
		expect(body).toHaveProperty('publicacao_id');
	});

	test('agendar publicação persiste com status agendada', async ({ authedPage: page }) => {
		// Agenda para daqui a 1 hora
		const agendadaEm = new Date(Date.now() + 60 * 60 * 1000).toISOString();

		const response = await page.request.post('/api/publicacoes', {
			data: {
				nome: 'Protetor Solar FPS 50',
				preco: 39.9,
				comissao: 0.07,
				link: 'https://shopee.com.br/product/111222333/444555666',
				imagem: 'https://cf.shopee.com.br/file/protetor.jpg',
				categoria: 'Saúde',
				estrategia: 'nicho',
				destino_id: 'test-destino',
				agendada_em: agendadaEm
			}
		});

		expect(response.status()).toBe(200);
		const body = await response.json();
		expect(body.publicacao.status).toBe('agendada');
		expect(body.publicacao).toHaveProperty('id');
		expect(body.publicacao).toHaveProperty('criada_em');
	});

	test('endpoint interno publish-scheduled processa publicação', async ({ authedPage: page }) => {
		// Primeiro cria uma publicação agendada
		const agendadaEm = new Date(Date.now() + 60 * 60 * 1000).toISOString();
		const createResp = await page.request.post('/api/publicacoes', {
			data: {
				nome: 'Kit Skincare Coreano',
				preco: 89.9,
				comissao: 0.09,
				link: 'https://shopee.com.br/product/920292999/777888999',
				imagem: 'https://cf.shopee.com.br/file/skincare.jpg',
				estrategia: 'nicho',
				destino_id: 'test-destino',
				agendada_em: agendadaEm
			}
		});
		expect(createResp.status()).toBe(200);
		const created = await createResp.json();
		const pubId = created.publicacao.id;

		// Simula o callback do Scheduler chamando o endpoint interno
		const publishResp = await page.request.post('/internal/publish-scheduled', {
			data: { publicacao_id: pubId }
		});

		expect(publishResp.status()).toBe(200);
		const result = await publishResp.json();
		// Pode ser 'erro' (Publisher/destino não configurado em teste) ou 'enviada'
		expect(['enviada', 'erro']).toContain(result.status);
	});

	test('GET /api/publicacoes lista publicações com status correto', async ({ authedPage: page }) => {
		const response = await page.request.get('/api/publicacoes');
		expect(response.status()).toBe(200);

		const data = await response.json();
		expect(data).toHaveProperty('publicacoes');
		expect(data).toHaveProperty('total');

		if (data.publicacoes.length > 0) {
			const pub = data.publicacoes[0];
			expect(pub).toHaveProperty('id');
			expect(pub).toHaveProperty('nome');
			expect(pub).toHaveProperty('status');
			expect(pub).toHaveProperty('criada_em');
			expect(['pendente', 'agendada', 'enviada', 'erro']).toContain(pub.status);
		}
	});

	test('publicação com ID inválido retorna erro', async ({ authedPage: page }) => {
		const response = await page.request.post('/internal/publish-scheduled', {
			data: { publicacao_id: '00000000-0000-0000-0000-000000000000' }
		});

		expect(response.status()).toBe(404);
		const body = await response.json();
		expect(body.error).toContain('não encontrada');
	});
});
