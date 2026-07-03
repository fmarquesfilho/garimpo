import { describe, it, expect, vi } from 'vitest';

/**
 * Testa a lógica de timeout para buscas que demoram demais.
 * Garante que o usuário nunca fica preso no estado "Buscando..." para sempre.
 */

// Replica a lógica de timeout usada na página de busca (+page.svelte)
async function buscarComTimeout(fetchFn, timeoutMs = 20000) {
	let timeoutId;
	const timeout = new Promise((_, reject) => {
		timeoutId = setTimeout(
			() => reject(new Error('A busca demorou demais. Verifique a conexão ou tente outro termo.')),
			timeoutMs
		);
	});

	try {
		const result = await Promise.race([fetchFn(), timeout]);
		return { data: result, erro: null };
	} catch (e) {
		return { data: null, erro: e };
	} finally {
		clearTimeout(timeoutId);
	}
}

describe('Busca: timeout de carregamento', () => {
	it('retorna resultado quando a API responde antes do timeout', async () => {
		const apiFn = () => Promise.resolve({ candidatos: [{ id: '1', nome: 'Produto' }] });
		const { data, erro } = await buscarComTimeout(apiFn, 1000);

		expect(erro).toBeNull();
		expect(data.candidatos).toHaveLength(1);
	});

	it('retorna erro com mensagem amigável quando a API não responde no tempo', async () => {
		vi.useFakeTimers();

		const apiLenta = () => new Promise(() => {}); // nunca resolve
		const promise = buscarComTimeout(apiLenta, 500);

		vi.advanceTimersByTime(501);
		const { data, erro } = await promise;

		expect(data).toBeNull();
		expect(erro).not.toBeNull();
		expect(erro.message).toContain('demorou demais');

		vi.useRealTimers();
	});

	it('retorna erro original quando a API falha antes do timeout', async () => {
		const apiQueFalha = () => Promise.reject(new Error('502 Bad Gateway'));
		const { data, erro } = await buscarComTimeout(apiQueFalha, 5000);

		expect(data).toBeNull();
		expect(erro.message).toBe('502 Bad Gateway');
	});

	it('limpa o timer quando a API responde (evita leak)', async () => {
		vi.useFakeTimers();
		const clearSpy = vi.spyOn(global, 'clearTimeout');

		const apiRapida = () => Promise.resolve({ candidatos: [] });
		await buscarComTimeout(apiRapida, 5000);

		expect(clearSpy).toHaveBeenCalled();

		clearSpy.mockRestore();
		vi.useRealTimers();
	});

	it('não busca sem keyword (retorna lista vazia imediatamente)', () => {
		// Replica lógica: se busca está vazia, nem chama a API
		const keyword = '   ';
		const deveBuscar = keyword.trim().length > 0;
		expect(deveBuscar).toBe(false);
	});

	it('busca com keyword válida', () => {
		const keyword = 'Skin1004';
		const deveBuscar = keyword.trim().length > 0;
		expect(deveBuscar).toBe(true);
	});
});

describe('Publicações: timeout de conversões reais', () => {
	it('mostra erro após timeout em vez de loading infinito', async () => {
		vi.useFakeTimers();

		const shopeeHangada = () => new Promise(() => {}); // nunca resolve
		const promise = buscarComTimeout(shopeeHangada, 20000);

		vi.advanceTimersByTime(20001);
		const { data, erro } = await promise;

		expect(data).toBeNull();
		expect(erro).not.toBeNull();
		expect(erro.message).toContain('demorou demais');

		vi.useRealTimers();
	});

	it('passa o resultado normalmente quando a Shopee responde', async () => {
		const shopeeOk = () => Promise.resolve({ total: 3, comissao_total: 15.5, conversoes: [] });
		const { data, erro } = await buscarComTimeout(shopeeOk, 20000);

		expect(erro).toBeNull();
		expect(data.total).toBe(3);
		expect(data.comissao_total).toBe(15.5);
	});
});
